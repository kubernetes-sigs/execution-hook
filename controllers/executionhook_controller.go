/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/execution-hook/util"
	"sigs.k8s.io/execution-hook/util/patch"

	appsv1alpha1 "sigs.k8s.io/execution-hook/api/v1alpha1"
)

// ExecutionHookReconciler reconciles a ExecutionHook object
type ExecutionHookReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	controller controller.Controller
	recorder   record.EventRecorder
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// SetupWithManager sets up executionhook controller with a controller manager
func (r *ExecutionHookReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&appsv1alpha1.ExecutionHook{}).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	r.controller = c
	r.recorder = mgr.GetEventRecorderFor("execution-hook-controller")

	return nil
}

// +kubebuilder:rbac:groups=apps.k8s.io,resources=executionhook,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.k8s.io,resources=executionhook/status,verbs=get;update;patch

func (r *ExecutionHookReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	log := r.Log.WithValues("executionhook", req.NamespacedName)

	// Fetch the executionhook instance
	hook := &appsv1alpha1.ExecutionHook{}
	if err := r.Client.Get(ctx, req.NamespacedName, hook); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	patchHelper, err := patch.NewHelper(hook, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := patchHelper.Patch(ctx, hook); err != nil {
			if reterr == nil {
				reterr = err
			}
		}
	}()

	log.Info("Reconciling hook")
	// Reconcile ExecutionHook deletion
	if !hook.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileHookDeletion(ctx, hook)
	}

	return r.reconcile(ctx, hook)
}

func (r *ExecutionHookReconciler) reconcileHookDeletion(ctx context.Context, hook *appsv1alpha1.ExecutionHook) (ctrl.Result, error) {
	log := r.Log.WithValues("executionhook", fmt.Sprintf("%s/%s", hook.Name, hook.Namespace))

	log.Info("Reconciling executionhook delete")
	// TODO (ashish-amarnath): if any pending hook actions, return with requeue after 30s
	// else remove ourselves from the finalizers and let the object be reclaimed
	hook.Finalizers = util.Filter(hook.Finalizers, appsv1alpha1.ExecutionHookFinalizer)
	return ctrl.Result{}, nil
}

func (r *ExecutionHookReconciler) reconcile(ctx context.Context, hook *appsv1alpha1.ExecutionHook) (ctrl.Result, error) {
	hookName := fmt.Sprintf("%s/%s", hook.Name, hook.Namespace)
	actionName := fmt.Sprintf("%s/%s", hook.Namespace, hook.Spec.ActionName)
	log := r.Log.WithValues("executionhook", hookName, "action", actionName)

	// add ourselves to the finalizer of the ExecutionHook object
	if !util.Contains(hook.Finalizers, appsv1alpha1.ExecutionHookFinalizer) {
		hook.Finalizers = append(hook.Finalizers, appsv1alpha1.ExecutionHookFinalizer)
	}

	log.Info("Looking up hook-action")
	//(TODO ashish-amarnath): lookup hook action associated with this hook

	log.Info("Selecting PodContainerNames to run hook-action")
	podContainerNamesList, err := r.selectPodContainers(hook.Namespace, &hook.Spec.PodSelection)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err,
			"failed to select podContainerNames to run hook action %s while reconciling execution-hook %s", actionName, hookName)
	}
	if len(podContainerNamesList) == 0 {
		log.Info("Nothing to do, PodSelection returned 0 PodContainerNames for execution hook",
			"podSelection", fmt.Sprintf("[%v]", hook.Spec.PodSelection))
		return ctrl.Result{}, nil
	}

	log.Info("Running HookAction", "targetContainerCount", len(podContainerNamesList))

	for _, pc := range podContainerNamesList {
		r.runHookAction(pc, hook, hook.Spec.ActionName)
	}
	return ctrl.Result{}, nil
}

func (r *ExecutionHookReconciler) selectPodContainers(ns string, ps *appsv1alpha1.PodSelection) ([]appsv1alpha1.PodContainerNames, error) {
	if ps == nil {
		return nil, errors.Errorf("Cannot use nil podSelection to select podContainers")
	}

	if ps.PodContainerNamesList != nil && len(ps.PodContainerNamesList) > 0 {
		return ps.PodContainerNamesList, nil
	}

	if ps.PodContainerSelector.PodSelector == nil {
		return []appsv1alpha1.PodContainerNames{}, nil
	}

	labelSelector, err := metav1.LabelSelectorAsSelector(ps.PodContainerSelector.PodSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert labelSelector to type slector")
	}
	r.Log.Info("Querying pods matching", "namespace", ns, "label-selector", labelSelector.String())
	listOpts := &client.ListOptions{
		LabelSelector: labelSelector,
	}

	res := []appsv1alpha1.PodContainerNames{}
	for {
		podList := corev1.PodList{}
		err := r.Client.List(context.TODO(), &podList, client.InNamespace(ns), listOpts)
		if err != nil {
			errMsg := fmt.Sprintf("failed to list pods in namespace %s matching labelselector %q", ns, labelSelector)
			return nil, errors.Wrapf(err, errMsg)
		}

		for _, p := range podList.Items {
			containers := []string{}
			for _, c := range p.Spec.Containers {
				containers = append(containers, c.Name)
			}
			pcn := appsv1alpha1.PodContainerNames{
				PodName:        p.Name,
				ContainerNames: containers,
			}
			res = append(res, pcn)
		}
		if podList.Continue == "" {
			break
		}
	}

	return res, nil
}

func (r *ExecutionHookReconciler) runHookAction(pc appsv1alpha1.PodContainerNames, hook *appsv1alpha1.ExecutionHook, hookAction string) {
	hookName := fmt.Sprintf("%s/%s", hook.Name, hook.Namespace)
	actionName := fmt.Sprintf("%s/%s", hook.Namespace, hook.Spec.ActionName)
	log := r.Log.WithValues("executionhook", hookName, "action", actionName)

	hookStatuses := []appsv1alpha1.ContainerExecutionHookStatus{}
	for _, c := range pc.ContainerNames {
		log.Info("Running hookaction on", "pod", pc.PodName, "container", c)
		// TODO (ashish-amarnath): run hook action using exec API
		result := false
		st := metav1.Now()
		cs := appsv1alpha1.ContainerExecutionHookStatus{
			PodName:       pc.PodName,
			ContainerName: c,
			Timestamp:     &st,
			Succeed:       &result,
			Error:         nil,
		}

		hookStatuses = append(hookStatuses, cs)
	}
	hook.Status.HookStatuses = hookStatuses
}

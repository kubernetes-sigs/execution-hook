/*

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
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "sigs.k8s.io/execution-hook/api/v1alpha1"
)

// ExecutionHookReconciler reconciles a ExecutionHook object
type ExecutionHookReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.k8s.io,resources=executionhook,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.k8s.io,resources=executionhook/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.k8s.io,resources=hookaction,verbs=get;list;watch

func (r *ExecutionHookReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("executionhook", req.NamespacedName)

	var hook appsv1alpha1.ExecutionHook
	if err := r.Client.Get(ctx, req.NamespacedName, &hook); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	log.Info("retrieved hook", "hook", hook)
	podContainerNamesList := hook.Spec.PodSelection.PodContainerNamesList
	if len(podContainerNamesList) == 0 {
		podContainerNamesList = r.podContainerNamesListFromSelector(hook.Spec.PodSelection.PodContainerSelector)
	}

	if len(podContainerNamesList) == 0 {
		return ctrl.Result{}, fmt.Errorf("PodSelection must define a list of Pods or a Pod selector and a list of container names [%+v]", hook.Spec.PodSelection)
	}

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ExecutionHookReconciler) podContainerNamesListFromSelector(podContainerSelector *appsv1alpha1.PodContainerSelector) []appsv1alpha1.PodContainerNames {
	if podContainerSelector == nil || podContainerSelector.PodSelector == nil {
		return []appsv1alpha1.PodContainerNames{}
	}

	// TODO: query pods using selector and return list of PodContainerNames for matching Pods

	return []appsv1alpha1.PodContainerNames{}
}

func (r *ExecutionHookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.ExecutionHook{}).
		Complete(r)
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

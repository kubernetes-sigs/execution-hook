/*
Copyright 2019 The Kubernetes Authors.

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

package controller

import (
	"fmt"
	"time"

	crdv1 "github.com/kubernetes-csi/execution-hook/pkg/apis/executionhook/v1alpha1"
	clientset "github.com/kubernetes-csi/execution-hook/pkg/client/clientset/versioned"
	hookinformers "github.com/kubernetes-csi/execution-hook/pkg/client/informers/externalversions/executionhook/v1alpha1"
	hooklisters "github.com/kubernetes-csi/execution-hook/pkg/client/listers/executionhook/v1alpha1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/goroutinemap"
)

type hookController struct {
	clientset         clientset.Interface
	client            kubernetes.Interface
	eventRecorder     record.EventRecorder
	hookQueue         workqueue.RateLimitingInterface
	hookTemplateQueue workqueue.RateLimitingInterface

	hookLister               hooklisters.ExecutionHookLister
	hookListerSynced         cache.InformerSynced
	hookTemplateLister       hooklisters.ExecutionHookTemplateLister
	hookTemplateListerSynced cache.InformerSynced

	hookStore         cache.Store
	hookTemplateStore cache.Store

	//handler Handler
	// Map of scheduled/running operations.
	runningOperations goroutinemap.GoRoutineMap

	resyncPeriod time.Duration
}

// NewHookController returns a new *hookController
func NewHookController(
	clientset clientset.Interface,
	client kubernetes.Interface,
	hookInformer hookinformers.ExecutionHookInformer,
	hookTemplateInformer hookinformers.ExecutionHookTemplateInformer,
	timeout time.Duration,
	resyncPeriod time.Duration,
) *hookController {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	var eventRecorder record.EventRecorder
	eventRecorder = broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("execution-hook")})

	ctrl := &hookController{
		clientset:         clientset,
		client:            client,
		eventRecorder:     eventRecorder,
		runningOperations: goroutinemap.NewGoRoutineMap(true),
		resyncPeriod:      resyncPeriod,
		hookStore:         cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc),
		hookTemplateStore: cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc),
		hookQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "hook"),
		hookTemplateQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "hook-template"),
	}

	hookInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueHookWork(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueHookWork(newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueHookWork(obj) },
		},
		ctrl.resyncPeriod,
	)
	ctrl.hookLister = hookInformer.Lister()
	ctrl.hookListerSynced = hookInformer.Informer().HasSynced

	hookTemplateInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueHookTemplateWork(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueHookTemplateWork(newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueHookTemplateWork(obj) },
		},
		ctrl.resyncPeriod,
	)
	ctrl.hookTemplateLister = hookTemplateInformer.Lister()
	ctrl.hookTemplateListerSynced = hookTemplateInformer.Informer().HasSynced

	return ctrl
}

func (ctrl *hookController) Run(workers int, stopCh <-chan struct{}) {
	defer ctrl.hookQueue.ShutDown()
	defer ctrl.hookTemplateQueue.ShutDown()

	klog.Infof("Starting ExecutionHook Controller")
	defer klog.Infof("Shutting ExecutionHook Controller")

	if !cache.WaitForCacheSync(stopCh, ctrl.hookListerSynced, ctrl.hookTemplateListerSynced) {
		klog.Errorf("Cannot sync caches")
		return
	}

	ctrl.initializeCaches(ctrl.hookLister, ctrl.hookTemplateLister)

	for i := 0; i < workers; i++ {
		go wait.Until(ctrl.hookWorker, 0, stopCh)
		go wait.Until(ctrl.hookTemplateWorker, 0, stopCh)
	}

	<-stopCh
}

// enqueueHookWork adds hook to given work queue.
func (ctrl *hookController) enqueueHookWork(obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	if hook, ok := obj.(*crdv1.ExecutionHook); ok {
		objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(hook)
		if err != nil {
			klog.Errorf("failed to get key from object: %v, %v", err, hook)
			return
		}
		klog.V(5).Infof("enqueued %q for sync", objName)
		ctrl.hookQueue.Add(objName)
	}
}

// enqueuehookTemplateWork adds hook template to given work queue.
func (ctrl *hookController) enqueueHookTemplateWork(obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	if hookTemplate, ok := obj.(*crdv1.ExecutionHookTemplate); ok {
		objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(hookTemplate)
		if err != nil {
			klog.Errorf("failed to get key from object: %v, %v", err, hookTemplate)
			return
		}
		klog.V(5).Infof("enqueued %q for sync", objName)
		ctrl.hookTemplateQueue.Add(objName)
	}
}

// hookWorker processes items from hookQueue. It must run only once,
// syncHook is not assured to be reentrant.
func (ctrl *hookController) hookWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.hookQueue.Get()
		if quit {
			return true
		}
		defer ctrl.hookQueue.Done(keyObj)
		key := keyObj.(string)
		klog.V(5).Infof("hookWorker[%s]", key)

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		klog.V(5).Infof("hookWorker: hook namespace [%s] name [%s]", namespace, name)
		if err != nil {
			klog.Errorf("error getting namespace & name of hook %q to get hook from informer: %v", key, err)
			return false
		}
		hook, err := ctrl.hookLister.ExecutionHooks(namespace).Get(name)
		if err == nil {
			// The hook still exists in informer cache, the event must have
			// been add/update/sync
			ctrl.updateHook(hook)
			return false
		}
		if err != nil && !errors.IsNotFound(err) {
			klog.V(2).Infof("error getting hook %q from informer: %v", key, err)
			return false
		}
		// The hook is not in informer cache, the event must have been "delete"
		hookObj, found, err := ctrl.hookStore.GetByKey(key)
		if err != nil {
			klog.V(2).Infof("error getting hook %q from cache: %v", key, err)
			return false
		}
		if !found {
			// The controller has already processed the delete event and
			// deleted the hook from its cache
			klog.V(2).Infof("deletion of hook %q was already processed", key)
			return false
		}
		hook, ok := hookObj.(*crdv1.ExecutionHook)
		if !ok {
			klog.Errorf("expected vs, got %+v", hookObj)
			return false
		}
		klog.V(5).Infof("deletion of hook %q will be processed", key)
		ctrl.deleteHook(hook)
		return false
	}

	for {
		if quit := workFunc(); quit {
			klog.Infof("hook worker queue shutting down")
			return
		}
	}
}

// hookTemplateWorker processes items from hookTemplateQueue. It must run only once,
// syncHookTemplate is not assured to be reentrant.
func (ctrl *hookController) hookTemplateWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.hookTemplateQueue.Get()
		if quit {
			return true
		}
		defer ctrl.hookTemplateQueue.Done(keyObj)
		key := keyObj.(string)
		klog.V(5).Infof("hookTemplateWorker[%s]", key)

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			klog.V(4).Infof("error getting name of hookTemplate %q to get hookTemplate from informer: %v", key, err)
			return false
		}
		hookTemplate, err := ctrl.hookTemplateLister.ExecutionHookTemplates(namespace).Get(name) //ctrl.hookTemplateLister.Get(name)
		// The hookTemplate still exists in informer cache, the event must have
		// been add/update/sync
		if err == nil {
			ctrl.updateHookTemplate(hookTemplate)
			return false
		}
		if !errors.IsNotFound(err) {
			klog.V(2).Infof("error getting hookTemplate %q from informer: %v", key, err)
			return false
		}

		// The hookTemplate is not in informer cache, the event must have been
		// "delete"
		hookTemplateObj, found, err := ctrl.hookTemplateStore.GetByKey(key)
		if err != nil {
			klog.V(2).Infof("error getting hookTemplate %q from cache: %v", key, err)
			return false
		}
		if !found {
			// The controller has already processed the delete event and
			// deleted the hookTemplate from its cache
			klog.V(2).Infof("deletion of hookTemplate %q was already processed", key)
			return false
		}
		hookTemplate, ok := hookTemplateObj.(*crdv1.ExecutionHookTemplate)
		if !ok {
			klog.Errorf("expected hookTemplate, got %+v", hookTemplate)
			return false
		}
		ctrl.deleteHookTemplate(hookTemplate)
		return false
	}

	for {
		if quit := workFunc(); quit {
			klog.Infof("hookTemplate worker queue shutting down")
			return
		}
	}
}

// updateHook runs in worker thread and handles "hook added",
// "hook updated" and "periodic sync" events.
func (ctrl *hookController) updateHook(hook *crdv1.ExecutionHook) {
	// Store the new hook version in the cache and do not process it if this is
	// an old version.
	klog.V(5).Infof("updateHook %q", hookKey(hook))
	newHook, err := ctrl.storeHookUpdate(hook)
	if err != nil {
		klog.Errorf("%v", err)
	}
	if !newHook {
		return
	}
	err = ctrl.syncHook(hook)
	if err != nil {
		if errors.IsConflict(err) {
			// Version conflict error happens quite often and the controller
			// recovers from it easily.
			klog.V(3).Infof("could not sync hook %q: %+v", hookKey(hook), err)
		} else {
			klog.Errorf("could not sync hook %q: %+v", hookKey(hook), err)
		}
	}
}

// updateHookTemplate runs in worker thread and handles "hookTemplate added",
// "hookTemplate updated" and "periodic sync" events.
func (ctrl *hookController) updateHookTemplate(hookTemplate *crdv1.ExecutionHookTemplate) {
	// Store the new hookTemplate version in the cache and do not process it if this is
	// an old version.
	new, err := ctrl.storeHookTemplateUpdate(hookTemplate)
	if err != nil {
		klog.Errorf("%v", err)
	}
	if !new {
		return
	}
	klog.V(5).Infof("msg1: updateHookTemplate: %+v", hookTemplate)
	err = ctrl.syncHookTemplate(hookTemplate)
	klog.V(5).Infof("msg2: updateHookTemplate: %+v", hookTemplate)
	if err != nil {
		if errors.IsConflict(err) {
			// Version conflict error happens quite often and the controller
			// recovers from it easily.
			klog.V(3).Infof("could not sync hookTemplate %q: %+v", hookTemplate.Name, err)
		} else {
			klog.Errorf("could not sync hookTemplate %q: %+v", hookTemplate.Name, err)
		}
	}
}

// deleteHook runs in worker thread and handles "hook deleted" event.
func (ctrl *hookController) deleteHook(hook *crdv1.ExecutionHook) {
	// Trigger PostAction
	err := ctrl.syncPostActionHook(hook)
	if err != nil {
		// TODO: log an event
		// PostAction failed, log an event but don't delete the hook
		// so that user can check the event and fix the problem manually
		klog.Errorf("failed to run PostAction hook: %v, %+v", err, hook)
		return
	}

	_ = ctrl.hookStore.Delete(hook)
	klog.V(4).Infof("hook %q deleted", hookKey(hook))
}

// deleteHookTemplate runs in worker thread and handles "hookTemplate deleted" event.
func (ctrl *hookController) deleteHookTemplate(hookTemplate *crdv1.ExecutionHookTemplate) {
	_ = ctrl.hookTemplateStore.Delete(hookTemplate)
	klog.V(4).Infof("hookTemplate %q deleted", hookTemplate.Name)
}

// initializeCaches fills all controller caches with initial data from etcd in
// order to have the caches already filled when first addHook/addHookTemplate to
// perform initial synchronization of the controller.
func (ctrl *hookController) initializeCaches(hookLister hooklisters.ExecutionHookLister, hookTemplateLister hooklisters.ExecutionHookTemplateLister) {
	hookList, err := hookLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("HookController can't initialize caches: %v", err)
		return
	}
	for _, hook := range hookList {
		hookClone := hook.DeepCopy()
		if _, err = ctrl.storeHookUpdate(hookClone); err != nil {
			klog.Errorf("error updating hook cache: %v", err)
		}
	}

	hookTemplateList, err := hookTemplateLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("HookController can't initialize caches: %v", err)
		return
	}
	for _, hookTemplate := range hookTemplateList {
		hookTemplateClone := hookTemplate.DeepCopy()
		if _, err = ctrl.storeHookTemplateUpdate(hookTemplateClone); err != nil {
			klog.Errorf("error updating hookTemplate cache: %v", err)
		}
	}

	klog.V(4).Infof("controller initialized")
}

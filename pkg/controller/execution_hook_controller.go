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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const controllerUpdateFailMsg = "hook controller failed to update"

// syncHookTemplate deals with one key off the queue.  It returns false when it's time to quit.
func (ctrl *hookController) syncHookTemplate(hookTemplate *crdv1.ExecutionHookTemplate) error {
	klog.V(5).Infof("synchronizing ExecutionHookTemplate[%s]", hookTemplate.Name)

	return nil
}

// syncHook is the main controller method to decide what to do with a hook.
// It's invoked by appropriate cache.Controller callbacks when a hook is
// created, updated or periodically synced. We do not differentiate between
// these events.
func (ctrl *hookController) syncHook(hook *crdv1.ExecutionHook) error {
	klog.V(5).Infof("synchonizing ExecutionHook[%s]", hook.Name) //hookKey(hook), getHookStatusForLogging(hook))

	return ctrl.syncPreActionHook(hook)
}

// syncPreActionHook is the main controller method to decide what to do with a hook whose PreActionSucceed summary status is not set to true.
func (ctrl *hookController) syncPreActionHook(hook *crdv1.ExecutionHook) error {
	// TODO: After Hook controller knows what containers and pods are selected,
	// it should create ContainerExecutionHookStatuses; otherwise we have to
	// save the container/pod list in cache
	// If PreActionSucceed summary status in ExecutionHookStatus is not set (nil),
	// loop around and wait until it is set

	statusFalse := false
	statusTrue := true
	klog.V(5).Infof("Entering syncPreActionHook with hook [%+v]", hook)

	// TODO: Make sure hook.Spec.PodContainerNamesList is populated before the next step
	if len(hook.Spec.PodContainerNamesList) == 0 {
		return fmt.Errorf("hook.Spec.PodContainerNamesList cannot be empty [%+v]", hook.Spec.PodContainerNamesList)
	}

	// hook.Status is of type crdv1.ExecutionHookStatus
	if hook.Status.PreActionSucceed == nil {
		// Initialize ContainerExecutionHookStatuses if it is empty based on
		// hook.Spec.PodContainerNamesList
		//var hookClone *crdv1.ExecutionHook
		//var hookUpdate *crdv1.ExecutionHook
		var err error
		klog.V(5).Infof("msg: syncPreActionHook with hook [%+v]", hook)
		if len(hook.Status.ContainerExecutionHookStatuses) == 0 {
			hookClone := hook.DeepCopy()
			klog.V(5).Infof("msg: syncPreActionHook with hookClone [%+v]", hookClone)
			// TODO: Move the code to build initial ContainerExecutionHookStatuses
			// outside of syncPreActionHook so we don't need to retrieve the hook
			// from hookLister again
			// podContainers is of type crdv1.PodContainerNames
			for _, podContainers := range hookClone.Spec.PodContainerNamesList {
				for _, containerName := range podContainers.ContainerNames {
					containerStatus := crdv1.ContainerExecutionHookStatus{}
					klog.V(5).Infof("msg: syncPreActionHook with PodContainerNamesList [%+v]", hook.Spec.PodContainerNamesList)
					containerStatus.PodName = podContainers.PodName
					containerStatus.ContainerName = containerName
					//containerStatus.PreActionStatus = nil
					klog.V(5).Infof("syncPreActionHook: add crdv1.ContainerExecutionHookStatus [%+v]", containerStatus)
					hookClone.Status.ContainerExecutionHookStatuses = append(hookClone.Status.ContainerExecutionHookStatuses, containerStatus)
					klog.V(5).Infof("msg: syncPreActionHook with containerStatuses [%+v]", hookClone.Status.ContainerExecutionHookStatuses)
				}
			}
			_, err = ctrl.updateHookStatus(hookClone)
			if err != nil {
				return fmt.Errorf("failed to update hook status [%+v]", hookClone)
			}
		}

		//if hookUpdate0 == nil {
		//        return fmt.Errorf("Here! failed to get updated hook status [%+v]", hookClone)
		//}

		klog.V(5).Infof("msg: syncPreActionHook before hookUpdate0.DeepCopy with hook [%+v]", hook)
		// syncPreAction will be entered multiple times. When it enters for the 2nd time,
		// hook.Status.ContainerExecutionHookStatuses is no longer 0
		hookUpdate0, err := ctrl.hookLister.ExecutionHooks(hook.Namespace).Get(hook.Name)
		//hookNew, err := ctrl.clientset.ExecutionhookV1alpha1().ExecutionHooks(hookUpdate.Namespace).Get(hookUpdate.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to retrieve hook %s from the informer: %q", hook.Name, err)
			return fmt.Errorf("failed to retrieve hook %s from the informer: %q", hook.Name, err)
		}
		hookUpdate := hookUpdate0.DeepCopy()
		statuses := []crdv1.ContainerExecutionHookStatus{}
		klog.V(5).Infof("msg: syncPreActionHook after hookUpdate0.DeepCopy with hook [%+v]", hook)
		// Loop through ContainerExecutionHookStatuses
		for i, _ := range hookUpdate0.Status.ContainerExecutionHookStatuses {
			// Hook PreAction has not started running if timestamp is nil
			// so run it now
			containerHookStatus := hookUpdate.Status.ContainerExecutionHookStatuses[i]
			if containerHookStatus.PreActionStatus == nil {
				containerHookStatus.PreActionStatus = &crdv1.ExecutionHookActionStatus{}
			}

			if containerHookStatus.PreActionStatus.ActionTimestamp == nil {
				// Set timestamp now
				var timeNow = metav1.Time{
					Time: time.Now(),
				}
				containerHookStatus.PreActionStatus.ActionTimestamp = &timeNow
				klog.V(5).Infof("msg: syncPreActionHook containerHookStatus actionTimestamp[%+v]", containerHookStatus.PreActionStatus.ActionTimestamp)
				// Run execution hook action
				// ExecPodContainerCommand will wait until it is done
				// TODO: add retry and timeout logic
				err := ExecPodContainerCommand(hook.Namespace, containerHookStatus.PodName, containerHookStatus.ContainerName, hook.Name, hook.Spec.PreAction.Action.Exec.Command, hook.Spec.PreAction.ActionTimeoutSeconds)
				// TODO: Remove this
				//var err error = nil

				// wait for it to come back
				// if success, set ActionSucceed = true
				// else, set ActionSucceed = false and
				// set hook.Status.PreActionSucceed = false
				// TODO: log an event
				// and bail out
				if err != nil { // failed on one command, no need to proceed, bail out
					containerHookStatus.PreActionStatus.ActionSucceed = &statusFalse
					hookUpdate.Status.PreActionSucceed = &statusFalse
					ctrl.updateHookStatus(hookUpdate)
					return fmt.Errorf("Failed to run PreAction %s in container %s in pod %s/%s", hookUpdate.Spec.PreAction.Action.Exec.Command, containerHookStatus.ContainerName, hookUpdate.Namespace, containerHookStatus.PodName)
				}
				// success
				successTrue := true
				containerHookStatus.PreActionStatus.ActionSucceed = &successTrue
				statuses = append(statuses, containerHookStatus)
				klog.V(5).Infof("msg: syncPreActionHook containerHookStatus actionTimestamp[%+v] actionStatus [%+v]", containerHookStatus.PreActionStatus.ActionTimestamp, containerHookStatus.PreActionStatus)
				klog.V(5).Infof("msg: syncPreActionHook containerHookStatus actionTimestamp[%+v] actionSucceed [%t]", containerHookStatus.PreActionStatus.ActionTimestamp, *(containerHookStatus.PreActionStatus.ActionSucceed))

				//klog.V(5).Infof("msg: syncPreActionHook update containerHookStatus [%+v]", containerHookStatus)
				//ctrl.updateHookStatus(hookUpdate)

			} else if containerHookStatus.PreActionStatus.ActionTimestamp != nil && containerHookStatus.PreActionStatus.ActionSucceed == nil {
				// Hook started but not complete yet
				// Wait for it to come back
				// if success, set ActionSucceed = true
				// else, set ActionSucceed = false and
				// set hook.Status.PreActionSucceed = false
				// log an event and bail out
			}
			if containerHookStatus.PreActionStatus.ActionTimestamp != nil && containerHookStatus.PreActionStatus.ActionSucceed != nil && *(containerHookStatus.PreActionStatus.ActionSucceed) == false {
				// It failed for this container preaction, bail out
				ctrl.updateHookStatus(hookUpdate)
				return fmt.Errorf("Failed to run PreAction %s in container %s in pod %s/%s", hookUpdate.Spec.PreAction.Action.Exec.Command, containerHookStatus.ContainerName, hookUpdate.Namespace, containerHookStatus.PodName)
			}
			//klog.V(5).Infof("msg: syncPreActionHook update containerHookStatus [%+v]", containerHookStatus)
			//ctrl.updateHookStatus(hookUpdate)
		}

		// successful
		if len(statuses) > 0 {
			hookUpdate.Status.ContainerExecutionHookStatuses = statuses
			ctrl.updateHookStatus(hookUpdate)
		}

		// Done with all containerHookStatus, set summary status to true
		// if it is not set to false yet, it means not failure occurred for a hook action on any container
		if hookUpdate.Status.PreActionSucceed == nil {
			//hookNew, err := ctrl.hookLister.ExecutionHooks(hookUpdate.Namespace).Get(hookUpdate.Name)
			hookNew, err := ctrl.clientset.ExecutionhookV1alpha1().ExecutionHooks(hookUpdate.Namespace).Get(hookUpdate.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("failed to retrieve hook %s from the informer: %q", hookUpdate.Name, err)
				return fmt.Errorf("failed to retrieve hook %s from the informer: %q", hookUpdate.Name, err)
			}
			klog.V(5).Infof("msg: syncPreActionHook update PreAction summary status with hook [%+v]", hookNew)
			hookNew2 := hookNew.DeepCopy()
			hookNew2.Status.PreActionSucceed = &statusTrue
			ctrl.updateHookStatus(hookNew2)
			klog.V(5).Infof("PreAction %s ran successfully in selected containers and pods", hookNew.Spec.PreAction.Action.Exec.Command)
			klog.V(5).Infof("msg: Finished syncPreActionHook with hook [%+v]", hookNew2)
			// Need an event?
		}
	}

	return nil
}

// syncPostActionHook is the main controller method to decide what to do with a hook whose PostActionSucceed is not set to true.
func (ctrl *hookController) syncPostActionHook(hook *crdv1.ExecutionHook) error {
	statusFalse := false
	statusTrue := true
	klog.V(5).Infof("Entering syncPostActionHook with hook [%+v]", hook)

	// TODO: Make sure hook.Spec.PodContainerNamesList is populated before the next step
	if len(hook.Spec.PodContainerNamesList) == 0 {
		return fmt.Errorf("hook.Spec.PodContainerNamesList cannot be empty [%+v]", hook.Spec.PodContainerNamesList)
	}

	// hook.Status is of type crdv1.ExecutionHookStatus
	if hook.Status.PostActionSucceed == nil {
		if len(hook.Status.ContainerExecutionHookStatuses) == 0 {
			return fmt.Errorf("hook.Status.ContainerExecutionHookStatuses should not be empty at this point [%+v]. It should have been initialized in syncPostActionHook. Something is wrong.", hook.Status.ContainerExecutionHookStatuses)
		}

		hookClone := hook.DeepCopy()

		// Loop through ContainerExecutionHookStatuses
		for _, containerHookStatus := range hookClone.Status.ContainerExecutionHookStatuses {
			// Hook PostAction has not started running if timestamp is nil
			// so run it now
			if containerHookStatus.PostActionStatus == nil {
				containerHookStatus.PostActionStatus = &crdv1.ExecutionHookActionStatus{}
			}

			if containerHookStatus.PostActionStatus.ActionTimestamp == nil {
				// Set timestamp now
				var timeNow = metav1.Time{
					Time: time.Now(),
				}
				containerHookStatus.PostActionStatus.ActionTimestamp = &timeNow
				// Run execution hook action
				// ExecPodContainerCommand will wait until it is done
				// TODO: add retry and timeout logic
				err := ExecPodContainerCommand(hook.Namespace, containerHookStatus.PodName, containerHookStatus.ContainerName, hook.Name, hook.Spec.PostAction.Action.Exec.Command, hook.Spec.PostAction.ActionTimeoutSeconds)

				// wait for it to come back
				// if success, set ActionSucceed = true
				// else, set ActionSucceed = false and
				// set hook.Status.PostActionSucceed = false
				// TODO: log an event
				// and bail out
				if err != nil { // failed on one command, no need to proceed, bail out
					containerHookStatus.PostActionStatus.ActionSucceed = &statusFalse
					hookClone.Status.PostActionSucceed = &statusFalse
					ctrl.updateHookStatus(hookClone)
					return fmt.Errorf("Failed to run PostAction %s in container %s in pod %s/%s", hookClone.Spec.PostAction.Action.Exec.Command, containerHookStatus.ContainerName, hookClone.Namespace, containerHookStatus.PodName)
				}
				// success
				containerHookStatus.PostActionStatus.ActionSucceed = &statusTrue
			} else if containerHookStatus.PostActionStatus.ActionTimestamp != nil && containerHookStatus.PostActionStatus.ActionSucceed == nil {
				// Hook started but not complete yet
				// Wait for it to come back
				// if success, set ActionSucceed = true
				// else, set ActionSucceed = false and
				// set hook.Status.PostActionSucceed = false
				// log an event and bail out
			}
			if containerHookStatus.PostActionStatus.ActionTimestamp != nil && containerHookStatus.PostActionStatus.ActionSucceed != nil && *(containerHookStatus.PostActionStatus.ActionSucceed) == false {
				// It failed for this container postaction, bail out
				ctrl.updateHookStatus(hookClone)
				return fmt.Errorf("Failed to run PostAction %s in container %s in pod %s/%s", hookClone.Spec.PostAction.Action.Exec.Command, containerHookStatus.ContainerName, hookClone.Namespace, containerHookStatus.PodName)
			}
			ctrl.updateHookStatus(hookClone)
		}

		// Done with all containerHookStatus, set summary status to true
		// if it is not set to false yet, it means not failure occurred for a hook action on any container
		if hookClone.Status.PostActionSucceed == nil {
			hookClone.Status.PostActionSucceed = &statusTrue
			ctrl.updateHookStatus(hookClone)
			klog.V(5).Infof("PostAction %s ran successfully in selected containers and pods", hookClone.Spec.PostAction.Action.Exec.Command)
			// Need an event?
		}
	}

	return nil
}

// updateHookStatus updates the status of the hook
// Note that hook passed in should be a clone of the original hook to make sure
// object update is successful
func (ctrl *hookController) updateHookStatus(hook *crdv1.ExecutionHook) (*crdv1.ExecutionHook, error) {
	klog.V(5).Infof("updateHookStatus[%s]", hookKey(hook))

	newHook, err := ctrl.clientset.ExecutionhookV1alpha1().ExecutionHooks(hook.Namespace).Update(hook)
	if err != nil {
		klog.V(4).Infof("updating ExecutionHook[%s] error status failed %v", hookKey(hook), err)
		return nil, err
	}

	_, err = ctrl.storeHookUpdate(newHook)
	if err != nil {
		klog.V(4).Infof("updating ExecutionHook[%s] error status: cannot update internal cache %v", hookKey(hook), err)
		return nil, err
	}
	//ctrl.eventRecorder.Event(newHook, eventtype, reason, message)
	//ctrl.eventRecorder.Event(newHook, eventtype, reason, message)

	return newHook, nil
}

func (ctrl *hookController) storeHookUpdate(hook interface{}) (bool, error) {
	return storeObjectUpdate(ctrl.hookStore, hook, "hook")
}

func (ctrl *hookController) storeHookTemplateUpdate(hookTemplate interface{}) (bool, error) {
	return storeObjectUpdate(ctrl.hookTemplateStore, hookTemplate, "hookTemplate")
}

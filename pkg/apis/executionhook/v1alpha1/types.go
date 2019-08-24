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

package v1alpha1

import (
	core_v1 "k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// HookActionResourcePlural is "hookactions"
	HookActionResourcePlural = "hookactions"
	// ExecutionHookResourcePlural is "executionhooks"
	ExecutionHookResourcePlural = "executionhooks"
)

// ExecutionHookSpec defines the desired state of ExecutionHook
// HookActionName is copied to ExecutionHookSpec by the controller such as
// the Snapshot Controller.
type ExecutionHookSpec struct {
	// PodSelection defines how to select pods and containers to run
	// the executionhook. If multiple pod/containers are selected, the action will executed on them
	// asynchronously. If execution ordering is required, caller has to implement the logic and create
	// different hooks in order.
	// This field is required.
	PodSelection PodSelection `json:"podSelection" protobuf:"bytes,1,opt,name=podSelection"`

	// Name of the HookAction. This is required.
	ActionName string `json:"actionName" protobuf:"bytes,2,opt,name=actionName"`
}

// PodSelection contains two fields, PodContainerNamesList and PodContainerSelector,
// where one of them must be defined so that the hook controller knows where to
// run the hook.
type PodSelection struct {
	// PodContainerNamesList lists the pods/containers on which the ExecutionHook
	// should be executed. If not specified, the ExecutionHook controller will find
	// all pods and containers based on PodContainerSelector.
	// If both PodContainerNamesList and PodContainerSelector are not
	// specified, the ExecutionHook cannot be executed and it will fail.
	// +optional
	PodContainerNamesList []PodContainerNames `json:"podContainerNamesList,omitempty" protobuf:"bytes,1,rep,name=podContainerNamesList"`

	// PodContainerSelector is for hook controller to find pods and containers
	// based on the pod label selector and container names
	// If PodContainerNamesList is specified, this field will not be used.
	// +optional
	PodContainerSelector *PodContainerSelector `json:"podContainerSelector,omitempty" protobuf:"bytes,2,opt,name=podContainerSelector"`
}

type PodContainerNames struct {
	// This field is required
	PodName string `json:"podName" protobuf:"bytes,1,opt,name=podName"`

	// +optional
	ContainerNames []string `json:"containerNames,omitempty" protobuf:"bytes,2,rep,name=containerNames"`
}

type PodContainerSelector struct {
	// PodSelector specifies a label query over a set of pods.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty" protobuf:"bytes,1,opt,name=podSelector"`

	// If specified, controller only select the containers that are listed from the selected pods based on PodSelector.
	// Otherwise, all containers of the pods will be selected
	// +optional
	ContainerList []string `json:"containerList,omitempty" protobuf:"bytes,2,rep,name=containerList"`
}

type ExecutionHookAction struct {
	// This field is required.
	Action *core_v1.Handler `json:"action" protobuf:"bytes,1,opt,name=action"`

	// +optional
	ActionTimeoutSeconds *int64 `json:"actionTimeoutSeconds,omitempty" protobuf:"bytes,2,opt,name=actionTimeoutSeconds"`
}

// ExecutionHookStatus represents the current state of a hook
type ExecutionHookStatus struct {
	// This is a list of ContainerExecutionHookStatus, with each status representing
	// information about how hook is executed in a container, including pod name,
	// container name, ActionTimestamp, ActionSucceed, etc.
	// +optional
	HookStatuses []ContainerExecutionHookStatus `json:"containerExecutionHookStatuses,omitempty" protobuf:"bytes,1,rep,name=containerExecutionHookStatuses"`
}

// ContainerExecutionHookStatus represents the current state of a hook for a specific
// container in a pod
type ContainerExecutionHookStatus struct {
	// This field is required
	PodName string `json:"podName" protobuf:"bytes,1,opt,name=podName"`

	// This field is required
	ContainerName string `json:"containerName" protobuf:"bytes,2,opt,name=containerName"`

	// If not set, it is nil, indicating Action has not started
	// If set, it means Action has started at the specified time
	// +optional
	Timestamp *metav1.Time `json:"actionTimestamp,omitempty" protobuf:"bytes,1,opt,name=actionTimestamp"`

	// Succeed is set to true when the action is executed in the container successfully.
	// It will be set to false if the action cannot be executed successfully after
	// ActionTimeoutSeconds passes.
	// +optional
	Succeed *bool `json:"actionSucceed,omitempty" protobuf:"varint,2,opt,name=actionSucceed"`

	// The last error encountered when executing the action. The hook controller might
	// update this field each time it retries the execution.
	// +optional
	Error *HookError `json:"error,omitempty" protobuf:"bytes,3,opt,name=error"`
}

type HookError struct {
	// Type of the error
	// This is required
	ErrorType ErrorType

	// Error message
	// +optional
	Message *string

	// More detailed reason why error happens
	// +optional
	Reason *string

	// It indicates when the error occurred
	// +optional
	Timestamp *metav1.Time
}

type ErrorType string

// More error types could be added, e.g., Forbidden, Unauthorized, AlreadyInProgress, etc.
const (
	// The execution hook times out
	Timeout ErrorType = "Timeout"

	// The execution hook fails with an error
	Error ErrorType = "Error"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExecutionHook is the Schema for the executionhooks API
// ExecutionHook is in the tenant namespace
// +k8s:openapi-gen=true
type ExecutionHook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a hook.
	// +optional
	Spec ExecutionHookSpec `json:"spec,omitempty"`

	// Status defines the current state of a hook.
	// +optional
	Status ExecutionHookStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExecutionHookList contains a list of ExecutionHook
type ExecutionHookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecutionHook `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HookAction describes action commands to run on pods/containers based
// on specified policies. HookAction will be created by the user and
// can be re-used later. Snapshot Controller will create ExecutionHooks
// based on HookActions specified in the snapshot spec. For example,
// two HookActions, preSnapshotExecutionHook and postSnapshotExecutionHook,
// are expected in the snapshot spec.
// HookAction does not contain information on pods/containers because those are
// runtime info.
// HookAction is namespaced.
// +k8s:openapi-gen=true
type HookAction struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// This contains the command to run on a container.
	// The command should be idempotent because the system does not guarantee exactly-once semantics.
	// Any action may be triggered more than once but only the latest results will be logged in status.
	// As alpha feature, only ExecAction type in Handler will be support, not the HTTPGETAction or TCPSocketAction.
	// This is required.
	Action core_v1.Handler `json:"action" protobuf:"bytes,2,opt,name=action"`

	// ActionTimeoutSeconds defines when the execution hook controller should stop retrying.
	// If execution fails, the execution hook controller will keep retrying until reaching
	// ActionTimeoutSeconds. If execution still fails or hangs, execution hook controller
	// stops retrying and updates executionhook status to failed.
	// If controller loses its state, counter restarts. In this case, controller will retry
	// for at least this long, before stopping.
	// Once an action is started, controller has no way to stop it even if
	// ActionTimeoutSeconds is exceeded. This simply controls if retry happens or not.
	// retry is based on exponential backoff policy. If ActionTimeoutSeconds is not
	// specified, it will retry until the hook object is deleted.
	// +optional
	ActionTimeoutSeconds *int64 `json:"actionTimeoutSeconds,omitempty" protobuf:"varint,3,opt,name=actionTimeoutSeconds"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HookActionList contains a list of HookAction
type HookActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HookAction `json:"items"`
}

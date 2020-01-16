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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ExecutionHookFinalizer allows ExecutionHook controller to clean up the custom resource before
	// removing it from the apiserver.
	ExecutionHookFinalizer = "executionhook.apps.x-k8s.io"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=executionhook,shortName=eh,scope=Namespaced,categories=executionhook
// +kubebuilder:subresource:status

// ExecutionHook is the Schema for the executionhook API
type ExecutionHook struct {
	metav1.TypeMeta `json:",inline"`
	// Metadata associated with persisted resources
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ExecutionHookSpec defines the desired state of ExecutionHook
	Spec ExecutionHookSpec `json:"spec,omitempty"`
	// ExecutionHookStatus defines the observed state of ExecutionHook
	// +optional
	Status *ExecutionHookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExecutionHookList contains a list of ExecutionHook
type ExecutionHookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecutionHook `json:"items"`
}

// ExecutionHookSpec defines the desired state of ExecutionHook
// HookActionName is copied to ExecutionHookSpec by the controller such as
// the Application Snapshot Controller.
type ExecutionHookSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PodSelection defines how to select pods and containers to run
	// the executionhook. If multiple pod/containers are selected, the action will executed on them
	// asynchronously. If execution ordering is required, caller has to implement the logic and create
	// different hooks in order.
	// This field is required.
	PodSelection PodSelection `json:"podSelection"`

	// Name of the HookAction. This is required.
	ActionName string `json:"actionName"`
}

// PodSelection contains two fields, PodContainerNamesList and PodContainerSelector,
// where exactly one of them must be defined so that the hook controller knows where to
// run the hook.
// TODO: Add validation to ensure exactly one of them are defined.
type PodSelection struct {
	// PodContainerNamesList lists the pods/containers on which the ExecutionHook
	// should be executed. If not specified, the ExecutionHook controller will find
	// all pods and containers based on PodContainerSelector.
	// If both PodContainerNamesList and PodContainerSelector are not
	// specified, the ExecutionHook cannot be executed and it will fail.
	// +optional
	PodContainerNamesList []PodContainerNames `json:"podContainerNamesList,omitempty"`

	// PodContainerSelector is for hook controller to find pods and containers
	// based on the pod label selector and container names
	// If PodContainerNamesList is specified, this field will not be used.
	// +optional
	PodContainerSelector *PodContainerSelector `json:"podContainerSelector,omitempty"`
}

// PodContainerNames lists the containers the ExecutionHook should be executed on in a Pod.
type PodContainerNames struct {
	// This field is required
	PodName string `json:"podName"`

	// +optional
	// If empty, hook action will be executed on all containers in the pod.
	ContainerNames []string `json:"containerNames,omitempty"`
}

// PodContainerSelector defines the selector and containers the ExecutionHook
// should be executed on.
type PodContainerSelector struct {
	// PodSelector specifies a label query over a set of pods.
	// If not specified, all pods in the namespace will be slected to look for containers
	// specified in the ContainerList.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// If specified, controller will only select these containers from the selected pods based on PodSelector.
	// Otherwise, all containers of the selected pods will be chosen for hook execution.
	// +optional
	ContainerList []string `json:"containerList,omitempty"`
}

// ExecutionHookStatus defines the observed state of ExecutionHook
type ExecutionHookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// This is a list of ContainerExecutionHookStatus, with each status representing
	// information about how hook is executed in a container, including pod name,
	// container name, Timestamp, Succeed, etc.
	// +optional
	HookStatuses []ContainerExecutionHookStatus `json:"hookStatuses,omitempty"`
}

// ContainerExecutionHookStatus represents the current state of a hook for a specific
// container in a pod
type ContainerExecutionHookStatus struct {
	// This field is required
	PodName string `json:"podName"`

	// This field is required
	ContainerName string `json:"containerName"`

	// If not set, it is nil, indicating Action has not started
	// If set, it means Action has started at the specified time
	// +optional
	Timestamp *metav1.Time `json:"timestamp,omitempty"`

	// Succeed is set to true when the action is executed in the container successfully.
	// It will be set to false if the action cannot be executed successfully after
	// ActionTimeoutSeconds passes.
	// +optional
	Succeed *bool `json:"succeed,omitempty"`

	// The last error encountered when executing the action. The hook controller might
	// update this field each time it retries the execution.
	// +optional
	Error *HookError `json:"error,omitempty"`
}

// HookError describes the error occurred from hook execution.
type HookError struct {
	// Type of the error
	// This is required
	ErrorType ErrorType `json:"errorType"`

	// Error message
	// +optional
	Message *string `json:"message,omitempty"`

	// More detailed reason why error happens
	// +optional
	Reason *string `json:"reason,omitempty"`

	// It indicates when the error occurred
	// +optional
	Timestamp *metav1.Time `json:"timestamp,omitempty"`
}

// ErrorType defines the type of error occurred from hook execution.
type ErrorType string

// More error types could be added, e.g., Forbidden, Unauthorized, AlreadyInProgress, etc.
const (
	// The execution hook times out
	Timeout ErrorType = "Timeout"

	// The execution hook fails with an error
	Error ErrorType = "Error"
)

func init() {
	SchemeBuilder.Register(&ExecutionHook{}, &ExecutionHookList{})
}

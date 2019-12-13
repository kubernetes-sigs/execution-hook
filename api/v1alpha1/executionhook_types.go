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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExecutionHookSpec defines the desired state of ExecutionHook
// HookActionName is copied to ExecutionHookSpec by the controller such as
// the Snapshot Controller.
type ExecutionHookSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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

// PodContainerNames lists the containers the ExecutionHook should be executed
// on in a Pod.
type PodContainerNames struct {
	// This field is required
	PodName string `json:"podName" protobuf:"bytes,1,opt,name=podName"`

	// +optional
	ContainerNames []string `json:"containerNames,omitempty" protobuf:"bytes,2,rep,name=containerNames"`
}

// PodContainerSelector defines the selector and containers the ExecutionHook
// should be executed on.
type PodContainerSelector struct {
	// PodSelector specifies a label query over a set of pods.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty" protobuf:"bytes,1,opt,name=podSelector"`

	// If specified, controller only select the containers that are listed from the selected pods based on PodSelector.
	// Otherwise, all containers of the pods will be selected
	// +optional
	ContainerList []string `json:"containerList,omitempty" protobuf:"bytes,2,rep,name=containerList"`
}

// ExecutionHookStatus defines the observed state of ExecutionHook
type ExecutionHookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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
	Timestamp *metav1.Time `json:"actionTimestamp,omitempty" protobuf:"bytes,3,opt,name=actionTimestamp"`

	// Succeed is set to true when the action is executed in the container successfully.
	// It will be set to false if the action cannot be executed successfully after
	// ActionTimeoutSeconds passes.
	// +optional
	Succeed *bool `json:"actionSucceed,omitempty" protobuf:"varint,4,opt,name=actionSucceed"`

	// The last error encountered when executing the action. The hook controller might
	// update this field each time it retries the execution.
	// +optional
	Error *HookError `json:"error,omitempty" protobuf:"bytes,5,opt,name=error"`
}

// HookError describes the error occurred from hook execution.
type HookError struct {
	// Type of the error
	// This is required
	ErrorType ErrorType `json:"errorType" protobuf:"bytes,1,opt,name=errorType"`

	// Error message
	// +optional
	Message *string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`

	// More detailed reason why error happens
	// +optional
	Reason *string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// It indicates when the error occurred
	// +optional
	Timestamp *metav1.Time `json:"timestamp,omitempty" protobuf:"bytes,4,opt,name=timestamp"`
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

// +kubebuilder:object:root=true

// ExecutionHook is the Schema for the executionhooks API
type ExecutionHook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecutionHookSpec   `json:"spec,omitempty"`
	Status ExecutionHookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExecutionHookList contains a list of ExecutionHook
type ExecutionHookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecutionHook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExecutionHook{}, &ExecutionHookList{})
}

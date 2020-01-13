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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:resource:path=hookaction,shortName=ha,scope=Namespaced,categories=hookaction
// +kubebuilder:object:root=true

// HookAction is the Schema for the hookactions API
type HookAction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// This contains the command to run in a container.
	// The command should be idempotent because the system does not guarantee exactly-once semantics.
	// Any action may be triggered more than once but only the latest results will be logged in status.
	// As alpha feature, only ExecAction type in Handler will be support, not the HTTPGETAction or TCPSocketAction.
	// This is required.
	Action corev1.Handler `json:"action"`

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
	ActionTimeoutSeconds int64 `json:"timeout"`
}

// +kubebuilder:object:root=true

// HookActionList contains a list of HookAction
type HookActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HookAction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HookAction{}, &HookActionList{})
}

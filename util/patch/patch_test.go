/*
Copyright 2017 The Kubernetes Authors.

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

package patch

import (
	"context"
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	appsv1alpha1 "sigs.k8s.io/execution-hook/api/v1alpha1"
)

const (
	timeout = time.Second * 10
)

func TestHelperPatchExecutionHook(t *testing.T) {
	testCases := []struct {
		name               string
		before             *appsv1alpha1.ExecutionHook
		after              *appsv1alpha1.ExecutionHook
		hookNamespacedName types.NamespacedName
		wantErr            bool
	}{
		{
			name: "No change between before and after",
			hookNamespacedName: types.NamespacedName{
				Name:      "test-exechook",
				Namespace: "test-namespace",
			},
			before: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-exechook",
					Namespace: "test-namespace",
				},
				Spec: appsv1alpha1.ExecutionHookSpec{
					ActionName: "exec-action1",
				},
			},
			after: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-exechook",
					Namespace: "test-namespace",
				},
				Spec: appsv1alpha1.ExecutionHookSpec{
					ActionName: "exec-action1",
				},
			},
		},
		{
			name: "Only remove finalizer update",
			hookNamespacedName: types.NamespacedName{
				Name:      "test-exechook",
				Namespace: "test-namespace",
			},
			before: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
					Finalizers: []string{
						appsv1alpha1.ExecutionHookFinalizer,
					},
				},
			},
			after: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
				},
			},
		},
		{
			name: "Only spec update",
			hookNamespacedName: types.NamespacedName{
				Name:      "test-exechook",
				Namespace: "test-namespace",
			},
			before: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
				},
			},
			after: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
				},
				Spec: appsv1alpha1.ExecutionHookSpec{
					ActionName: "test-action",
					PodSelection: appsv1alpha1.PodSelection{
						PodContainerNamesList: []appsv1alpha1.PodContainerNames{
							{
								PodName:        "test-exechook-pod-1",
								ContainerNames: []string{"c1", "c2", "c3"},
							},
							{
								PodName:        "test-exechook-pod-2",
								ContainerNames: []string{"c1", "c2", "c3"},
							},
						},
					},
				},
			},
		},
		{
			name: "Only status update",
			hookNamespacedName: types.NamespacedName{
				Name:      "test-exechook",
				Namespace: "test-namespace",
			},
			before: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
				},
			},
			after: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
				},
				Status: appsv1alpha1.ExecutionHookStatus{
					HookStatuses: []appsv1alpha1.ContainerExecutionHookStatus{
						{
							PodName:       "exechook-pod",
							ContainerName: "exechook-container",
						},
					},
				},
			},
		},
		{
			name: "Update spec and status",
			hookNamespacedName: types.NamespacedName{
				Name:      "test-exechook",
				Namespace: "test-namespace",
			},
			before: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
				},
			},
			after: &appsv1alpha1.ExecutionHook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-executionhook",
					Namespace: "test-namespace",
					Finalizers: []string{
						appsv1alpha1.ExecutionHookFinalizer,
					},
				},
				Status: appsv1alpha1.ExecutionHookStatus{
					HookStatuses: []appsv1alpha1.ContainerExecutionHookStatus{
						{
							PodName:       "exechook-pod",
							ContainerName: "exechook-container-1",
						},
						{PodName: "exechook-pod",
							ContainerName: "exechook-container-2",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(appsv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
			ctx := context.Background()
			fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme)

			beforeCopy := tc.before.DeepCopyObject()
			g.Expect(fakeClient.Create(ctx, beforeCopy)).To(Succeed())

			h, err := NewHelper(beforeCopy, fakeClient)
			if err != nil {
				t.Fatalf("Expected no error initializing helper: %v", err)
			}

			afterCopy := tc.after.DeepCopyObject()
			if err := h.Patch(ctx, afterCopy); (err != nil) != tc.wantErr {
				t.Errorf("Helper.Patch() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !reflect.DeepEqual(tc.after, afterCopy) {
				t.Errorf("Expected after to be the same after patching\n tc.after: %v\n afterCopy: %v\n", tc.after, afterCopy)
			}
		})
	}
}

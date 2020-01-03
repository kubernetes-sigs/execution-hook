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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "sigs.k8s.io/execution-hook/api/v1alpha1"
)

func getTestPods() []runtime.Object {
	return []runtime.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app1-1",
				Namespace: "test-ns",
				Labels: map[string]string{
					"app":         "app1",
					"owner":       "admin",
					"environment": "prod",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "my-awesome-app",
					},
					{
						Name: "auth-proxy",
					},
					{
						Name: "metrics-proxy",
					},
					{
						Name: "rbac-proxy",
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app1-1",
				Namespace: "super-ns",
				Labels: map[string]string{
					"app":         "app1",
					"owner":       "admin",
					"environment": "prod",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "my-awesome-app",
					},
					{
						Name: "auth-proxy",
					},
					{
						Name: "metrics-proxy",
					},
					{
						Name: "rbac-proxy",
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app1-2",
				Namespace: "test-ns",
				Labels: map[string]string{
					"app":         "app1",
					"owner":       "admin",
					"environment": "test",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "my-awesome-app",
					},
					{
						Name: "auth-proxy",
					},
					{
						Name: "metrics-proxy",
					},
					{
						Name: "rbac-proxy",
					},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		},
	}
}

var _ = Describe("selectPodContainers", func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("Select PodContainer to run execution hook action", func() {
		It("should returns nil result and error when PodSelection is nil", func() {
			r := &ExecutionHookReconciler{
				Client: k8sClient,
				Log:    log.Log,
			}
			actual, err := r.selectPodContainers("test-ns", nil)
			Expect(err).NotTo(BeNil())
			Expect(actual).To(BeNil())
		})
		It("should returns empty result and nil error when label selector in the PodSelection.PodSelector is nil", func() {
			r := &ExecutionHookReconciler{
				Client: k8sClient,
				Log:    log.Log,
			}
			podSelector := appsv1alpha1.PodSelection{
				PodContainerNamesList: nil,
				PodContainerSelector: &appsv1alpha1.PodContainerSelector{
					PodSelector: nil,
				},
			}
			actual, err := r.selectPodContainers("test-ns", &podSelector)
			Expect(err).To(BeNil())
			Expect(actual).To(BeEmpty())

		})
		It("should return pod with multiple containers matching labelSelector with MatchLables only", func() {
			// setup test pods
			objs := getTestPods()

			r := &ExecutionHookReconciler{
				Client: fake.NewFakeClient(objs...),
				Log:    log.Log,
			}

			podSelector := appsv1alpha1.PodSelection{
				PodContainerNamesList: nil,
				PodContainerSelector: &appsv1alpha1.PodContainerSelector{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":   "app1",
							"owner": "admin",
						},
					},
				},
			}

			podContainers, err := r.selectPodContainers("test-ns", &podSelector)
			expected := []appsv1alpha1.PodContainerNames{
				{
					PodName:        "app1-1",
					ContainerNames: []string{"my-awesome-app", "auth-proxy", "metrics-proxy", "rbac-proxy"},
				},
				{
					PodName:        "app1-2",
					ContainerNames: []string{"my-awesome-app", "auth-proxy", "metrics-proxy", "rbac-proxy"},
				},
			}

			Expect(err).To(BeNil())
			Expect(podContainers).To(BeEquivalentTo(expected))
		})

		It("Should return pod with multiple containers matching labelSelector with MatchLables and MatchExpressions", func() {
			objs := getTestPods()
			r := &ExecutionHookReconciler{
				Client: fake.NewFakeClient(objs...),
				Log:    log.Log,
			}

			podSelector := appsv1alpha1.PodSelection{
				PodContainerNamesList: nil,
				PodContainerSelector: &appsv1alpha1.PodContainerSelector{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":   "app1",
							"owner": "admin",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "environment",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"test", "staging"},
							},
						},
					},
				},
			}

			podContainers, err := r.selectPodContainers("test-ns", &podSelector)
			expected := []appsv1alpha1.PodContainerNames{
				{PodName: "app1-1",
					ContainerNames: []string{"my-awesome-app", "auth-proxy", "metrics-proxy", "rbac-proxy"}},
			}

			Expect(err).To(BeNil())
			Expect(podContainers).To(BeEquivalentTo(expected))
		})

		It("should return podContainerNamesList when supplied", func() {
			objs := getTestPods()
			r := &ExecutionHookReconciler{
				Client: fake.NewFakeClient(objs...),
				Log:    log.Log,
			}

			podSelector := appsv1alpha1.PodSelection{
				PodContainerNamesList: []appsv1alpha1.PodContainerNames{
					{
						PodName:        "some-random=pod",
						ContainerNames: []string{"c1", "c2", "c3"},
					},
				},
				PodContainerSelector: &appsv1alpha1.PodContainerSelector{
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":   "app1",
							"owner": "admin",
						},
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "environment",
								Operator: metav1.LabelSelectorOpNotIn,
								Values:   []string{"test", "staging"},
							},
						},
					},
				},
			}

			acutal, err := r.selectPodContainers("test-ns", &podSelector)
			expected := []appsv1alpha1.PodContainerNames{
				{
					PodName:        "some-random=pod",
					ContainerNames: []string{"c1", "c2", "c3"},
				},
			}
			Expect(err).To(BeNil())
			Expect(acutal).To(BeEquivalentTo(expected))

		})
	})
})

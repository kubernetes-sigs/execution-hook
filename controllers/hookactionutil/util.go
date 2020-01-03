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

package hookactionutil

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

// ExecuteActionCommand executes ExecAction in the supplied ns/pod/container
func ExecuteActionCommand(ea *corev1.ExecAction, ns, pod, container, hookName, actionName string, kubeClient kubernetes.Interface, log logr.Logger) error {
	if ea == nil || len(ea.Command) == 0 {
		return errors.Errorf("for hook-action %s, ExecAction.Command is required", actionName)
	}
	podapi := new(corev1.Pod)
	if !containerExists(ns, pod, container, podapi) {
		return errors.Errorf("in namespace %s, container %s/%s does not exist", ns, pod, container)
	}
	log.WithValues(
		"executionhook", hookName,
		"hook-action", actionName,
		"namespace", ns,
		"pod", pod,
		"container", container)
	log.Info("Running ExecAction", "command", fmt.Sprintf("[%s", strings.Join(ea.Command, " ")))
	return nil
}

func containerExists(ns, podName, container string, podAPI *corev1.Pod) bool {
	// TODO (ashish-amarnath) Implement this
	return true
}

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

package hookactionrunner

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecActionRunner implements the HookActionRunner interface to run exec hook actions
type ExecActionRunner struct {
	// TODO (ashish-amarnath) exec Action params
	Action     *corev1.ExecAction
	Namespace  string
	Pod        string
	Container  string
	HookName   string
	ActionName string
	Log        logr.Logger
	RestConfig *rest.Config
}

// Run executes the ExecHookAction
func (e *ExecActionRunner) Run() error {
	if e.Action == nil || len(e.Action.Command) == 0 {
		return errors.Errorf("for hook-action %s, ExecAction.Command is required", e.ActionName)
	}
	podapi := new(corev1.Pod)
	if !containerExists(e.Namespace, e.Pod, e.Container, podapi) {
		return errors.Errorf("in namespace %s, container %s/%s does not exist", e.Namespace, e.Pod, e.Container)
	}
	e.Log.WithValues(
		"executionhook", e.HookName,
		"hook-action", e.ActionName,
		"namespace", e.Namespace,
		"pod", e.Pod,
		"container", e.Container)
	e.Log.Info("Running ExecAction", "command", fmt.Sprintf("[%s]", strings.Join(e.Action.Command, " ")))

	kubeClient, err := kubernetes.NewForConfig(e.RestConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to instantiate kubernetes clientset to run hook action")
	}

	req := kubeClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(e.Namespace).
		Name(e.Pod).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: e.Container,
		Command:   e.Action.Command,
		Stdout:    true,
		Stderr:    true,
	}, kscheme.ParameterCodec)

	e.Log.Info("Setting up for remote hook action")
	executor, err := remotecommand.NewSPDYExecutor(e.RestConfig, "POST", req.URL())
	if err != nil {
		return errors.Wrapf(err, "failed to setup for hook execution in container")
	}
	var stdout, stderr bytes.Buffer

	streamOptions := remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}
	// TODO (ashish-amarnath) implement timeout scenario
	e.Log.Info("Executing remote command and waiting it for it to complete")
	err = executor.Stream(streamOptions)

	e.Log.Info(fmt.Sprintf("stdout: %s", stdout.String()))
	e.Log.Info(fmt.Sprintf("stderr: %s", stderr.String()))
	return err
}

func containerExists(ns, podName, container string, podAPI *corev1.Pod) bool {
	// TODO (ashish-amarnath) Implement this
	return true
}

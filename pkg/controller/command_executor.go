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
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	kapiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
)

const defaultTimeout = 30 * time.Second

// PodCommandExecutor is capable of executing a command in a container in a pod.
type PodCommandExecutor interface {
	// ExecutePodCommand executes a command in a container in a pod. If the command takes longer than
	// the specified timeout, an error is returned.
	ExecutePodCommand( /*log logrus.FieldLogger,*/ item map[string]interface{}, namespace, name, hookName string, hook *ExecHook) error
}

type poster interface {
	Post() *rest.Request
}

// TODO: This is a temp fix
// ExecHook is a hook that uses the pod exec API to execute a command in a container in a pod.
type ExecHook struct {
	// Container is the container in the pod where the command should be executed. If not specified,
	// the pod's first container is used.
	Container string `json:"container"`
	// Command is the command and arguments to execute.
	Command []string `json:"command"`
	// OnError specifies how Ark should behave if it encounters an error executing this hook.
	OnError string/*HookErrorMode*/ `json:"onError"`
	// Timeout defines the maximum amount of time Ark should wait for the hook to complete before
	// considering the execution a failure.
	Timeout time.Duration `json:"timeout"`
}

type defaultPodCommandExecutor struct {
	restClientConfig *rest.Config
	restClient       poster

	streamExecutorFactory streamExecutorFactory
}

// TODO: This is a temp fix
// GetSlice returns the slice at root[path], where path is a dot separated string.
func GetSlice(root map[string]interface{}, path string) ([]interface{}, error) {
	obj, err := GetValue(root, path)
	if err != nil {
		return nil, err
	}

	ret, ok := obj.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value at path %v is not a []interface{}", path)
	}

	return ret, nil
}

// GetValue returns the object at root[path], where path is a dot separated string.
func GetValue(root map[string]interface{}, path string) (interface{}, error) {
	if root == nil {
		return "", fmt.Errorf("root is nil")
	}

	pathParts := strings.Split(path, ".")
	key := pathParts[0]

	obj, found := root[pathParts[0]]
	if !found {
		return "", fmt.Errorf("key %v not found", pathParts[0])
	}

	if len(pathParts) == 1 {
		return obj, nil
	}

	subMap, ok := obj.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("value at key %v is not a map[string]interface{}", key)
	}

	return GetValue(subMap, strings.Join(pathParts[1:], "."))
}

// ExecutePodCommand uses the pod exec API to execute a command in a container in a pod. If the
// command takes longer than the specified timeout, an error is returned (NOTE: it is not currently
// possible to ensure the command is terminated when the timeout occurs, so it may continue to run
// in the background).
//func (e *defaultPodCommandExecutor) ExecutePodCommand( /*log logrus.FieldLogger,*/ item map[string]interface{}, namespace, name, hookName string, hook *ExecHook) error {
//func ExecutePodCommand(namespace, name, hookName string, hook *ExecHook) error {
// Note: Run command in one container in one pod
//func ExecutePodCommand(namespace, name, hookName string, hook *crdv1.ExecutionHook) error {
func ExecPodContainerCommand(namespace, podName, containerName, hookName string, command []string, timeoutSeconds *int64) error {
	// timeoutSeconds = 0, means no wait; -1 means infinite wait
	// Wait forever or no wait or wait for specified seconds
	/*if item == nil {
		return fmt.Errorf("item is required")
	}
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if hookName == "" {
		return fmt.Errorf("hookName is required")
	}
	if hook == nil {
		return fmt.Errorf("hook is required")
	}

	if hook.Container == "" {
		if err := setDefaultHookContainer(item, hook); err != nil {
			return err
		}
	} else if err := ensureContainerExists(item, hook.Container); err != nil {
		return err
	}

	if len(hook.Command) == 0 {
		return fmt.Errorf("command is required")
	}*/

	/*switch hook.OnError {
	case api.HookErrorModeFail, api.HookErrorModeContinue:
		// use the specified value
	default:
		// default to fail
		hook.OnError = api.HookErrorModeFail
	}*/

	//if hook.Timeout.Duration == 0 {
	//	hook.Timeout.Duration = defaultTimeout
	//}

	/*hookLog := log.WithFields(
		logrus.Fields{
			"hookName":      hookName,
			"hookContainer": hook.Container,
			"hookCommand":   hook.Command,
			"hookOnError":   hook.OnError,
			"hookTimeout":   hook.Timeout,
		},
	)
	hookLog.Info("running exec hook")
	*/

	klog.V(5).Infof("Enter ExecutePodCommand name %s namespace %s", podName, namespace)
	//e := &defaultPodCommandExecutor{}

	/*type defaultPodCommandExecutor struct {
	        restClientConfig *rest.Config
	        restClient       poster

	        streamExecutorFactory streamExecutorFactory
		}*/

	//podexec.NewPodCommandExecutor(s.kubeClientConfig, s.kubeClient.CoreV1().RESTClient()),
	//restClientConfig, err := rest.InClusterConfig()
	restClientConfig, err := buildConfig("") //("/var/run/kubernetes/admin.kubeconfig")
	if err != nil {
		klog.Errorf("ExecutePodCommand failed for name %s namespace %s. buildConfig failed %v", podName, namespace, err)
		return fmt.Errorf(err.Error())
	}
	klog.V(5).Infof("ExecutePodCommand: restClientConfig %#v", restClientConfig)
	kubeClient, err := kubernetes.NewForConfig(restClientConfig)
	e := &defaultPodCommandExecutor{
		restClientConfig:      restClientConfig,
		restClient:            kubeClient.CoreV1().RESTClient(),
		streamExecutorFactory: &defaultStreamExecutorFactory{},
	}
	if e.restClient == nil {
		klog.Errorf("ExecutePodCommand failed for name %s namespace %s. e.restClient is nil", podName, namespace)
		return fmt.Errorf("e.restClient is nil")
	}
	klog.V(5).Infof("defaultPodCommandExecutor %v", e)
	// TODO: namespace?
	req := e.restClient.Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("exec").
		Param("container", containerName)

	req.VersionedParams(&kapiv1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, kscheme.ParameterCodec)

	klog.V(5).Infof("ExecutePodCommand name %s namespace %s Container %s Command %s", podName, namespace, containerName, command)
	//restClientConfig,err := rest.InClusterConfig()
	//var streamExecutorFactory streamExecutorFactory
	klog.V(5).Infof("ExecutePodCommand name %s namespace %s Config %v URL %v", podName, namespace, restClientConfig, req.URL())

	if e.streamExecutorFactory == nil {
		klog.Errorf("ExecutePodCommand failed for name %s namespace %s. e.streamExecutorFactory is nil", podName, namespace)
		return fmt.Errorf("e.streamExecutorFactory is nil")
	}
	//executor, err := e.streamExecutorFactory.NewSPDYExecutor(restClientConfig, "POST", req.URL())
	executor, err := remotecommand.NewSPDYExecutor(restClientConfig, "POST", req.URL())
	if err != nil {
		klog.Errorf("ExecutePodCommand failed. name %s namespace %s Container %s Command %s", podName, namespace, containerName, command)
		return err
	}
	klog.Infof("NewSPDYExecutor: %v", executor)

	var stdout, stderr bytes.Buffer

	streamOptions := remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	/*err = executor.Stream(remotecommand.StreamOptions{
	        //Stdin:             stdin,
	        Stdout:            &stdout,
	        Stderr:            &stderr,
	        //Tty:               tty,
	        //TerminalSizeQueue: terminalSizeQueue,
	}) works!!!*/

	errCh := make(chan error)

	// TODO(xyang): Add background flag in the hook.
	// If background is true, need to run process in the background.
	// This is needed for mysql freeze script.
	// If background is false, need to run process normally.
	// This is for mysql unfreeze script and all other scripts.
	background := false
	if background {
		go func() {
			err = executor.Stream(streamOptions)
			errCh <- err
		}()
	} else {
		err = executor.Stream(streamOptions)
		klog.Infof("ExecutePodCommand: executor:Stream is complete.")
	}

	/*
		var timeoutCh <-chan time.Time
		if hook.Timeout > 0 {
			timer := time.NewTimer(hook.Timeout)
			defer timer.Stop()
			timeoutCh = timer.C
		}

		select {
		case err = <-errCh:
		case <-timeoutCh:
			return fmt.Errorf("timed out after %v", hook.Timeout)
		}*/

	klog.Infof("stdout: %s", stdout.String())
	klog.Infof("stderr: %s", stderr.String())
	klog.V(5).Infof("Exit ExecutePodCommand name %s namespace %s", podName, namespace)

	return err
}

func ensureContainerExists(pod map[string]interface{}, container string) error {
	containers, err := GetSlice(pod, "spec.containers")
	if err != nil {
		return err
	}
	for _, obj := range containers {
		c, ok := obj.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected type for container %T", obj)
		}
		name, ok := c["name"].(string)
		if !ok {
			return fmt.Errorf("unexpected type for container name %T", c["name"])
		}
		if name == container {
			return nil
		}
	}

	return fmt.Errorf("no such container: %q", container)
}

func setDefaultHookContainer(pod map[string]interface{}, hook *ExecHook) error {
	containers, err := GetSlice(pod, "spec.containers")
	if err != nil {
		return err
	}

	if len(containers) < 1 {
		return fmt.Errorf("need at least 1 container")
	}

	container, ok := containers[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected type for container %T", pod)
	}

	name, ok := container["name"].(string)
	if !ok {
		return fmt.Errorf("unexpected type for container name %T", container["name"])
	}
	hook.Container = name

	return nil
}

type streamExecutorFactory interface {
	NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error)
}

type defaultStreamExecutorFactory struct{}

func (f *defaultStreamExecutorFactory) NewSPDYExecutor(config *rest.Config, method string, url *url.URL) (remotecommand.Executor, error) {
	return remotecommand.NewSPDYExecutor(config, method, url)
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

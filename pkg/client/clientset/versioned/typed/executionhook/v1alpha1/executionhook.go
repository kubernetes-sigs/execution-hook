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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/kubernetes-csi/execution-hook/pkg/apis/executionhook/v1alpha1"
	scheme "github.com/kubernetes-csi/execution-hook/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ExecutionHooksGetter has a method to return a ExecutionHookInterface.
// A group's client should implement this interface.
type ExecutionHooksGetter interface {
	ExecutionHooks(namespace string) ExecutionHookInterface
}

// ExecutionHookInterface has methods to work with ExecutionHook resources.
type ExecutionHookInterface interface {
	Create(*v1alpha1.ExecutionHook) (*v1alpha1.ExecutionHook, error)
	Update(*v1alpha1.ExecutionHook) (*v1alpha1.ExecutionHook, error)
	UpdateStatus(*v1alpha1.ExecutionHook) (*v1alpha1.ExecutionHook, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ExecutionHook, error)
	List(opts v1.ListOptions) (*v1alpha1.ExecutionHookList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExecutionHook, err error)
	ExecutionHookExpansion
}

// executionHooks implements ExecutionHookInterface
type executionHooks struct {
	client rest.Interface
	ns     string
}

// newExecutionHooks returns a ExecutionHooks
func newExecutionHooks(c *ExecutionhookV1alpha1Client, namespace string) *executionHooks {
	return &executionHooks{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the executionHook, and returns the corresponding executionHook object, and an error if there is any.
func (c *executionHooks) Get(name string, options v1.GetOptions) (result *v1alpha1.ExecutionHook, err error) {
	result = &v1alpha1.ExecutionHook{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("executionhooks").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ExecutionHooks that match those selectors.
func (c *executionHooks) List(opts v1.ListOptions) (result *v1alpha1.ExecutionHookList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ExecutionHookList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("executionhooks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested executionHooks.
func (c *executionHooks) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("executionhooks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a executionHook and creates it.  Returns the server's representation of the executionHook, and an error, if there is any.
func (c *executionHooks) Create(executionHook *v1alpha1.ExecutionHook) (result *v1alpha1.ExecutionHook, err error) {
	result = &v1alpha1.ExecutionHook{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("executionhooks").
		Body(executionHook).
		Do().
		Into(result)
	return
}

// Update takes the representation of a executionHook and updates it. Returns the server's representation of the executionHook, and an error, if there is any.
func (c *executionHooks) Update(executionHook *v1alpha1.ExecutionHook) (result *v1alpha1.ExecutionHook, err error) {
	result = &v1alpha1.ExecutionHook{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("executionhooks").
		Name(executionHook.Name).
		Body(executionHook).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *executionHooks) UpdateStatus(executionHook *v1alpha1.ExecutionHook) (result *v1alpha1.ExecutionHook, err error) {
	result = &v1alpha1.ExecutionHook{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("executionhooks").
		Name(executionHook.Name).
		SubResource("status").
		Body(executionHook).
		Do().
		Into(result)
	return
}

// Delete takes name of the executionHook and deletes it. Returns an error if one occurs.
func (c *executionHooks) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("executionhooks").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *executionHooks) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("executionhooks").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched executionHook.
func (c *executionHooks) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExecutionHook, err error) {
	result = &v1alpha1.ExecutionHook{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("executionhooks").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
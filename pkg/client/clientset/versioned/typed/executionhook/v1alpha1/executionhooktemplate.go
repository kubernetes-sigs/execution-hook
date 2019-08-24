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

// ExecutionHookTemplatesGetter has a method to return a ExecutionHookTemplateInterface.
// A group's client should implement this interface.
type ExecutionHookTemplatesGetter interface {
	ExecutionHookTemplates(namespace string) ExecutionHookTemplateInterface
}

// ExecutionHookTemplateInterface has methods to work with ExecutionHookTemplate resources.
type ExecutionHookTemplateInterface interface {
	Create(*v1alpha1.ExecutionHookTemplate) (*v1alpha1.ExecutionHookTemplate, error)
	Update(*v1alpha1.ExecutionHookTemplate) (*v1alpha1.ExecutionHookTemplate, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.ExecutionHookTemplate, error)
	List(opts v1.ListOptions) (*v1alpha1.ExecutionHookTemplateList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExecutionHookTemplate, err error)
	ExecutionHookTemplateExpansion
}

// executionHookTemplates implements ExecutionHookTemplateInterface
type executionHookTemplates struct {
	client rest.Interface
	ns     string
}

// newExecutionHookTemplates returns a ExecutionHookTemplates
func newExecutionHookTemplates(c *ExecutionhookV1alpha1Client, namespace string) *executionHookTemplates {
	return &executionHookTemplates{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the executionHookTemplate, and returns the corresponding executionHookTemplate object, and an error if there is any.
func (c *executionHookTemplates) Get(name string, options v1.GetOptions) (result *v1alpha1.ExecutionHookTemplate, err error) {
	result = &v1alpha1.ExecutionHookTemplate{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ExecutionHookTemplates that match those selectors.
func (c *executionHookTemplates) List(opts v1.ListOptions) (result *v1alpha1.ExecutionHookTemplateList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ExecutionHookTemplateList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested executionHookTemplates.
func (c *executionHookTemplates) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a executionHookTemplate and creates it.  Returns the server's representation of the executionHookTemplate, and an error, if there is any.
func (c *executionHookTemplates) Create(executionHookTemplate *v1alpha1.ExecutionHookTemplate) (result *v1alpha1.ExecutionHookTemplate, err error) {
	result = &v1alpha1.ExecutionHookTemplate{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		Body(executionHookTemplate).
		Do().
		Into(result)
	return
}

// Update takes the representation of a executionHookTemplate and updates it. Returns the server's representation of the executionHookTemplate, and an error, if there is any.
func (c *executionHookTemplates) Update(executionHookTemplate *v1alpha1.ExecutionHookTemplate) (result *v1alpha1.ExecutionHookTemplate, err error) {
	result = &v1alpha1.ExecutionHookTemplate{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		Name(executionHookTemplate.Name).
		Body(executionHookTemplate).
		Do().
		Into(result)
	return
}

// Delete takes name of the executionHookTemplate and deletes it. Returns an error if one occurs.
func (c *executionHookTemplates) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *executionHookTemplates) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("executionhooktemplates").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched executionHookTemplate.
func (c *executionHookTemplates) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.ExecutionHookTemplate, err error) {
	result = &v1alpha1.ExecutionHookTemplate{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("executionhooktemplates").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	executionhookv1alpha1 "github.com/kubernetes-csi/execution-hook/pkg/apis/executionhook/v1alpha1"
	versioned "github.com/kubernetes-csi/execution-hook/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kubernetes-csi/execution-hook/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kubernetes-csi/execution-hook/pkg/client/listers/executionhook/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ExecutionHookTemplateInformer provides access to a shared informer and lister for
// ExecutionHookTemplates.
type ExecutionHookTemplateInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ExecutionHookTemplateLister
}

type executionHookTemplateInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewExecutionHookTemplateInformer constructs a new informer for ExecutionHookTemplate type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewExecutionHookTemplateInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredExecutionHookTemplateInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredExecutionHookTemplateInformer constructs a new informer for ExecutionHookTemplate type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredExecutionHookTemplateInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ExecutionhookV1alpha1().ExecutionHookTemplates(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ExecutionhookV1alpha1().ExecutionHookTemplates(namespace).Watch(options)
			},
		},
		&executionhookv1alpha1.ExecutionHookTemplate{},
		resyncPeriod,
		indexers,
	)
}

func (f *executionHookTemplateInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredExecutionHookTemplateInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *executionHookTemplateInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&executionhookv1alpha1.ExecutionHookTemplate{}, f.defaultInformer)
}

func (f *executionHookTemplateInformer) Lister() v1alpha1.ExecutionHookTemplateLister {
	return v1alpha1.NewExecutionHookTemplateLister(f.Informer().GetIndexer())
}
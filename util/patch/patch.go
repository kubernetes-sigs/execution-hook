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

package patch

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Helper is a utility for ensuring the proper Patching of resources
// and their status
type Helper struct {
	client        client.Client
	before        map[string]interface{}
	hasStatus     bool
	beforeStatus  interface{}
	resourcePatch client.Patch
	statusPatch   client.Patch
	log           logr.Logger
}

// Debug method
func logInterface(r string, l logr.Logger, o map[string]interface{}) {
	l.Info(r)
	for k, v := range o {
		l.Info(fmt.Sprintf("%q-> %v", k, v))
	}
}

// NewHelper returns an initialized Helper
func NewHelper(resource runtime.Object, crClient client.Client) (*Helper, error) {
	if resource == nil {
		return nil, errors.Errorf("expected non-nil resource")
	}
	logger := ctrl.Log.WithName("patch-helper")

	// If the object is already unstructured, we need to perform a deepcopy first
	// because the `DefaultUnstructuredConverter.ToUnstructured` function returns
	// the underlying unstructured object map without making a copy.
	if _, ok := resource.(runtime.Unstructured); ok {
		resource = resource.DeepCopyObject()
	}

	// Convert the resource to unstructured for easier comparison later.
	before, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return nil, err
	}
	logInterface("before", logger, before)

	hasStatus := false
	// attempt to extract the status from the resource for easier comparison later
	beforeStatus, ok, err := unstructured.NestedFieldCopy(before, "status")
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("beforeStatus: %v", beforeStatus))
	if ok {
		hasStatus = true
		// if the resource contains a status remove it from our unstructured copy
		// to avoid unnecessary patching later
		unstructured.RemoveNestedField(before, "status")
		logInterface("before-status-removed", logger, before)
	}

	return &Helper{
		client:        crClient,
		before:        before,
		beforeStatus:  beforeStatus,
		hasStatus:     hasStatus,
		resourcePatch: client.MergeFrom(resource.DeepCopyObject()),
		statusPatch:   client.MergeFrom(resource.DeepCopyObject()),
		log:           logger,
	}, nil
}

// Patch will attempt to patch the given resource and its status
func (h *Helper) Patch(ctx context.Context, resource runtime.Object) error {
	h.log.Info("patching", "object", resource.GetObjectKind())
	if resource == nil {
		return errors.Errorf("expected non-nil resource")
	}

	// If the object is already unstructured, we need to perform a deepcopy first
	// because the `DefaultUnstructuredConverter.ToUnstructured` function returns
	// the underlying unstructured object map without making a copy.
	if _, ok := resource.(runtime.Unstructured); ok {
		resource = resource.DeepCopyObject()
		h.log.Info("deepcopied resource")
	}

	// Convert the resource to unstructured to compare against our before copy.
	after, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	h.log.Info("converted to unstructured resource")
	logInterface("after", h.log, after)

	hasStatus := false
	// attempt to extract the status from the resource to compare against our
	// beforeStatus copy
	afterStatus, ok, err := unstructured.NestedFieldCopy(after, "status")
	if err != nil {
		return err
	}
	h.log.Info("copied status to after status")
	h.log.Info(fmt.Sprintf("afterStatus: %v", afterStatus))
	if ok {
		hasStatus = true
		// if the resource contains a status remove it from our unstructured copy
		// to avoid unnecessary patching.
		unstructured.RemoveNestedField(after, "status")
		h.log.Info("removed status from after status")
		logInterface("after-status-removed", h.log, after)
	}

	var errs []error

	if !reflect.DeepEqual(h.before, after) {
		// only issue a Patch if the before and after resources (minus status) differ
		h.log.Info("applying-patch")
		if err := h.client.Patch(ctx, resource.DeepCopyObject(), h.resourcePatch); err != nil {
			h.log.Error(err, "failed to patch resource")
			errs = append(errs, err)
		}
		h.log.Info("patched resource", "errCount", len(errs))
	}

	if (h.hasStatus || hasStatus) && !reflect.DeepEqual(h.beforeStatus, afterStatus) {
		h.log.Info("patching resource status")
		h.log.Info(fmt.Sprintf("beforeStatus: %v", h.beforeStatus))
		h.log.Info(fmt.Sprintf("afterStatus: %v", afterStatus))
		// only issue a Status Patch if the resource has a status and the beforeStatus
		// and afterStatus copies differ
		if err := h.client.Status().Patch(ctx, resource.DeepCopyObject(), h.statusPatch); err != nil {
			h.log.Error(err, "failed to patch resource status")
			errs = append(errs, err)
		}
	}

	h.log.Info("finished patching", "len(errs)", len(errs))

	return kerrors.NewAggregate(errs)
}

# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

REGISTRY ?= quay.io/ashish_amarnath
IMAGE_NAME ?= executionhook-controller
TAG ?= dev-20191227-02
# Image URL to use all building/pushing image targets
CONTROLLER_IMAGE ?= $(REGISTRY)/$(IMAGE_NAME):$(TAG)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Directories.
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Tool binaries.
KUSTOMIZE := $(TOOLS_BIN_DIR)/kustomize
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); go build -tags=tools -o ./bin/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen

$(KUSTOMIZE): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); go build -tags=tools -o ./bin/kustomize sigs.k8s.io/kustomize/kustomize/v3

.PHONY: all
all: manager

.PHONY: test
test: generate fmt vet manifests # Run tests
	go test ./... -coverprofile cover.out

.PHONY: manager
manager: generate fmt vet # Build manager binary
	go build -o bin/manager main.go

.PHONY: run
run: generate fmt vet manifests # Run against the configured Kubernetes cluster in ~/.kube/config
	go run ./main.go

.PHONY: install
install: manifests $(KUSTOMIZE) # Install CRDs into a cluster
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests $(KUSTOMIZE) # Uninstall CRDs from a cluster
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy
# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
# hacky, works for now. TODO: ashish-amarnath make this better
deploy: manifests $(KUSTOMIZE)
	cd config/manager && ../../$(KUSTOMIZE) edit set image controller=${CONTROLLER_IMAGE}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: manifests
# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

.PHONY: docker-build
docker-build: test # Build the controller image
	docker build . -t ${CONTROLLER_IMAGE}

.PHONY: docker-push
docker-push: docker-build # Push the controller image
	docker push ${CONTROLLER_IMAGE}


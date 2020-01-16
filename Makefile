# Copyright 2020 The Kubernetes Authors.
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

.DEFAULT_GOAL:=help

REGISTRY ?= gcr.io/$(shell gcloud config get-value project)
IMAGE_NAME ?= executionhook-controller
TAG ?= dev
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

.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean-bin
clean-bin: ## Remove all generated binaries
	rm -rf bin
	rm -rf hack/tools/bin

.PHONY: all
all: manager

.PHONY: test
test: generate fmt vet manifests ## Run tests
	go test -v ./... -coverprofile cover.out

.PHONY: manager
manager: generate fmt vet test ## Build manager binary
	go build -o bin/manager main.go

.PHONY: install
install: manifests $(KUSTOMIZE) ## Install CRDs into a cluster in the current context at ~/.kube/config
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests $(KUSTOMIZE) ## Uninstall latest version of CRDs from a cluster in the current context at ~/.kube/config
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy
# hacky, works for now. TODO: ashish-amarnath make this better
deploy: manifests $(KUSTOMIZE) ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && ../../$(KUSTOMIZE) edit set image controller=${CONTROLLER_IMAGE}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: manifests
manifests: $(CONTROLLER_GEN) ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

fmt: ## Run go fmt against code
	go fmt ./...

vet: ## Run go vet against code
	go vet ./...

.PHONY: modules
modules: ## Runs go mod to ensure modules are up-to-date.
	go mod tidy
	cd $(TOOLS_DIR); go mod tidy

.PHONY: verify-modules
verify-modules: modules
	@if !(git diff --quiet HEAD -- go.sum go.mod hack/tools/go.mod hack/tools/go.sum); then \
		echo "go module files are out of date"; exit 1; \
	fi

generate: $(CONTROLLER_GEN) ## Generate code
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."


.PHONY: verify-gen
verify-gen: generate manifests
	@if !(git diff --quiet HEAD); then \
		echo "generated code and manifest files are out of date, run make generate manifests"; exit 1; \
	fi

.PHONY: docker-build
docker-build: test # Build the controller image
	docker build . -t ${CONTROLLER_IMAGE}

.PHONY: docker-push
docker-push: docker-build # Push the controller image
	docker push ${CONTROLLER_IMAGE}


# Image URL to use all building/pushing image targets
IMG ?= k8s-controller-sample:latest

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.33.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: ## Generate RBAC manifests.
	@echo "RBAC manifests are in the config/ directory"

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	@echo "No code generation needed for this controller"

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: ## Run golangci-lint against code.
	golangci-lint run

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: fmt vet ## Run the controller from your host.
	go run ./main.go manager

.PHONY: run-leader-elect
run-leader-elect: fmt vet ## Run the controller with leader election from your host.
	go run ./main.go manager --leader-elect

.PHONY: run-server
run-server: fmt vet ## Run the HTTP server from your host.
	go run ./main.go server

.PHONY: run-controller
run-controller: fmt vet ## Run the basic controller from your host.
	go run ./main.go controller

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have a buildx builder with platform support. More info: https://docs.docker.com/build/buildx/create/#usage
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: ## Install RBAC into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/rbac.yaml

.PHONY: uninstall
uninstall: ## Uninstall RBAC from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/rbac.yaml --ignore-not-found=$(ignore-not-found)

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/rbac.yaml
	kubectl apply -f config/deployment.yaml

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/deployment.yaml --ignore-not-found=$(ignore-not-found)
	kubectl delete -f config/rbac.yaml --ignore-not-found=$(ignore-not-found)

.PHONY: logs
logs: ## Show logs from the deployed controller.
	kubectl logs -l app=k8s-controller-sample -f

.PHONY: status
status: ## Show status of the deployed controller.
	kubectl get pods -l app=k8s-controller-sample
	kubectl get leases -l app.kubernetes.io/name=k8s-controller-sample

##@ Testing

.PHONY: test-leader-election
test-leader-election: ## Test leader election by scaling deployment.
	kubectl scale deployment k8s-controller-sample --replicas=3
	@echo "Watch the logs to see leader election in action:"
	@echo "kubectl logs -l app=k8s-controller-sample -f"

.PHONY: test-metrics
test-metrics: ## Test metrics endpoint.
	kubectl port-forward svc/k8s-controller-sample-metrics 8080:8080 &
	@echo "Metrics available at: http://localhost:8080/metrics"
	@echo "Health check at: http://localhost:8081/healthz"

##@ Cleanup

.PHONY: clean
clean: ## Clean up build artifacts.
	rm -rf bin/
	rm -f cover.out 
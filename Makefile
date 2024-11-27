VERSION ?= 0.0.33

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25.0

# kustomize for deploy
KUSTOMIZE = go run sigs.k8s.io/kustomize/kustomize/v3

IMAGE_TAG_BASE ?= ghcr.io/ionos-cloud/octopinger/operator

IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)

##@ Development

generate:
	@go generate ./...
	@go run cmd/manifest/manifest.go --file manifests/crd/bases/octopinger.io_octopingers.yaml \
		--file manifests/install/service_account.yaml \
		--file manifests/install/cluster_role.yaml \
		--file manifests/install/cluster_role_binding.yaml \
		--file manifests/install/statefulset.yaml \
		--output manifests/install.yaml

docker-minikube:
	minikube -p octopinger image rm ${IMG}
	minikube -p octopinger image rm ghcr.io/ionos-cloud/octopinger/octopinger:v$(VERSION)

	@docker build --build-arg BINARY=octopinger-linux-amd64 -f Dockerfile -t ghcr.io/ionos-cloud/octopinger/octopinger:v$(VERSION) ./dist
	@docker build --build-arg BINARY=operator-linux-amd64 -f Dockerfile.nonroot -t ${IMG} ./dist

	minikube -p octopinger image load ${IMG}
	minikube -p octopinger image load ghcr.io/ionos-cloud/octopinger/octopinger:v$(VERSION)

##@ Build

.PHONY: fmt
fmt: ## Run go fmt against code.
	go run mvdan.cc/gofumpt -w .

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: build
build:
	@goreleaser build --rm-dist --snapshot

.PHONY: vendor
vendor: export GOPRIVATE=github.com/ionos-cloud
vendor:
	@go mod tidy
	@go mod vendor
	@go get -u ./...

##@ Deployment

deploy: generate ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd manifests/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build manifests/default | kubectl apply -f -

remove: generate ## Remove controller to the K8s cluster specified in ~/.kube/config.
	cd manifests/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build manifests/default | kubectl delete -f -

##@ Test

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ENVTEST ?= $(LOCALBIN)/setup-envtest

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@release-0.17

.PHONY: test
test: envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out
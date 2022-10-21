# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Controller generation command
CONTROLLER_GEN = go run sigs.k8s.io/controller-tools/cmd/controller-gen

# kustomize for deploy
KUSTOMIZE = go run sigs.k8s.io/kustomize/kustomize/v3

IMAGE_TAG_BASE ?= ghcr.io/ionos-cloud/octopinger/operator

IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)

##@ Development

manifests:
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate:
	$(CONTROLLER_GEN) object:headerFile="hack/copyright.go.txt" paths="./..."

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

.PHONY: test
test:
	@go test -race ./... -count=1 -cover -coverprofile cover.out

.PHONY: vendor
vendor: export GOPRIVATE=github.com/ionos-cloud
vendor:
	@go mod tidy
	@go mod vendor
	@go get -u ./...

##@ Deployment
deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -
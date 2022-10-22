VERSION ?= 0.0.3

# kustomize for deploy
KUSTOMIZE = go run sigs.k8s.io/kustomize/kustomize/v3

IMAGE_TAG_BASE ?= ghcr.io/ionos-cloud/octopinger/manager

IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)

##@ Development

generate:
	@go generate ./...

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

deploy: generate ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

remove: generate ## Remove controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl delete -f -
VERSION ?= 0.0.24

# kustomize for deploy
KUSTOMIZE = go run sigs.k8s.io/kustomize/kustomize/v3

IMAGE_TAG_BASE ?= ghcr.io/ionos-cloud/octopinger/operator

IMG ?= $(IMAGE_TAG_BASE):v$(VERSION)

##@ Development

generate:
	@go generate ./...

docker-minikube:
	minikube -p octopinger image rm ${IMG}
	minikube -p octopinger image rm ghcr.io/ionos-cloud/octopinger/octopinger:latest

	@docker build --build-arg BINARY=octopinger-linux-amd64 -f Dockerfile -t ghcr.io/ionos-cloud/octopinger/octopinger:latest ./dist
	@docker build --build-arg BINARY=operator-linux-amd64 -f Dockerfile.nonroot -t ${IMG} ./dist

	minikube -p octopinger image load ${IMG}
	minikube -p octopinger image load ghcr.io/ionos-cloud/octopinger/octopinger:latest

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
	cd manifests/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build manifests/default | kubectl apply -f -

remove: generate ## Remove controller to the K8s cluster specified in ~/.kube/config.
	cd manifests/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build manifests/default | kubectl delete -f -
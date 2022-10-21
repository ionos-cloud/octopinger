.PHONY: generate
generate:
	go generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	go run mvdan.cc/gofumpt -w .

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: build
build:
	@goreleaser build --rm-dist --snapshot --single-target

.PHONY: test
test:
	@go test -race ./... -count=1 -cover -coverprofile cover.out

.PHONY: stop-scylla
stop-scylla:
	@docker stop scylla

.PHONY: vendor
vendor: export GOPRIVATE=github.com/ionos-cloud
vendor:
	@go mod tidy
	@go mod vendor
	@go get -u ./...

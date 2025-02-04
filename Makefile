GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}

ifeq ($(GOOS),windows)
OUTPUT_PATH = ${BUILD_DIR}/baton-pingfederate.exe
else
OUTPUT_PATH = ${BUILD_DIR}/baton-pingfederate
endif

.PHONY: build
build:
	go build -o ${OUTPUT_PATH} ./cmd/baton-pingfederate

.PHONY: update-deps
update-deps:
	go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-dep
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix ./...

.PHONY: release
release:
	goreleaser release --rm-dist

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --rm-dist

.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest

.PHONY: ci
ci: install-tools lint build

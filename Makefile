
# go tool nm ./luet | grep Commit
override LDFLAGS += -X "github.com/macaroni-os/macaronictl/cmd.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
override LDFLAGS += -X "github.com/macaroni-os/macaronictl/cmd.BuildCommit=$(shell git rev-parse HEAD)"

NAME ?= macaronictl
PACKAGE_NAME ?= $(NAME)
REVISION := $(shell git rev-parse --short HEAD || echo dev)
VERSION := $(shell git describe --tags || echo $(REVISION))
VERSION := $(shell echo $(VERSION) | sed -e 's/^v//g')
BUILD_PLATFORMS ?= -osarch="linux/amd64" -osarch="linux/386" -osarch="linux/arm"
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

GOBIN := $(shell go env GOPATH)/bin/
GINKGO := $(GOBIN)/ginkgo

.PHONY: all
all: deps build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test: deps
	$(GINKGO) -r -flake-attempts 3 ./...

.PHONY: coverage
coverage:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: test-coverage
test-coverage:
	scripts/ginkgo.coverage.sh --codecov

.PHONY: clean
clean:
	rm -rf release/

.PHONY: deps
deps:
	go env
	# Installing dependencies...
	GOBIN=$(GOBIN) go install github.com/onsi/ginkgo/v2/ginkgo
	$(GINKGO) version

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)'

.PHONY: build-small
build-small:
	@$(MAKE) LDFLAGS+="-s -w" build
	upx --brute -1 $(NAME)

.PHONY: lint
lint:
	golint ./... | grep -v "be unexported"

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: goreleaser-snapshot
goreleaser-snapshot:
	rm -rf dist/ || true
	goreleaser release --skip=validate,publish --snapshot --verbose

SHELL=/bin/bash
GOPATH:=$(shell go env GOPATH | tr '\\' '/')
GOEXE:=$(shell go env GOEXE)
GORELEASER:=$(GOPATH)/bin/goreleaser$(GOEXE)
HOSTNAME=registry.terraform.io
NAMESPACE=rgl
NAME=saml
BINARY=terraform-provider-${NAME}
VERSION?=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: install

$(GORELEASER):
	go install github.com/goreleaser/goreleaser@v1.15.2

release-snapshot: $(GORELEASER)
	$(GORELEASER) release --snapshot --skip-publish --skip-sign --clean

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall:
	rm -f .terraform.lock.hcl
	rm -rf .terraform/providers/${HOSTNAME}/${NAMESPACE}/${NAME}
	rm -rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}

.PHONY: default build release-snapshot install uninstall

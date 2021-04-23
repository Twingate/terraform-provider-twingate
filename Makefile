TEST?=$$(go list ./...)
WEBSITE_REPO=github.com/hashicorp/terraform-website
HOSTNAME=registry.terraform.io
PKG_NAME=twingate
BINARY=terraform-provider-${PKG_NAME}
VERSION=0.1
OS_ARCH=darwin_amd64
GOBINPATH=$(shell go env GOPATH)/bin

default: build

.PHONY: docs

vendor:
	go mod vendor

build: vendor fmtcheck
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/twingate/${PKG_NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/twingate/${PKG_NAME}/${VERSION}/${OS_ARCH}

test:
	./scripts/test.sh

testacc:
	TF_ACC=1 TF_SCHEMA_PANIC_ON_ERROR=1 ./scripts/test.sh

fmtcheck:
	@echo "==> Checking source code against gofmt..."
	@sh -c $(CURDIR)/scripts/gofmtcheck.sh

lint: tools
	@echo "==> Checking source code against linters..."
	@$(GOBINPATH)/golangci-lint run -c golangci.yml ./$(PKG_NAME)

lint-fix: tools
	@echo "==> Checking source code against linters with fix enabled..."
	@$(GOBINPATH)/golangci-lint run --fix -c golangci.yml ./$(PKG_NAME)

sec: tools
	@echo "==> Checking source code against security issues..."
	@$(GOBINPATH)/gosec ./$(PKG_NAME)

docs: tools
	tfplugindocs

tools:
	@echo "==> installing required tools ..."
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install gotest.tools/gotestsum@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

docscheck:
	@sh -c "'$(CURDIR)/scripts/docscheck.sh'"

.PHONY: build test testacc vet fmt fmtcheck lint tools errcheck test-compile website website-test docscheck

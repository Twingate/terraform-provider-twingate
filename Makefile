TEST?=$$(go list ./...)
WEBSITE_REPO=github.com/hashicorp/terraform-website
HOSTNAME=registry.terraform.io
PKG_NAME=twingate
BINARY=terraform-provider-${PKG_NAME}
VERSION=0.1
OS_ARCH=darwin_amd64
GOBINPATH=$(shell go env GOPATH)/bin

default: build

.PHONY: ci-checks
ci-checks: docs
	echo "Checking if latest docs generated"
	git diff --exit-code || echo "ERROR: Update and push the latest documentation"; exit 1

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: build
build: vendor fmtcheck
	go build -o ${BINARY}

.PHONY: build-release
build-release:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -a -o build/terraform-provider-twingate-darwin-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-s -w" -a -o build/terraform-provider-twingate-linux-386
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -a -o build/terraform-provider-twingate-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -a -o build/terraform-provider-twingate-linux-arm
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags="-s -w" -a -o build/terraform-provider-twingate-windows-386.exe
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -a -o build/terraform-provider-twingate-windows-amd64.exe

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/twingate/${PKG_NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/twingate/${PKG_NAME}/${VERSION}/${OS_ARCH}

.PHONY: test
test:
	./scripts/test.sh

.PHONY: testacc
testacc:
	TF_ACC=1 TF_SCHEMA_PANIC_ON_ERROR=1 ./scripts/test.sh

.PHONY: fmt
fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -w -s ./$(PKG_NAME)

.PHONY: fmtcheck
fmtcheck:
	@echo "==> Checking source code against gofmt..."
	@sh -c $(CURDIR)/scripts/gofmtcheck.sh

.PHONY: lint
lint: tools
	@echo "==> Checking source code against linters..."
	@$(GOBINPATH)/golangci-lint run -c golangci.yml ./$(PKG_NAME)

.PHONY: lint-fix
lint-fix: tools
	@echo "==> Checking source code against linters with fix enabled..."
	@$(GOBINPATH)/golangci-lint run --fix -c golangci.yml ./$(PKG_NAME)

.PHONY: sec
sec: tools
	@echo "==> Checking source code against security issues..."
	@$(GOBINPATH)/gosec ./$(PKG_NAME)


.PHONY: doc-tools
docs: doc-tools
	@$(GOBINPATH)/tfplugindocs generate

.PHONY: tools
tools:
	@echo "==> installing required tools ..."
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install gotest.tools/gotestsum@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

.PHONY: doc-tools
doc-tools:
	@echo "==> installing required doc tools ..."
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

.PHONY: test-compile
test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

.PHONY: website
website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: website-test
website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

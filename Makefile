TEST?=$$(go list ./...)
WEBSITE_REPO=github.com/hashicorp/terraform-website
HOSTNAME=registry.terraform.io
PKG_NAME=twingate
BINARY=terraform-provider-${PKG_NAME}
VERSION=0.1
OS_ARCH=darwin_amd64

default: build

.PHONY: docs

build: fmtcheck
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/twingate/${PKG_NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/twingate/${PKG_NAME}/${VERSION}/${OS_ARCH}

test: fmtcheck
	./scripts/test.sh

testacc: fmtcheck
	TF_ACC=1 TF_SCHEMA_PANIC_ON_ERROR=1 go test $(TEST) -v $(TESTARGS) -timeout 10m

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -w -s ./$(PKG_NAME)

# Currently required by tf-deploy compile
fmtcheck:
	@echo "==> Checking source code against gofmt..."
	@sh -c $(CURDIR)/scripts/gofmtcheck.sh

lint: tools
	@echo "==> Checking source code against linters..."
	@golangci-lint run ./$(PKG_NAME)

docs: tools
	tfplugindocs generate

tools:
	@echo "==> installing required toolilintng..."
	go install github.com/client9/misspell/cmd/misspell
	go install github.com/golangci/golangci-lint/cmd/golangci-lint

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

.PHONY: build test testacc vet fmt fmtcheck lint tools errcheck test-compile website website-test docscheck generate

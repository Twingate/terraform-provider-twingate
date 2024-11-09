TEST?=$$(go list ./...)
HOSTNAME=registry.terraform.io
PKG_NAME=twingate
BINARY=terraform-provider-${PKG_NAME}
VERSION=0.1
OS_ARCH=darwin_amd64
GOBINPATH=$(shell go env GOPATH)/bin
SWEEP_TENANT=terraformtests
SWEEP_FOLDER=./twingate/internal/test/sweepers
GOLINT_VERSION=v1.61.0


check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

.PHONY: sweep
sweep:
	$(call check_defined, TWINGATE_NETWORK)
	$(call check_defined, TWINGATE_API_TOKEN)
	$(call check_defined, TWINGATE_URL)
	go test ${SWEEP_FOLDER} -v -sweep=${SWEEP_TENANT} -timeout 60m

default: build

.PHONY: ci-checks
ci-checks: docs
	echo "Checking if latest docs generated"
	git diff --exit-code || echo "ERROR: Update and push the latest documentation"; exit 1


.PHONY: build
build: fmtcheck
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
lint:
	@echo "==> Checking source code against linters..."
	docker run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:$(GOLINT_VERSION) golangci-lint run -c /app/golangci.yml /app/$(PKG_NAME)/...


.PHONY: lint-fix
lint-fix:
	@echo "==> Checking source code against linters with fix enabled..."
	docker run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint:$(GOLINT_VERSION) golangci-lint run --fix -c /app/golangci.yml /app/$(PKG_NAME)/...

.PHONY: sec
sec:
	@echo "==> Checking source code against security issues..."
	go run github.com/securego/gosec/v2/cmd/gosec ./$(PKG_NAME)/...


.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

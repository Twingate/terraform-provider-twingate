TEST?=$$(go list . | grep -v 'vendor')
NAME=twingate
BINARY=terraform-provider-${NAME}
OS_ARCH=darwin_amd64
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: install

build:
	go build -o ${BINARY}

test-deps: ## Install test dependencies
	GO111MODULE=off go get gotest.tools/gotestsum

test: test-deps  ## Run tests
	@echo "running all tests for all packages"
	ops/test.sh .

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -w $(GOFMT_FILES)

lint:
	./ops/golintsec.sh

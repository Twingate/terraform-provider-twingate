TEST?=$$(go list . | grep -v 'vendor')
NAME=twingate
BINARY=terraform-provider-${NAME}
OS_ARCH=darwin_amd64
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: install

build:
	go build -o ${BINARY}

test:
	go test $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -w $(GOFMT_FILES)

lint:
	scripts/golintsec.sh

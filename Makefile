
BUILD := gomtree
CWD := $(shell pwd)
SOURCE_FILES := $(shell find . -type f -name "*.go")

default: build validation 

.PHONY: validation
validation: .test .lint .vet .cli.test

.PHONY: test
test: .test

.test: $(SOURCE_FILES)
	go test -v ./... && touch $@

.PHONY: lint
lint: .lint

.lint: $(SOURCE_FILES)
	golint -set_exit_status ./... && touch $@

.PHONY: vet
vet: .vet

.vet: $(SOURCE_FILES)
	go vet ./... && touch $@

.PHONY: cli.test
cli.test: .cli.test

.cli.test: $(BUILD) $(wildcard ./test/cli/*.sh)
	@ for test in ./test/cli/*.sh ; do \
	bash $$test $(CWD) ; \
	done && touch $@

.PHONY: build
build: $(BUILD)

$(BUILD): $(SOURCE_FILES)
	go build ./cmd/$(BUILD)

clean:
	rm -rf $(BUILD) .test .vet .lint .cli.test


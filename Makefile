
BUILD := gomtree
CWD := $(shell pwd)
SOURCE_FILES := $(shell find . -type f -name "*.go")
CLEAN_FILES := *~

default: build validation 

.PHONY: validation
validation: test .lint .vet .cli.test

.PHONY: test
test: .test .test.tags

CLEAN_FILES += .test .test.tags

.test: $(SOURCE_FILES)
	go test -v ./... && touch $@

.test.tags: $(SOURCE_FILES)
	go test -tags govis -v ./... && touch $@

.PHONY: lint
lint: .lint

CLEAN_FILES += .lint

.lint: $(SOURCE_FILES)
	golint -set_exit_status ./... && touch $@

.PHONY: vet
vet: .vet

CLEAN_FILES += .vet

.vet: $(SOURCE_FILES)
	go vet ./... && touch $@

.PHONY: cli.test
cli.test: .cli.test

CLEAN_FILES += .cli.test

.cli.test: $(BUILD) $(wildcard ./test/cli/*.sh)
	@go run ./test/cli.go ./test/cli/*.sh && touch $@

.PHONY: build
build: $(BUILD)

$(BUILD): $(SOURCE_FILES)
	go build ./cmd/$(BUILD)

clean:
	rm -rf $(BUILD) $(CLEAN_FILES)



BUILD := gomtree
CWD := $(shell pwd)
SOURCE_FILES := $(shell find . -type f -name "*.go")
CLEAN_FILES := *~
TAGS := cvis

default: build validation 

.PHONY: validation
validation: .test .lint .vet .cli.test

.PHONY: validation.tags
validation.tags: .test.tags .vet.tags .cli.test

.PHONY: test
test: .test

CLEAN_FILES += .test .test.tags

.test: $(SOURCE_FILES)
	go test -v ./... && touch $@

.test.tags: $(SOURCE_FILES)
	set -e ; for tag in $(TAGS) ; do go test -tags $$tag -v ./... ; done && touch $@

.PHONY: lint
lint: .lint

CLEAN_FILES += .lint

.lint: $(SOURCE_FILES)
	golint -set_exit_status ./... && touch $@

.PHONY: vet
vet: .vet .vet.tags

CLEAN_FILES += .vet .vet.tags

.vet: $(SOURCE_FILES)
	go vet ./... && touch $@

.vet.tags: $(SOURCE_FILES)
	set -e ; for tag in $(TAGS) ; do go vet -tags $$tag -v ./... ; done && touch $@

.PHONY: cli.test
cli.test: .cli.test

CLEAN_FILES += .cli.test .cli.test.tags

.cli.test: $(BUILD) $(wildcard ./test/cli/*.sh)
	@go run ./test/cli.go ./test/cli/*.sh && touch $@

.cli.test.tags: $(BUILD) $(wildcard ./test/cli/*.sh)
	@set -e ; for tag in $(TAGS) ; do go run -tags $$tag ./test/cli.go ./test/cli/*.sh ; done && touch $@

.PHONY: build
build: $(BUILD)

$(BUILD): $(SOURCE_FILES)
	go build ./cmd/$(BUILD)

clean:
	rm -rf $(BUILD) $(CLEAN_FILES)


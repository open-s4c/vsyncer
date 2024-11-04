PREFIX     ?= /usr/local
BUILD       = build
TAG        ?= $(shell git describe --always --tags --dirty)
DOCKER_TAG ?= latest
USE_DOCKER ?= true
LDXFLAGS    = -X main.version=$(TAG) \
              -X vsync/tools.useDocker=$(USE_DOCKER) \
              -X vsync/tools.dockerTag=$(DOCKER_TAG)
LDXFLAGS_L   = -X main.version=$(TAG) \
              -X vsync/tools.useDocker=$(USE_DOCKER) \
              -X vsync/tools.dockerTag=latest

LDFLAGS     = -ldflags='-extldflags=-static $(LDXFLAGS)'
LDFLAGS_L   = -ldflags='-extldflags=-static $(LDXFLAGS_L)'

SOURCES     = $(shell find -name '*.go' -not -name '*_string.go')
GENERATED   = $(shell find -name *_string.go)
################################################################################
# user goals
################################################################################
.PHONY: all build install clean help

help:
	@echo "Help:"
	@echo "- make generate             run go generate"
	@echo "- make build                build vsyncer in $(BUILD)/"
	@echo "- make all                  generate and build"
	@echo "- make install              copy vsyncer into /usr/local/bin"
	@echo "- make install PREFIX=path  copy vsyncer into \$$PREFIX/bin"
	@echo "- make clean                delete $(BUILD)"
	@echo "- make test                 run unit tests"

all: build

build: generate $(BUILD)/vsyncer $(BUILD)/vsyncer-latest

install: $(BUILD)/vsyncer $(BUILD)/vsyncer-latest
	install $(BUILD)/vsyncer $(PREFIX)/bin/
	install $(BUILD)/vsyncer-latest $(PREFIX)/bin/
clean:
	rm -rf $(BUILD) $(GENERATED)

################################################################################
# build goals
################################################################################
.PHONY: build-dir generate

build-dir:
	@mkdir -p $(BUILD)

$(BUILD)/vsyncer: build-dir $(SOURCES)
	env CGO_ENABLED=0 go build $(LDFLAGS) -o $@ ./cmd/vsyncer

$(BUILD)/vsyncer-latest: build-dir $(SOURCES)
	env CGO_ENABLED=0 go build $(LDFLAGS_L) -o $@ ./cmd/vsyncer
################################################################################
# support goals
################################################################################
.PHONY: generate lint test fmt-c forbidigo revive

generate:
	go generate ./...

lint: forbidigo

forbidigo:
	@go run github.com/ashanbrown/forbidigo \
		-set_exit_status os.LookupEnv os.Getenv -- ./...

revive:
	revive -exclude vendor/... ./...

test:
	go test ./...

fmt-c:
	find . -name '*.c' -o -name '*.h' \
        	-exec clang-format -style WebKit -i {} +

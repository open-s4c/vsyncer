PREFIX ?= /usr/local
TAG     	= $(shell git describe --always --tags --dirty)
FLAGS   	= -ldflags '-X main.version=$(TAG)'
BUILD   	= build
GENERATED 	= $(shell find -name *_string.go)
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

build: generate build/vsyncer

install: build
	install $(BUILD)/vsyncer $(PREFIX)/bin/
clean:
	rm -rf $(BUILD) $(GENERATED)

################################################################################
# build goals
################################################################################
.PHONY: build-dir generate

build-dir:
	mkdir -p $(BUILD)

$(BUILD)/vsyncer: build-dir $(shell find -name '*.go')
	go build $(FLAGS) -o $@ ./cmd/vsyncer

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

all: generate vsyncer

TAG   = $(shell git describe --always --tags --dirty)
FLAGS = -ldflags '-X main.version=$(TAG)'

vsyncer: $(shell find -name '*.go')
	go build $(FLAGS) -o $@ ./cmd/$@

PKGS = core module checker optimizer cmd/vsyncer

build:
	mkdir -p build

TESTERS = $(addprefix build/tester-,$(subst /,-,$(PKGS)))
build-testers: $(TESTERS)

build/tester-%: build
	go test -c -o $@ ./$(subst -,/,$*)

generate:
	go generate ./...

lint:
	revive -exclude vendor/... ./...

test:
	go test -cover ./...

cover:
	go test -coverpkg=./... -coverprofile=cover.out ./...
	go tool cover -func=cover.out

fmt-c:
	find . -name '*.c' -o -name '*.h' -exec clang-format -style WebKit -i {} +

clean:
	rm -rf vsyncer $(TESTERS)

.PHONY: all clean test lint cover generate build-testers


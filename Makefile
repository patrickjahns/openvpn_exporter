SHELL := bash
NAME := openvpn_exporter
IMPORT := github.com/patrickjahns/$(NAME)
BIN := bin
DIST := dist
GO := go

ifeq ($(OS), Windows_NT)
	EXECUTABLE := $(NAME).exe
else
	EXECUTABLE := $(NAME)
endif

PACKAGES ?= $(shell go list ./...)
SOURCES ?= $(shell find . -name "*.go" -type f)
GENERATE ?= $(PACKAGES)

ifndef DATE
	DATE := $(shell date -u '+%Y%m%d')
endif

ifndef VERSION
	VERSION ?= $(shell git rev-parse --short HEAD)
endif

ifndef REVISION
	REVISION ?= $(shell git rev-parse --short HEAD)
endif

LDFLAGS += -s -w
LDFLAGS += -X "$(IMPORT)/pkg/version.Version=$(VERSION)"
LDFLAGS += -X "$(IMPORT)/pkg/version.BuildDate=$(DATE)"
LDFLAGS += -X "$(IMPORT)/pkg/version.Revision=$(REVISION)"

.PHONY: all
all: build

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -rf $(BIN)/
	rm -rf $(DIST)/

.PHONY: sync
sync:
	$(GO) mod download

.PHONY: fmt
fmt:
	$(GO) fmt $(PACKAGES)

.PHONY: vet
vet:
	$(GO) vet $(PACKAGES)

.PHONY: generate
generate:
	$(GO) generate $(GENERATE)

.PHONY: lint
lint:
	@which golangci-lint > /dev/null; if [ $$? -ne 0 ]; then \
		(echo "please install golangci-lint"; exit 1) \
	fi
	golangci-lint run -v

.PHONY: test
test:
	@which goverage > /dev/null; if [ $$? -ne 0 ]; then \
		GO111MODULE=off $(GO) get -u github.com/haya14busa/goverage; \
	fi
	goverage -v -coverprofile coverage.out $(PACKAGES)

.PHONY: build
build: $(BIN)/$(EXECUTABLE)

$(BIN)/$(EXECUTABLE): $(SOURCES)
	$(GO) build -i -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

.PHONY: release
release: release-dirs release-build release-checksums

.PHONY: release-dirs
release-dirs:
	mkdir -p $(DIST)

.PHONY: release-build
release-build:
	@which gox > /dev/null; if [ $$? -ne 0 ]; then \
		GO111MODULE=off  $(GO) get -u github.com/mitchellh/gox; \
	fi
	gox -arch="386 amd64 arm" -osarch '!darwin/386' -verbose -ldflags '-w $(LDFLAGS)' -output="$(DIST)/$(EXECUTABLE)-{{.OS}}-{{.Arch}}" ./cmd/$(NAME)

.PHONY: release-checksums
release-checksums:
	cd $(DIST); $(foreach file, $(wildcard $(DIST)/$(EXECUTABLE)-*), sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

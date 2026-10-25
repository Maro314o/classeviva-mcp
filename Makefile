GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w

ifeq ($(OS),Windows_NT)
OUTPUT := bin/classeviva.exe
else
OUTPUT := bin/classeviva
endif

export CGO_ENABLED := 0

.PHONY: build

build:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "$(OUTPUT)" ./cmd/classeviva

GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w

ifeq ($(OS),Windows_NT)
OUTPUT := bin/classeviva.exe
else
OUTPUT := bin/classeviva
endif

export CGO_ENABLED := 0

.PHONY: build install-local-plugin

build:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "$(OUTPUT)" ./cmd/classeviva

# Refresh the personal Codex installation from this checkout after local changes.
install-local-plugin: build
	@if codex plugin list --marketplace personal --json 2>/dev/null | python3 -c 'import json,sys; raw=sys.stdin.read(); data=json.loads(raw[raw.index("{"):]); print(str(any(p.get("pluginId") == "classeviva@personal" and p.get("installed") for p in data.get("installed", []))).lower())' | grep -qx true; then \
		codex plugin remove classeviva@personal --json; \
	fi
	codex plugin add classeviva@personal --json

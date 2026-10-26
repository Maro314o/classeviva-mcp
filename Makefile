GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w

ifeq ($(OS),Windows_NT)
OUTPUT := bin/classeviva.exe
else
OUTPUT := bin/classeviva
endif

export CGO_ENABLED := 0

.PHONY: build install-local-mcp install-local-plugin

build:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "$(OUTPUT)" ./cmd/classeviva

# Register the local MCP server in the user config so every local Codex task
# can start it with an absolute path, independent of the task's working folder.
install-local-mcp: build
	@codex mcp remove classeviva --json >/dev/null 2>&1 || true
	codex mcp add classeviva -- "$(CURDIR)/$(OUTPUT)"

# Refresh the local plugin and MCP installation from this checkout.
install-local-plugin: build
	@codex plugin remove classeviva@personal --json >/dev/null 2>&1 || true
	@codex plugin remove classeviva@classeviva-repo --json >/dev/null 2>&1 || true
	@codex plugin marketplace remove classeviva-repo --json >/dev/null 2>&1 || true
	codex plugin marketplace add "$(CURDIR)" --json
	codex plugin add classeviva@classeviva-repo --json
	@codex mcp remove classeviva --json >/dev/null 2>&1 || true
	codex mcp add classeviva -- "$(CURDIR)/$(OUTPUT)"

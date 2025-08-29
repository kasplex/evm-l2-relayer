include version.mk

GO_BASE := $(shell pwd)
GO_BIN := $(GO_BASE)/dist
GO_ENV_VARS := GO_BIN=$(GO_BIN)
GO_BINARY := relayer
GO_CMD := $(GO_BASE)/cmd/relayer

LDFLAGS += -X 'github.com/kasplex-evm/kasplex-relayer.Version=$(VERSION)'
LDFLAGS += -X 'github.com/kasplex-evm/kasplex-relayer.GitRev=$(GITREV)'
LDFLAGS += -X 'github.com/kasplex-evm/kasplex-relayer.GitBranch=$(GITBRANCH)'
LDFLAGS += -X 'github.com/kasplex-evm/kasplex-relayer.BuildDate=$(DATE)'

BUILD := $(GO_ENV_VARS) go build -ldflags "all=$(LDFLAGS)" -o $(GO_BIN)/$(GO_BINARY) $(GO_CMD)

.PHONY: build
build: ## Build the binary locally into ./dist
	$(BUILD)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
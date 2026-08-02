.DEFAULT_GOAL := help

GOFUMPT ?= gofumpt
CDK8SPLUS_VERSIONS := 34 35 36

.PHONY: help
help: ## Show available commands
	@cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: format
format: ## Format source code with gofumpt
	$(GOFUMPT) -w .

.PHONY: format-check
format-check: ## Check if source code is gofumpt-formatted
	@files="$$($(GOFUMPT) -l .)"; \
	if [ -n "$$files" ]; then \
		echo 'Go source is not formatted with gofumpt:' >&2; \
		echo "$$files" >&2; \
		exit 1; \
	fi

.PHONY: build
build: ## Compile every Go package
	go build ./...

.PHONY: build-cli
build-cli: ## Build bin/purecdk8s
	mkdir -p bin
	go build -o bin/purecdk8s ./cmd/purecdk8s

.PHONY: unittest
unittest: ## Run Go unit tests
	go test ./...

.PHONY: api
api: api-constructs api-cdkplus ## Run every API compatibility check

.PHONY: api-constructs
api-constructs: ## Check constructs API compatibility
	./scripts/check-constructs-api.sh

.PHONY: api-cdkplus
api-cdkplus: ## Check cdk8s+ API compatibility
	@for version in $(CDK8SPLUS_VERSIONS); do \
		./scripts/check-cdk8splus-api.sh "$$version"; \
	done

.PHONY: integration
integration: ## Run every integration test with Docker
	./integration/run-all.sh

.PHONY: integration-example
integration-example: ## Run one example (EXAMPLE=name, optional MODE=upstream|pure|both)
	@test -n "$(EXAMPLE)" || { echo 'EXAMPLE is required, for example: make integration-example EXAMPLE=helm' >&2; exit 2; }
	./integration/run.sh "$(EXAMPLE)" "$(MODE)"

.PHONY: test
test: format-check build unittest api integration ## Run every local check

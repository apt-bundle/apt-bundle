IMAGE_NAME ?= with-lockfile
IMAGE_TAG ?= latest
DOCKERFILE ?= Dockerfile

.DEFAULT_GOAL := help

.PHONY: help build run test shell clean

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Variables:"
	@echo "  IMAGE_NAME  Docker image name (default: $(IMAGE_NAME))"
	@echo "  IMAGE_TAG   Docker image tag (default: $(IMAGE_TAG))"
	@echo "  DOCKERFILE  Dockerfile to use (default: $(DOCKERFILE))"
	@echo ""

build: ## Build the Docker image
	docker build -f $(DOCKERFILE) -t $(IMAGE_NAME):$(IMAGE_TAG) .

run: ## Run the container interactively
	docker run -it $(IMAGE_NAME):$(IMAGE_TAG)

test: ## Test that packages were installed at the locked versions
	docker run --rm $(IMAGE_NAME):$(IMAGE_TAG) bash -c "curl --version && echo '---' && git --version && echo '---' && jq --version"
	@echo "Verify that curl, git, and jq versions are displayed above (separated by ---)."

shell: ## Open an interactive shell in the container
	docker run -it $(IMAGE_NAME):$(IMAGE_TAG) /bin/bash

clean: ## Remove the Docker image
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) || true

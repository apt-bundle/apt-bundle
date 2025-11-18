IMAGE_NAME ?= complex-via-apt-get
IMAGE_TAG ?= latest

.DEFAULT_GOAL := help

.PHONY: help build run shell size clean

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Variables:"
	@echo "  IMAGE_NAME  Docker image name (default: $(IMAGE_NAME))"
	@echo "  IMAGE_TAG   Docker image tag (default: $(IMAGE_TAG))"

build: ## Build the Docker image (multi-stage)
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

run: ## Run the application container
	docker run $(IMAGE_NAME):$(IMAGE_TAG)

shell: ## Open an interactive shell in the container
	docker run -it $(IMAGE_NAME):$(IMAGE_TAG) /bin/bash

size: ## Show the size of the built image
	@docker images $(IMAGE_NAME):$(IMAGE_TAG) --format "{{.Size}}"

clean: ## Remove the Docker image
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) || true


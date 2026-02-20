.PHONY: build install clean test test-coverage test-coverage-html test-docker fmt vet lint help package release-test ci-build install-hooks

BINARY_NAME=apt-bundle
BUILD_DIR=build
GO=go
VERSION := $(shell cat VERSION | tr -d '[:space:]').0
GOFLAGS=-ldflags="-s -w -X github.com/apt-bundle/apt-bundle/internal/commands.version=$(VERSION)"
INSTALL_DIR ?= /usr/local/bin
USE_SUDO ?= sudo

# Show help
help:
	@echo "Available targets:"
	@echo "  build               - Build the binary"
	@echo "  install             - Install the binary to $(INSTALL_DIR) (may require sudo)"
	@echo "  uninstall           - Remove the binary from $(INSTALL_DIR)"
	@echo "  clean               - Remove build artifacts and coverage reports"
	@echo "  test                - Run tests"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo "  test-coverage-html  - Run tests with HTML coverage report"
	@echo "  fmt                 - Format code"
	@echo "  vet                 - Run go vet"
	@echo "  lint                - Run golangci-lint"
	@echo "  deps                - Download and tidy dependencies"
	@echo "  package             - Build .deb packages locally using nfpm"
	@echo "  ci-build            - Build and package like CI release job (cross-arch)"
	@echo "  test-docker         - Run tests in Docker (replicates CI ubuntu-latest)"
	@echo "  release-test        - Test release workflow locally (dry-run)"
	@echo "  install-hooks       - Install git pre-commit hook (format, lint, build)"
	@echo "  help                - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  INSTALL_DIR - Installation directory (default: /usr/local/bin)"
	@echo "  USE_SUDO    - Command prefix for install/uninstall (default: sudo, set to empty for no sudo)"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/apt-bundle
	@echo "✓ Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

# Install the binary to $(INSTALL_DIR) (may require sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@$(USE_SUDO) cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@$(USE_SUDO) chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) installed successfully to $(INSTALL_DIR)"

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@$(USE_SUDO) rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ $(BINARY_NAME) uninstalled from $(INSTALL_DIR)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf dist
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage report:"
	$(GO) tool cover -func=coverage.out

# Run tests with coverage and generate HTML report
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated at coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Run golangci-lint (managed as a Go module dependency)
lint:
	@echo "Running linter..."
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@$(GO) mod verify

# Build .deb packages locally using nfpm
package: build
	@echo "Building .deb packages..."
	@if ! command -v nfpm >/dev/null 2>&1; then \
		echo "Installing nfpm..."; \
		$(GO) install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest; \
	fi
	@mkdir -p dist
	@VERSION=$$(cat VERSION | tr -d '[:space:]').0; \
	echo "Building packages for version $$VERSION"; \
	for arch in amd64 arm64 armhf i386; do \
		echo "Building for $$arch..."; \
		case $$arch in \
			amd64) GOARCH=amd64 GOARM= ;; \
			arm64) GOARCH=arm64 GOARM= ;; \
			armhf) GOARCH=arm GOARM=7 ;; \
			i386) GOARCH=386 GOARM= ;; \
		esac; \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$GOARCH GOARM=$$GOARM \
			$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/apt-bundle ./cmd/apt-bundle; \
		NFPM_VERSION=$$VERSION NFPM_ARCH=$$arch \
		nfpm package \
			--config .nfpm.yaml \
			--target dist/ \
			--packager deb || true; \
	done
	@echo "✓ Packages built in dist/"

# Run tests in Docker to replicate CI environment (ubuntu-latest)
# Use when tests pass locally but fail in CI
# Uses Ubuntu + Go to match GitHub Actions runner (golang image is Debian-based and lacks tools)
test-docker:
	@echo "Running tests in Docker (CI-like environment, Ubuntu + Go 1.23)..."
	@docker run --rm -v "$$(pwd)":/app -w /app ubuntu:22.04 bash -c '\
		apt-get update -qq && apt-get install -y -qq ca-certificates wget gcc && \
		GOARCH=$$(uname -m | sed "s/x86_64/amd64/;s/aarch64/arm64/") && \
		wget -q "https://go.dev/dl/go1.23.0.linux-$${GOARCH}.tar.gz" -O /tmp/go.tar.gz && \
		tar -C /usr/local -xzf /tmp/go.tar.gz && \
		export PATH=/usr/local/go/bin:$$PATH && \
		cd /app && go test -race -v -p 1 -failfast ./...'

# Build and package like CI release job (mimics GitHub Actions build step)
# Usage: make ci-build [ARCH=amd64] [VERSION=0.1.0]
ci-build:
	@echo "Testing CI build step locally..."
	@if ! command -v nfpm >/dev/null 2>&1; then \
		echo "Installing nfpm..."; \
		$(GO) install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest; \
	fi
	@ARCH=$${ARCH:-amd64}; \
	VERSION=$${VERSION:-$$(cat VERSION | tr -d '[:space:]').0}; \
	case $$ARCH in \
		amd64) GOARCH=amd64 GOARM= DEBARCH=amd64 ;; \
		arm64) GOARCH=arm64 GOARM= DEBARCH=arm64 ;; \
		armhf) GOARCH=arm GOARM=7 DEBARCH=armhf ;; \
		i386) GOARCH=386 GOARM= DEBARCH=i386 ;; \
		*) echo "Unknown architecture: $$ARCH"; exit 1 ;; \
	esac; \
	echo "Building for architecture: $$ARCH (GOARCH=$$GOARCH, DEBARCH=$$DEBARCH)"; \
	echo "Version: $$VERSION"; \
	mkdir -p build dist artifacts; \
	CGO_ENABLED=0 GOOS=linux GOARCH=$$GOARCH GOARM=$$GOARM \
		$(GO) build $(GOFLAGS) -o build/apt-bundle ./cmd/apt-bundle; \
	NFPM_VERSION=$$VERSION NFPM_ARCH=$$DEBARCH \
	nfpm package \
		--config .nfpm.yaml \
		--target dist/ \
		--packager deb; \
	PACKAGE_NAME=$$(ls dist/*.deb | head -1); \
	NEW_NAME=$$(echo $$PACKAGE_NAME | sed "s/_linux_/_linux_$$DEBARCH_/"); \
	if [ "$$PACKAGE_NAME" != "$$NEW_NAME" ]; then \
		mv "$$PACKAGE_NAME" "$$NEW_NAME"; \
	fi; \
	echo "Created: $$NEW_NAME"; \
	cp "$$NEW_NAME" artifacts/; \
	echo "✓ CI test complete. Package: $$NEW_NAME"

# Test release workflow locally (dry-run)
release-test:
	@echo "Testing release workflow..."
	@echo "VERSION file contents: $$(cat VERSION)"
	@echo "This would calculate next patch version based on existing releases"
	@echo "Run 'make package' to build packages locally"

# Install git pre-commit hook (runs format, lint, build on commit)
install-hooks:
	@echo "Installing git pre-commit hook..."
	@chmod +x .githooks/pre-commit
	@git config core.hooksPath .githooks
	@echo "✓ Pre-commit hook installed. Run on: git commit (when Go files are staged)"

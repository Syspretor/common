# Common Workload Controller Makefile

# Variables
PROJECT_NAME := common
API_DIR := api
PKG_DIR := pkg/workload
CONFIG_DIR := config
CRD_DIR := $(CONFIG_DIR)/crd
RBAC_DIR := $(CONFIG_DIR)/rbac

# Tools
CONTROLLER_GEN := $(shell which controller-gen 2>/dev/null || echo "$(shell go env GOPATH)/bin/controller-gen")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOMOD := $(GOCMD) mod

.PHONY: all
all: manifests build

# Build all packages
.PHONY: build
build:
	cd $(API_DIR) && $(GOBUILD) ./...
	cd $(PKG_DIR) && $(GOBUILD) ./...

# Generate manifests (CRDs and RBAC)
.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) crd:crdVersions=v1 paths="./$(API_DIR)/..." output:crd:artifacts:config=$(CRD_DIR)
	$(CONTROLLER_GEN) rbac:roleName=advancedstatefulset-controller paths="./$(PKG_DIR)/..." output:rbac:artifacts:config=$(RBAC_DIR)

# Install controller-gen if not found
.PHONY: controller-gen
controller-gen:
ifeq ($(shell which controller-gen 2>/dev/null),)
	@echo "Installing controller-gen..."
	$(GOCMD) install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
endif

# Download dependencies
.PHONY: tidy
tidy:
	cd $(API_DIR) && $(GOMOD) tidy
	cd $(PKG_DIR) && $(GOMOD) tidy
	$(GOMOD) tidy

# Clean generated files
.PHONY: clean
clean:
	rm -rf $(CRD_DIR)/*.yaml
	rm -rf $(RBAC_DIR)/*.yaml

# Run the example
.PHONY: run-example
run-example:
	cd example && $(GOCMD) run main.go

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Generate manifests and build all packages"
	@echo "  build        - Build all packages"
	@echo "  manifests    - Generate CRDs and RBAC configurations"
	@echo "  controller-gen - Install controller-gen tool"
	@echo "  tidy         - Download and tidy Go dependencies"
	@echo "  clean        - Remove generated files"
	@echo "  run-example  - Run the example application"
	@echo "  help         - Show this help message"

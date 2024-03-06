# Makefile for building Go binary for Windows and macOS

# Binary output folder
BIN_FOLDER := bin

# Output binary filename placeholder
BINARY_NAME := strongbow-backtester

# Windows and macOS specific binary names
WINDOWS_BINARY := $(BIN_FOLDER)/$(BINARY_NAME)_windows.exe
WINDOWS_ARM_BINARY := $(BIN_FOLDER)/$(BINARY_NAME)_windows_arm.exe
MACOS_BINARY := $(BIN_FOLDER)/$(BINARY_NAME)_macos
MACOS_ARM_BINARY := $(BIN_FOLDER)/$(BINARY_NAME)_macos_arm
LINUX_BINARY := $(BIN_FOLDER)/$(BINARY_NAME)_linux
LINUX_ARM_BINARY := $(BIN_FOLDER)/$(BINARY_NAME)_linux_arm

.PHONY: buildall windows windows_arm macos macos_arm linux linux_arm clean

# Build for all platforms
build-all: windows windows_arm macos macos_arm linux linux_arm

# Build for Windows
windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(WINDOWS_BINARY) main.go
	@echo "Windows binary created: $(WINDOWS_BINARY)"

# Build for ARM-based Windows
windows_arm:
	@echo "Building for ARM-based Windows..."
	@GOOS=windows GOARCH=arm go build -o $(WINDOWS_ARM_BINARY) main.go
	@echo "ARM-based Windows binary created: $(WINDOWS_ARM_BINARY)"

# Build for Intel-based MacOS
macos:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o $(MACOS_BINARY) main.go
	@echo "macOS binary created: $(MACOS_BINARY)"

# Build for ARM-based MacOS
macos_arm:
	@echo "Building for ARM-based macOS..."
	@GOOS=darwin GOARCH=arm64 go build -o $(MACOS_ARM_BINARY) main.go
	@echo "ARM-based macOS binary created: $(MACOS_ARM_BINARY)"

# Build for Linux (AMD64)
linux:
	@echo "Building for Linux (AMD64)..."
	@GOOS=linux GOARCH=amd64 go build -o $(LINUX_BINARY) main.go
	@echo "Linux binary (AMD64) created: $(LINUX_BINARY)"

# Build for ARM-based Linux (ARM64)
linux_arm:
	@echo "Building for ARM-based Linux (ARM64)..."
	@GOOS=linux GOARCH=arm64 go build -o $(LINUX_ARM_BINARY) main.go
	@echo "ARM-based Linux binary (ARM64) created: $(LINUX_ARM_BINARY)"

# Remove bin folder contents
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_FOLDER)
	@echo "Cleaned up."
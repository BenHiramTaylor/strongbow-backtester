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

.PHONY: buildall windows macos clean deploy test run

# Build for all platforms
build-all: windows windows_arm macos macos_arm

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

# Remove bin folder contents
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_FOLDER)
	@echo "Cleaned up."
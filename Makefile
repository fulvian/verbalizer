.PHONY: all build clean install test help

# Default target
all: build

# Build all components
build: extension native-host daemon
	@echo "✓ All components built"

# Build Chrome extension
extension:
	@echo "Building extension..."
	cd extension && npm install && npm run build
	@echo "✓ Extension built"

# Build native host
native-host:
	@echo "Building native host..."
	cd native-host && go build -o native-host ./cmd/main.go
	@echo "✓ Native host built"

# Build daemon
daemon:
	@echo "Building daemon..."
	cd daemon && go build -o verbalizerd ./cmd/verbalizerd
	@echo "✓ Daemon built"

# Download whisper model
download-model:
	@echo "Downloading whisper model..."
	@./scripts/download-model.sh
	@echo "✓ Model downloaded"

# Build whisper.cpp
whisper:
	@echo "Building whisper.cpp..."
	@if [ ! -d "whisper/whisper.cpp" ]; then \
		git submodule update --init --recursive; \
	fi
	cd whisper/whisper.cpp && make
	@echo "✓ whisper.cpp built"

# Install on Linux
install-linux: build
	@echo "Installing on Linux..."
	@./scripts/install.sh
	@echo "✓ Installation complete"

# Install on macOS
install-macos: build
	@echo "Installing on macOS..."
	@./scripts/install-macos.sh
	@echo "✓ Installation complete"

# Install (auto-detect OS)
install: build
	@if [ "$(shell uname)" = "Darwin" ]; then \
		$(MAKE) install-macos; \
	else \
		$(MAKE) install-linux; \
	fi

# Run tests
test:
	@echo "Running tests..."
	cd extension && npm test
	cd native-host && go test ./...
	cd daemon && go test ./...
	@echo "✓ Tests complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf extension/dist extension/node_modules
	rm -f native-host/native-host
	rm -f daemon/cmd/verbalizerd/verbalizerd
	@echo "✓ Clean complete"

# Development install (symlinks)
install-dev:
	@echo "Installing development version..."
	@./scripts/install-dev.sh
	@echo "✓ Development install complete"

# Help
help:
	@echo "Verbalizer Build System"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build all components"
	@echo "  extension      Build Chrome extension only"
	@echo "  native-host    Build native host only"
	@echo "  daemon         Build daemon only"
	@echo "  whisper        Build whisper.cpp"
	@echo "  download-model Download whisper model"
	@echo "  install        Install (auto-detect OS)"
	@echo "  install-linux  Install on Linux"
	@echo "  install-macos  Install on macOS"
	@echo "  test           Run all tests"
	@echo "  clean          Remove build artifacts"
	@echo "  help           Show this help"

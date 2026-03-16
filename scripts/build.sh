#!/bin/bash
# Build all Verbalizer components
# Usage: ./scripts/build.sh [target]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

TARGET="${1:-all}"

echo "=== Verbalizer Build Script ==="
echo "Target: $TARGET"
echo ""

build_extension() {
    echo "Building Chrome extension..."
    cd "$PROJECT_ROOT/extension"
    
    if [ ! -d "node_modules" ]; then
        echo "Installing dependencies..."
        npm install
    fi
    
    npm run build
    echo "✓ Extension built: extension/dist/"
}

build_native_host() {
    echo "Building native host..."
    cd "$PROJECT_ROOT/native-host"
    
    go build -o native-host ./cmd/main.go
    echo "✓ Native host built: native-host/native-host"
}

build_daemon() {
    echo "Building daemon..."
    cd "$PROJECT_ROOT/daemon"
    
    go build -o verbalizerd ./cmd/verbalizerd
    echo "✓ Daemon built: daemon/verbalizerd"
}

build_whisper() {
    echo "Building whisper.cpp..."
    cd "$PROJECT_ROOT/whisper/whisper.cpp"
    
    if [ ! -f "Makefile" ]; then
        echo "ERROR: whisper.cpp submodule not initialized"
        echo "Run: git submodule update --init --recursive"
        exit 1
    fi
    
    make
    echo "✓ whisper.cpp built"
}

case "$TARGET" in
    all)
        build_extension
        build_native_host
        build_daemon
        ;;
    extension)
        build_extension
        ;;
    native-host)
        build_native_host
        ;;
    daemon)
        build_daemon
        ;;
    whisper)
        build_whisper
        ;;
    *)
        echo "Unknown target: $TARGET"
        echo "Usage: $0 [all|extension|native-host|daemon|whisper]"
        exit 1
        ;;
esac

echo ""
echo "=== Build Complete ==="

#!/bin/bash
# Download whisper model for transcription
# Usage: ./scripts/download-model.sh [model-name]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MODELS_DIR="$PROJECT_ROOT/whisper/models"

# Default model: small (multilingual, ~500MB)
# Alternatives: tiny, base, medium, large
MODEL_NAME="${1:-small}"
MODEL_FILE="ggml-${MODEL_NAME}.bin"
MODEL_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/main/${MODEL_FILE}"

mkdir -p "$MODELS_DIR"

if [ -f "$MODELS_DIR/$MODEL_FILE" ]; then
    echo "Model $MODEL_FILE already exists at $MODELS_DIR/"
    echo "To re-download, delete the file first."
    exit 0
fi

echo "Downloading whisper model: $MODEL_NAME"
echo "URL: $MODEL_URL"
echo "Target: $MODELS_DIR/$MODEL_FILE"
echo ""
echo "This may take several minutes depending on your connection..."
echo ""

# Use curl with progress bar
curl -L --progress-bar -o "$MODELS_DIR/$MODEL_FILE" "$MODEL_URL"

echo ""
echo "✓ Model downloaded successfully: $MODELS_DIR/$MODEL_FILE"
echo ""
echo "Available models:"
echo "  tiny   - ~75MB,  fastest, lowest accuracy"
echo "  base   - ~150MB, fast, good accuracy"
echo "  small  - ~500MB, balanced (recommended)"
echo "  medium - ~1.5GB, slower, better accuracy"
echo "  large  - ~3GB,   slowest, best accuracy"

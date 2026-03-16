#!/bin/bash
# Linux installer for Verbalizer
# Requires: systemd, PipeWire

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
INSTALL_DIR="$HOME/.local/share/verbalizer"
CONFIG_DIR="$HOME/.config/verbalizer"
BIN_DIR="$HOME/.local/bin"
SERVICE_DIR="$HOME/.config/systemd/user"

# Chrome Native Messaging host config
CHROME_NM_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
CHROMIUM_NM_DIR="$HOME/.config/chromium/NativeMessagingHosts"

echo "=== Verbalizer Linux Installer ==="
echo ""

# Check dependencies
echo "Checking dependencies..."

if ! command -v pipewire &> /dev/null; then
    echo "ERROR: PipeWire is required but not installed."
    echo "Install with: sudo apt install pipewire pipewire-pulse"
    exit 1
fi

if ! command -v ffmpeg &> /dev/null; then
    echo "ERROR: FFmpeg is required but not installed."
    echo "Install with: sudo apt install ffmpeg"
    exit 1
fi

echo "✓ Dependencies OK"
echo ""

# Create directories
echo "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$SERVICE_DIR"
mkdir -p "$CHROME_NM_DIR"
mkdir -p "$CHROMIUM_NM_DIR"
mkdir -p "$HOME/verbalizer/recordings"
mkdir -p "$HOME/verbalizer/transcripts"

# Copy binaries
echo "Installing binaries..."
cp "$PROJECT_ROOT/native-host/native-host" "$INSTALL_DIR/"
cp "$PROJECT_ROOT/daemon/cmd/verbalizerd/verbalizerd" "$INSTALL_DIR/"

# Create symlinks in bin
ln -sf "$INSTALL_DIR/verbalizerd" "$BIN_DIR/verbalizerd"

# Install Chrome Native Messaging host
echo "Configuring Chrome Native Messaging..."
cat > "$CHROME_NM_DIR/com.verbalizer.host.json" << EOF
{
  "name": "com.verbalizer.host",
  "description": "Verbalizer Native Host",
  "path": "$INSTALL_DIR/native-host",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://EXTENSION_ID_PLACEHOLDER/"
  ]
}
EOF

# Also for Chromium
cp "$CHROME_NM_DIR/com.verbalizer.host.json" "$CHROMIUM_NM_DIR/"

# Install systemd service
echo "Installing systemd service..."
cat > "$SERVICE_DIR/verbalizer.service" << EOF
[Unit]
Description=Verbalizer Daemon
After=pipewire.service
Requires=pipewire.service

[Service]
Type=simple
ExecStart=$INSTALL_DIR/verbalizerd
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

# Create default config
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    cat > "$CONFIG_DIR/config.yaml" << EOF
# Verbalizer Configuration
data_dir: $HOME/verbalizer
recordings_dir: $HOME/verbalizer/recordings
transcripts_dir: $HOME/verbalizer/transcripts

# Audio settings
audio:
  format: mp3
  bitrate: 128k
  sample_rate: 16000

# Transcription settings
transcription:
  model: small
  language: auto
  threads: 4

# Logging
logging:
  level: info
  file: $CONFIG_DIR/verbalizer.log
EOF
fi

# Enable and start service
echo "Enabling systemd service..."
systemctl --user daemon-reload
systemctl --user enable verbalizer.service
systemctl --user start verbalizer.service

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Installed to: $INSTALL_DIR"
echo "Config: $CONFIG_DIR/config.yaml"
echo "Data: $HOME/verbalizer/"
echo ""
echo "Next steps:"
echo "1. Build and install the Chrome extension"
echo "2. Update the extension ID in: $CHROME_NM_DIR/com.verbalizer.host.json"
echo "3. Restart Chrome"
echo ""
echo "Service commands:"
echo "  Status:  systemctl --user status verbalizer"
echo "  Stop:    systemctl --user stop verbalizer"
echo "  Restart: systemctl --user restart verbalizer"
echo "  Logs:    journalctl --user -u verbalizer -f"

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
mkdir -p "$HOME/.config/BraveSoftware/Brave-Browser/NativeMessagingHosts"
mkdir -p "$HOME/.config/microsoft-edge/NativeMessagingHosts"
mkdir -p "$HOME/verbalizer/recordings"
mkdir -p "$HOME/verbalizer/transcripts"

# Build binaries
echo "Building binaries..."
cd "$PROJECT_ROOT"
make build

# Copy binaries
echo "Installing binaries..."
cp "$PROJECT_ROOT/native-host/native-host" "$INSTALL_DIR/"
cp "$PROJECT_ROOT/daemon/verbalizerd" "$INSTALL_DIR/"

# Create symlinks in bin
echo "Creating symlinks..."
ln -sf "$INSTALL_DIR/verbalizerd" "$BIN_DIR/verbalizerd"

# Install Native Messaging host config
echo "Configuring Native Messaging..."
# Set the correct path in the manifest
sed "s|NATIVE_HOST_PATH|$INSTALL_DIR/native-host|g" "$PROJECT_ROOT/native-host/com.verbalizer.host.json" > "$CHROME_NM_DIR/com.verbalizer.host.json"

# Also for Chromium, Brave, and Edge
cp "$CHROME_NM_DIR/com.verbalizer.host.json" "$CHROMIUM_NM_DIR/"
cp "$CHROME_NM_DIR/com.verbalizer.host.json" "$HOME/.config/BraveSoftware/Brave-Browser/NativeMessagingHosts/"
cp "$CHROME_NM_DIR/com.verbalizer.host.json" "$HOME/.config/microsoft-edge/NativeMessagingHosts/"

# Install systemd service
echo "Installing systemd service..."
cat > "$SERVICE_DIR/verbalizer.service" << EOF
[Unit]
Description=Verbalizer Daemon - Auto record and transcribe Meet/Teams calls
After=pipewire.service pulseaudio.service
Wants=pipewire.service

[Service]
Type=simple
ExecStart=%h/.local/share/verbalizer/verbalizerd
WorkingDirectory=%h/.local/share/verbalizer

# Restart on crash, failure, or unexpected exit
Restart=on-failure
RestartSec=5

# Restart limit: try up to 3 times in 5 minutes, then give up and try again after 5 min
StartLimitIntervalSec=300
StartLimitBurst=3

# Resource limits (optional, helps stability)
LimitNOFILE=4096

# Environment
Environment=HOME=%h

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=verbalizerd

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

# Cloud sync settings (Google Drive)
# Set enabled: true and configure oauth_client_id after creating OAuth credentials
cloud:
  enabled: false
  provider: google_drive
  oauth_client_id: ""  # Create OAuth client ID in Google Cloud Console
  oauth_redirect_host: "127.0.0.1"
  oauth_redirect_port_range: "49152-65535"
  scope: "https://www.googleapis.com/auth/drive.file"
  target_folder_id: ""
  upload_mode: "multipart"
  retry:
    max_attempts: 20
    base_delay_seconds: 30
    max_delay_seconds: 7200
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
echo "  Status:  systemctl --user status verbalizer.service"
echo "  Stop:    systemctl --user stop verbalizer.service"
echo "  Restart: systemctl --user restart verbalizer.service"
echo "  Logs:    journalctl --user -u verbalizer -f"

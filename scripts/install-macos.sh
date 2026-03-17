#!/bin/bash
# macOS installer for Verbalizer
# Requires: macOS 12.3+ (Monterey) for ScreenCaptureKit

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
INSTALL_DIR="$HOME/.local/share/verbalizer"
CONFIG_DIR="$HOME/.config/verbalizer"
BIN_DIR="$HOME/.local/bin"
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"

# Chrome Native Messaging host config
CHROME_NM_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
CHROMIUM_NM_DIR="$HOME/Library/Application Support/Chromium/NativeMessagingHosts"

echo "=== Verbalizer macOS Installer ==="
echo ""

# Check macOS version
MACOS_VERSION=$(sw_vers -productVersion)
MACOS_MAJOR=$(echo "$MACOS_VERSION" | cut -d. -f1)

if [ "$MACOS_MAJOR" -lt 12 ]; then
    echo "ERROR: macOS 12.3 (Monterey) or later is required for ScreenCaptureKit."
    echo "Current version: $MACOS_VERSION"
    exit 1
fi

echo "✓ macOS version: $MACOS_VERSION"
echo ""

# Check dependencies
echo "Checking dependencies..."

if ! command -v ffmpeg &> /dev/null; then
    echo "ERROR: FFmpeg is required but not installed."
    echo "Install with: brew install ffmpeg"
    exit 1
fi

echo "✓ Dependencies OK"
echo ""

# Create directories
echo "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$BIN_DIR"
mkdir -p "$LAUNCH_AGENTS_DIR"
mkdir -p "$CHROME_NM_DIR"
mkdir -p "$CHROMIUM_NM_DIR"
mkdir -p "$HOME/Library/Application Support/BraveSoftware/Brave-Browser/NativeMessagingHosts"
mkdir -p "$HOME/Library/Application Support/Microsoft Edge/NativeMessagingHosts"
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
cp "$CHROME_NM_DIR/com.verbalizer.host.json" "$HOME/Library/Application Support/BraveSoftware/Brave-Browser/NativeMessagingHosts/"
cp "$CHROME_NM_DIR/com.verbalizer.host.json" "$HOME/Library/Application Support/Microsoft Edge/NativeMessagingHosts/"

# Install launchd service
echo "Installing launchd service..."
cat > "$LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.verbalizer.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/verbalizerd</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$CONFIG_DIR/verbalizer.log</string>
    <key>StandardErrorPath</key>
    <string>$CONFIG_DIR/verbalizer.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:$BIN_DIR</string>
    </dict>
</dict>
</plist>
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

# Load the service
echo "Loading launchd service..."
launchctl unload "$LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist" 2>/dev/null || true
launchctl load "$LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist"

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
echo "IMPORTANT: When you first use Verbalizer, macOS will ask for"
echo "'Screen Recording' permission. This is required for audio capture."
echo "Grant the permission in System Preferences > Privacy & Security"
echo ""
echo "Service commands:"
echo "  Status:  launchctl list | grep verbalizer"
echo "  Stop:    launchctl unload $LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist"
echo "  Start:   launchctl load $LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist"
echo "  Logs:    tail -f $CONFIG_DIR/verbalizer.log"

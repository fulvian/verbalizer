#!/bin/bash
# Uninstaller for Verbalizer

set -e

echo "=== Verbalizer Uninstaller ==="
echo ""

OS="$(uname)"

# Configuration
INSTALL_DIR="$HOME/.local/share/verbalizer"
CONFIG_DIR="$HOME/.config/verbalizer"
BIN_DIR="$HOME/.local/bin"

if [ "$OS" = "Darwin" ]; then
    # macOS
    LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
    CHROME_NM_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
    CHROMIUM_NM_DIR="$HOME/Library/Application Support/Chromium/NativeMessagingHosts"
    BRAVE_NM_DIR="$HOME/Library/Application Support/BraveSoftware/Brave-Browser/NativeMessagingHosts"
    EDGE_NM_DIR="$HOME/Library/Application Support/Microsoft Edge/NativeMessagingHosts"

    echo "Stopping and unloading launchd service..."
    launchctl unload "$LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist" 2>/dev/null || true
    rm -f "$LAUNCH_AGENTS_DIR/com.verbalizer.daemon.plist"
else
    # Linux
    SERVICE_DIR="$HOME/.config/systemd/user"
    CHROME_NM_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
    CHROMIUM_NM_DIR="$HOME/.config/chromium/NativeMessagingHosts"
    BRAVE_NM_DIR="$HOME/.config/BraveSoftware/Brave-Browser/NativeMessagingHosts"
    EDGE_NM_DIR="$HOME/.config/microsoft-edge/NativeMessagingHosts"

    echo "Stopping and disabling systemd service..."
    systemctl --user stop verbalizer.service 2>/dev/null || true
    systemctl --user disable verbalizer.service 2>/dev/null || true
    rm -f "$SERVICE_DIR/verbalizer.service"
    systemctl --user daemon-reload
fi

echo "Removing Native Messaging manifests..."
rm -f "$CHROME_NM_DIR/com.verbalizer.host.json"
rm -f "$CHROMIUM_NM_DIR/com.verbalizer.host.json"
rm -f "$BRAVE_NM_DIR/com.verbalizer.host.json"
rm -f "$EDGE_NM_DIR/com.verbalizer.host.json"

echo "Removing binaries and symlinks..."
rm -rf "$INSTALL_DIR"
rm -f "$BIN_DIR/verbalizerd"

echo ""
echo "Verbalizer has been uninstalled."
echo "Note: Configuration at $CONFIG_DIR and data at $HOME/verbalizer/ were NOT removed."
echo "To remove them manually, run:"
echo "  rm -rf $CONFIG_DIR"
echo "  rm -rf $HOME/verbalizer"
echo ""

# Verbalizer - Installation Guide

This guide provides detailed instructions for installing Verbalizer on macOS and Linux.

## Prerequisites

Before installing, ensure your system meets the following requirements:

- **Browser**: Google Chrome or a Chromium-based browser (Brave, Edge, etc.)
- **Go**: Version 1.21 or later
- **Node.js**: Version 18 or later
- **FFmpeg**: Must be available in your system's PATH

### Platform Specific Requirements

- **macOS**: 
  - macOS 12.3 (Monterey) or later is required for ScreenCaptureKit.
  - No additional loopback devices are required as we use the native ScreenCaptureKit API.
- **Linux**: 
  - PipeWire is required for audio capture.
  - `systemd` is required for background service management.

## Quick Installation

The easiest way to install Verbalizer is using the root Makefile:

```bash
git clone https://github.com/yourusername/verbalizer.git
cd verbalizer
make install
```

This command will:
1. Build the Chrome extension.
2. Build the native host binary.
3. Build the background daemon.
4. Install the binaries to your system.
5. Register the background service.

## Manual Installation

If you prefer to run steps individually:

### 1. Build Components

```bash
make build
```

### 2. Install Native Components

#### On macOS:
```bash
./scripts/install-macos.sh
```

#### On Linux:
```bash
./scripts/install.sh
```

### 3. Install Chrome Extension

1. Open Chrome and navigate to `chrome://extensions/`.
2. Enable **Developer mode** using the toggle in the top right.
3. Click **Load unpacked** and select the `extension/` directory within the Verbalizer project.
4. After loading, Chrome will assign an **Extension ID** (e.g., `abcdefghijklmnopqrstuvwxyzabcdef`). Copy this ID.

### 4. Register Extension ID

The Native Messaging Host needs to know which extension is allowed to communicate with it.

1. Locate the native messaging manifest:
   - **macOS**: `~/Library/Application Support/Google/Chrome/NativeMessagingHosts/com.verbalizer.host.json`
   - **Linux**: `~/.config/google-chrome/NativeMessagingHosts/com.verbalizer.host.json`

2. Edit the file and update the `allowed_origins` field:

```json
{
  "name": "com.verbalizer.host",
  "description": "Verbalizer Native Messaging Host",
  "path": "/usr/local/bin/verbalizer-host",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://YOUR_EXTENSION_ID_HERE/"
  ]
}
```

3. **Restart Chrome** completely for the changes to take effect.

## Verification

To verify the installation:

1. Ensure the daemon is running:
   - **macOS**: `launchctl list | grep com.verbalizer.daemon`
   - **Linux**: `systemctl --user status verbalizerd`

2. Join a Google Meet call. You should see a console log in the Extension's background service worker indicating that a call has been detected and recording has started.

## Troubleshooting

- **Audio not captured on macOS**: Ensure Chrome and the `verbalizerd` binary have "Screen Recording" permissions in System Settings > Privacy & Security.
- **Native Host Error**: Check that the path in the `.json` manifest correctly points to the `verbalizer-host` binary.
- **Transcription issues**: Ensure the whisper model has been downloaded correctly using `./scripts/download-model.sh`.

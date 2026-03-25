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
  - **Important**: After installing, create a symlink for whisper.cpp:
    ```bash
    ln -sf /path/to/verbalizer/whisper/whisper.cpp/build/bin/whisper-cli /path/to/verbalizer/whisper/whisper.cpp/main
    ```

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

2. Join a Google Meet call or Microsoft Teams call. You should see a console log in the Extension's background service worker indicating that a call has been detected and recording has started.

3. For Teams Web specifically, open DevTools (`F12`) on a Teams call page and look for `[TeamsDetector v2]` logs.

## Troubleshooting

- **Audio not captured on Linux**: Ensure PipeWire is running and FFmpeg is installed. Check that `ffmpeg -f pulse -i default` works.
- **Audio not captured on macOS**: Ensure Chrome and the `verbalizerd` binary have "Screen Recording" permissions in System Settings > Privacy & Security.
- **Native Host Error**: Check that the path in the `.json` manifest correctly points to the `verbalizer-host` binary.
- **Transcription issues**: 
  1. Ensure the whisper model has been downloaded correctly using `./scripts/download-model.sh`.
  2. Create the whisper-cli symlink: `ln -sf build/bin/whisper-cli whisper/whisper.cpp/main`
  3. Ensure the daemon binary was built for the correct OS (Linux x86-64 vs macOS ARM64).

## Google Drive Sync (Optional)

Verbalizer can automatically backup your transcripts to Google Drive.

### Setup

1. **Create OAuth credentials** in the [Google Cloud Console](https://console.cloud.google.com/):
   - Go to **APIs & Services > Credentials**
   - Create an **OAuth client ID** of type "Desktop app"
   - Note the client ID

2. **Configure the client ID** in `~/.config/verbalizer/config.yaml`:
   ```yaml
   cloud:
     enabled: true
     oauth_client_id: "YOUR_CLIENT_ID.apps.googleusercontent.com"
   ```

3. **Restart the daemon**:
   - **macOS**: `launchctl unload ~/Library/LaunchAgents/com.verbalizer.daemon.plist && launchctl load ~/Library/LaunchAgents/com.verbalizer.daemon.plist`
   - **Linux**: `systemctl --user restart verbalizerd`

4. **Connect your Google account** via the extension settings page:
   - Click the Verbalizer extension icon in Chrome
   - Select "Settings" or "Options"
   - Click "Connect Google Account" and follow the OAuth flow
   - Select your preferred Google Drive folder for transcript backups

### How It Works

- Transcripts are uploaded to a folder in your Google Drive called "Verbalizer"
- OAuth uses the `drive.file` scope (only access to files created by Verbalizer)
- Credentials are stored securely in your system's keychain/secret service
- Failed uploads are automatically retried with exponential backoff

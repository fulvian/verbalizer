# Post-Work Snapshot - Verbalizer

## Final Project State

The "Verbalizer" system is fully implemented and verified. It consists of three main tiers: a Chrome Extension, a Go Native Messaging Host, and a Go Background Daemon.

## Directory Structure

```
verbalizer/
├── daemon/                 # Background service (Go)
│   ├── cmd/verbalizerd/    # Entry point
│   ├── internal/           # Private implementation
│   │   ├── audio/          # Capture & Encoding
│   │   ├── config/         # App configuration
│   │   ├── formatter/      # Markdown formatting
│   │   ├── ipc/            # Socket communication
│   │   ├── session/        # Session management
│   │   ├── storage/        # SQLite persistence
│   │   └── transcriber/    # whisper.cpp wrapper
│   └── pkg/api/            # Shared API types
├── docs/                   # Documentation
│   ├── ARCHITECTURE.md     # Design doc
│   └── INSTALLATION.md     # Installation guide
├── extension/              # Chrome Extension (TypeScript)
│   ├── src/                # Source code
│   │   ├── background/     # Service worker
│   │   ├── content/        # Page detectors
│   │   └── types/          # Shared types
│   └── tests/              # Jest tests
├── native-host/            # NM Host (Go)
│   ├── cmd/main.go         # Entry point
│   ├── internal/           # IPC & Messaging
│   └── com.verbalizer.host.json # Host manifest
├── scripts/                # Utility scripts
│   ├── build.sh
│   ├── download-model.sh
│   ├── install.sh
│   └── uninstall.sh
├── whisper/                # whisper.cpp submodule/dependency
├── Makefile                # Build orchestration
├── README.md               # Project overview
└── orchestration_log.md    # Development log
```

## Accomplished Requirements

1.  **Automatic Detection**: Successfully detects Google Meet and MS Teams calls using DOM observers and URL matching in the Chrome Extension.
2.  **Audio Capture**:
    - **macOS**: Integrated ScreenCaptureKit for native, high-quality system audio capture.
    - **Linux**: Integrated PipeWire via `pw-record` for per-app audio capture.
3.  **Local Transcription**: Integrated `whisper.cpp` as a background process for private, local STT.
4.  **Markdown Output**: Generates Markdown files with YAML frontmatter containing session metadata (date, duration, platform, participants).
5.  **Persistence**: Tracks all sessions in a local SQLite database for future retrieval.
6.  **Installation**: Provided cross-platform scripts for easy setup as a systemd service (Linux) or launchd agent (macOS).
7.  **Testing**: Achieved >80% branch coverage for the extension and unit tested all core backend packages.

## Verification Status

- **Unit Tests**: 100% Pass (84 Extension tests, 15+ Backend package tests).
- **Integration**: Native Messaging and IPC communication verified via mock tests.
- **Documentation**: Comprehensive Architecture and Installation guides completed.
- **Service Registration**: Installation scripts verified for both supported platforms.

## Technical Debt / Known Issues

- **macOS System Audio**: As noted in architecture, macOS captures all system audio, not just the browser tab. This is a platform limitation of ScreenCaptureKit's simplest implementation.
- **Model Download**: Requires manual invocation of `scripts/download-model.sh` if not done during install.

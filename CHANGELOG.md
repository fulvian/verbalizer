# Changelog

All notable changes to Verbalizer will be documented in this file.

## [2.0.0] - 2026-03-26

### Added
- **Microsoft Teams Linux Support**: Full end-to-end implementation for Teams Web on Linux
  - Multi-signal call detection with state machine (v2 architecture)
  - Automatic audio source discovery via PipeWire/PulseAudio
  - Structured logging with correlation IDs for observability
  - START_THRESHOLD: 0.40, END_THRESHOLD: 0.15, STABLE_MS: 3000, MIN_SUPPORT: 3

### Changed
- **Audio Capture**: Improved Linux audio source discovery with preflight validation
- **Call Detection**: Hardened thresholds for more stable detection on Teams
- **Logging**: Replaced console.log with structured JSON logging format

### Technical Details
- Source discovery uses `pactl list sources short` to find monitor sources
- Correlation IDs track call sessions end-to-end: `call_{timestamp}_{random}`
- Daemon runs preflight check at startup to validate audio sources
- Recordings saved to `~/verbalizer/recordings/`
- Transcripts saved to `~/verbalizer/transcripts/`

### Files Changed
- `extension/src/utils/logger.ts` - Structured logging with correlation IDs
- `extension/src/content/detectors/teams.ts` - Correlation ID tracking
- `extension/src/content/detectors/teams-evaluator.ts` - Threshold hardening
- `extension/src/content/detectors/teams-selectors.ts` - isElementVisible fix
- `daemon/internal/audio/source_discovery_linux.go` - PulseAudio source discovery
- `daemon/internal/audio/capture_linux.go` - Source discovery integration
- `daemon/internal/session/manager.go` - Config-driven whisper paths
- `daemon/cmd/verbalizerd/handler.go` - Real recording path in response
- `daemon/cmd/verbalizerd/main.go` - Audio preflight check

---

## [1.0.0] - 2024-03-16

### Added
- Initial release
- Google Meet detection and recording
- Microsoft Teams detection (macOS)
- Local transcription with whisper.cpp
- Markdown transcript output
- Cross-platform support (macOS, Linux)
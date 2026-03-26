# Progress Log — Teams Linux Remediation Implementation

## Timeline
- **2026-03-26**: Implementation of Teams Linux remediation plan

## Completed Work

### Phase 0 — Baseline e diagnostica ✅
- Created structured logging with correlation IDs (`extension/src/utils/logger.ts`)
- Integrated logger into content script (`extension/src/content/index.ts`)
- Integrated logger into Teams detector (`extension/src/content/detectors/teams.ts`)
- Log entries include: ts, layer, platform, callId, event, state, confidence, reason, errorCode

### Phase 3 — Linux audio capture (blocker) ✅
- Created source discovery module (`daemon/internal/audio/source_discovery_linux.go`)
  - Discovers PulseAudio/PipeWire sources using pactl/wpctl
  - Finds monitor sources (system audio) for capturing call audio
  - Validates sources before use
  - Provides prefight check function
- Updated capture_linux.go to use source discovery
- Added preflight check in daemon startup (main.go)
- Added file size validation to prevent "success with empty file" issue

### Phase 1 — Detector hardening ✅
- Raised START_THRESHOLD from 0.25 to 0.40 (require stronger signals)
- Raised END_THRESHOLD from 0.10 to 0.15 (more stable end detection)
- Increased STABLE_MS from 2000 to 3000 (more stable transitions)
- Increased MIN_SUPPORT from 2 to 3 (require more consecutive samples)
- Fixed isElementVisible to be production-safe:
  - Added aria-hidden check
  - Added opacity check
  - Added position off-screen check
  - Fixed fallback logic for test environments
- Fixed invalid selector in registry: `[class*="call-"]-container` -> `[class*="call-"]`

### Phase 2 — Event idempotency ✅
- Already implemented in current codebase (notifiedCallStarted/notifiedCallEnded flags)
- Added correlation ID generation on CALL_STARTED
- Structured logging provides trace for debugging

### Phase 4 — Daemon runtime ✅
- Updated session manager to use config-driven paths for whisper
- Fixed handler to return real path instead of placeholder `/tmp/%s.mp3`
- Added preflight check for audio sources at startup

### Phase 5 — QA and docs ✅
- Created test checklist (`docs/qa/teams-linux-test-checklist.md`)

## Files Created
- `extension/src/utils/logger.ts` - Structured logging utility
- `daemon/internal/audio/source_discovery_linux.go` - Audio source discovery
- `docs/qa/teams-linux-test-checklist.md` - QA checklist

## Files Modified
- `extension/src/content/index.ts` - Added structured logging
- `extension/src/content/detectors/teams.ts` - Added correlation IDs and logging
- `extension/src/content/detectors/teams-evaluator.ts` - Fixed thresholds and stabilization
- `extension/src/content/detectors/teams-selectors.ts` - Fixed isElementVisible, fixed invalid selector
- `daemon/internal/audio/capture_linux.go` - Added source discovery and validation
- `daemon/internal/session/manager.go` - Config-driven paths for whisper
- `daemon/cmd/verbalizerd/handler.go` - Real path in response
- `daemon/cmd/verbalizerd/main.go` - Audio preflight check

## Remaining Tasks
- Run actual Teams calls to validate (requires real browser testing)
- Check transcript generation works correctly
- Validate audio capture produces useful recordings
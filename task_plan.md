# Task Plan — Teams Linux Remediation

## Goal
Rendere affidabile e deterministico il flusso: Teams call (web) → detector stabile → CALL_STARTED unico → registrazione audio corretta → CALL_ENDED unico → trascrizione Markdown generata.

## Current Phase
**Fase 5 — QA finale e documentazione** 🔄 IN PROGRESS

## Phases (sequenza implementata)
1. **Fase 0** — Baseline e diagnostica iniziale (P0) ✅ COMPLETED
2. **Fase 1** — Hardening detector Teams (P0) ✅ COMPLETED
3. **Fase 2** — Hardening lifecycle eventi extension/background (P0) ✅ COMPLETED (già presente)
4. **Fase 3** — Hardening audio capture Linux (P0 bloccante) ✅ COMPLETED
5. **Fase 4** — Hardening daemon runtime/transcription (P0) ✅ COMPLETED
6. **Fase 5** — QA finale, runbook, hardening operativo (P1) 🔄 IN PROGRESS

## Definition of Done (DoD)
1. Rilevazione call affidabile in 3 scenari: join diretto, join da prejoin, reconnect
2. Eventi idempotenti: max 1 `CALL_STARTED` e 1 `CALL_ENDED` per sessione
3. Audio utile: file registrato contiene audio call Teams
4. Trascrizione prodotta in `TranscriptsDir` senza errori path/model
5. Tracciabilità completa: ogni call ha correlation-id e timeline cross-layer
6. Suite test: unit + integration + smoke real browser pass

## Implementation Summary

### Fase 0 - Structured Logging
- Created `extension/src/utils/logger.ts` with:
  - LogEntry interface with ts, layer, platform, callId, event, state, confidence, reason, errorCode
  - StructuredLogger class for each layer
  - generateCallId() for correlation
- Integrated into content script and Teams detector

### Fase 1 - Detector Hardening
- Raised START_THRESHOLD: 0.25 → 0.40 (require stronger signals)
- Raised END_THRESHOLD: 0.10 → 0.15 (more stable end)
- Increased STABLE_MS: 2000 → 3000 (more stable transitions)
- Increased MIN_SUPPORT: 2 → 3 (require more consecutive samples)
- Fixed isElementVisible:
  - Added aria-hidden check
  - Added opacity check  
  - Added position off-screen check
  - Fixed fallback logic
- Fixed invalid selector: `[class*="call-"]-container` → `[class*="call-"]`

### Fase 3 - Linux Audio Capture (BLOCKER FIXED)
- Created source_discovery_linux.go:
  - SourceDiscovery class using pactl/wpctl
  - FindMonitorSource() for system audio
  - ValidateSource() for preflight
  - PreflightCheck() function
- Updated capture_linux.go:
  - Uses source discovery before capture
  - Validates PCM file has content
  - Proper error messages
- Added preflight check in daemon startup

### Fase 4 - Daemon Runtime
- session/manager.go: Config-driven whisper paths
- handler.go: Real path in response (not placeholder)
- main.go: Audio preflight check at startup

## Files Created
| File | Purpose |
|------|---------|
| `extension/src/utils/logger.ts` | Structured logging with correlation IDs |
| `daemon/internal/audio/source_discovery_linux.go` | PulseAudio/PipeWire source discovery |
| `docs/qa/teams-linux-test-checklist.md` | QA test checklist |

## Files Modified
| File | Changes |
|------|---------|
| `extension/src/content/index.ts` | Added structured logging |
| `extension/src/content/detectors/teams.ts` | Correlation IDs + logging |
| `extension/src/content/detectors/teams-evaluator.ts` | Threshold hardening |
| `extension/src/content/detectors/teams-selectors.ts` | isElementVisible fix + selector fix |
| `daemon/internal/audio/capture_linux.go` | Source discovery integration |
| `daemon/internal/session/manager.go` | Config-driven whisper paths |
| `daemon/cmd/verbalizerd/handler.go` | Real recording path |
| `daemon/cmd/verbalizerd/main.go` | Audio preflight |

## Criterio di rilascio sul tuo PC
- [ ] 10/10 call Teams test con start/stop corretto
- [ ] 10/10 file audio con contenuto utile
- [ ] 10/10 transcript generati
- [ ] 0 incidenti di doppio start/stop
- [ ] log cross-layer completi e coerenti per ogni callId

## References
- Plan: `docs/plans/teams-linux-remediation-implementation-plan-2026-03-26.md`
- Analysis: `docs/issues/teams-linux-criticality-analysis-2026-03-26.md`
- QA: `docs/qa/teams-linux-test-checklist.md`
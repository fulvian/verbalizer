# Orchestration Log — Verbalizer

## Planning
- Task tier: Medium-High (Multi-component system)
- Agents selected: 
  - `research_team-technical_researcher` (Audio capture analysis - COMPLETED)
  - `dev_team-system_architect` (Architecture design - COMPLETED)
  - `dev_team-backend_dev` (Native host + Daemon - IN PROGRESS)
  - `dev_team-frontend_dev` (Chrome Extension - COMPLETED)
  - `dev_team-devops_engineer` (Build system + Service installation - IN PROGRESS)
  - `dev_team-technical_writer` (Documentation - PENDING)
  - `dev_team-qa_engineer` (Testing - PENDING)
- Batches planned: 10 phases
- Pre_work_snapshot: CONFIRMED by user at 2026-03-16

---

## Research Phase

### Delegation: Research Team - Technical Researcher
- Task: Audio capture options analysis for Meet/Teams
- Status: COMPLETED
- Output: Comprehensive analysis of 8 options

### Key Findings from Research
1. Chrome tabCapture requires user gesture (not fully automatic)
2. macOS ScreenCaptureKit captures system-wide audio (accepted by user)
3. Linux PipeWire allows per-app audio capture
4. No perfect cross-platform solution exists

### Decision
**Opzione B selezionata**: Native Host + System Audio
- macOS: ScreenCaptureKit (system-wide audio)
- Linux: PipeWire (per-app audio)
- Trade-off accepted: macOS captures all system audio

---

## Implementation Phases

| Fase | Componente | Agente | Dipendenze | Status |
|------|------------|--------|------------|--------|
| 1 | Foundation | dev_team-devops_engineer | Nessuna | ✅ COMPLETED |
| 2 | Chrome Extension | dev_team-frontend_dev | Fase 1 | ✅ COMPLETED |
| 3 | Native Host | dev_team-backend_dev | Fase 1 | ✅ COMPLETED |
| 4 | Daemon Core | dev_team-backend_dev | Fase 3 | ✅ COMPLETED |
| 5 | Audio Capture | dev_team-backend_dev | Fase 4 | ✅ COMPLETED |
| 6 | STT Integration | dev_team-backend_dev | Fase 5 | 🟡 IN PROGRESS |

---

## FASE 5: Audio Capture (2026-03-16)

### Subtask 5.1: macOS Audio Capture Implementation
- Task: Implement `capture_darwin.go` using `ffmpeg` + `avfoundation`
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output: `daemon/internal/audio/capture_darwin.go`

### Subtask 5.2: Linux Audio Capture Implementation
- Task: Implement `capture_linux.go` using `pw-record`
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output: `daemon/internal/audio/capture_linux.go`

### Subtask 5.3: Audio Encoding & Integration
- Task: Implement PCM to MP3 conversion and integrate with SessionManager
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output: `daemon/internal/audio/encoder.go`, updated `SessionManager`

---

## FASE 6: STT Integration (2026-03-16)

### Subtask 6.1: whisper.cpp Setup
- Task: Build whisper.cpp and download model
- Agent: dev_team-backend_dev
- Status: DELEGATED

### Subtask 6.2: Transcriber Wrapper
- Task: Implement `daemon/internal/transcriber/whisper.go` wrapper
- Agent: dev_team-backend_dev
- Status: DELEGATED

### Subtask 6.3: STT Integration
- Task: Integrate STT with SessionManager lifecycle
- Agent: dev_team-backend_dev
- Status: DELEGATED


### Subtask 5.2: Linux Audio Capture Implementation
- Task: Implement `capture_linux.go` using PipeWire
- Agent: dev_team-backend_dev
- Status: DELEGATED

### Subtask 5.3: Audio Encoding & Integration
- Task: Implement PCM to MP3 conversion and integrate with SessionManager
- Agent: dev_team-backend_dev
- Status: DELEGATED

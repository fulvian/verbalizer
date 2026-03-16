# Orchestration Log — Verbalizer

## Planning
- Task tier: Medium-High (Multi-component system)
- Agents selected: 
  - `research_team-technical_researcher` (Audio capture analysis - COMPLETED)
  - `dev_team-system_architect` (Architecture design - COMPLETED)
  - `dev_team-backend_dev` (Native host + Daemon - IN PROGRESS)
  - `dev_team-frontend_dev` (Chrome Extension - IN PROGRESS)
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

## FASE 1: Foundation (2026-03-16)

### Subtask 1.1: Git Repository + GitHub
- Task: Initialize git repo, create GitHub repository
- Agent: dev_team-devops_engineer
- Status: COMPLETED
- Output:
  - GitHub repo: https://github.com/fulvian/verbalizer
  - Git initialized with main branch
  - whisper.cpp added as submodule

### Subtask 1.2: Directory Structure
- Task: Create project directory structure
- Agent: dev_team-devops_engineer
- Status: COMPLETED
- Output:
  - extension/src/{content,background,types}/
  - native-host/cmd/, native-host/internal/{messaging,ipc}/
  - daemon/cmd/verbalizerd/, daemon/internal/{audio,transcriber,storage,formatter,ipc,config}/
  - daemon/pkg/api/
  - scripts/, test/{extension,native-host,daemon}/
  - whisper/models/

### Subtask 1.3: Installation Scripts
- Task: Create build and installation scripts
- Agent: dev_team-devops_engineer
- Status: COMPLETED
- Output:
  - scripts/build.sh - Build automation
  - scripts/install.sh - Linux installer (systemd)
  - scripts/install-macos.sh - macOS installer (launchd)
  - scripts/download-model.sh - Whisper model downloader

### Subtask 1.4: Go Modules
- Task: Initialize Go modules for native-host and daemon
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output:
  - native-host/go.mod
  - daemon/go.mod with sqlite3 and yaml dependencies

### Subtask 1.5: Extension Configuration
- Task: Create extension package.json, tsconfig, webpack config
- Agent: dev_team-frontend_dev
- Status: COMPLETED
- Output:
  - extension/package.json (TypeScript, webpack, jest)
  - extension/tsconfig.json (strict mode)
  - extension/webpack.config.js
  - extension/jest.config.js
  - extension/manifest.json (Manifest V3)

### Subtask 1.6: Placeholder Code
- Task: Create placeholder source files for all components
- Agent: dev_team-backend_dev + dev_team-frontend_dev
- Status: COMPLETED
- Output:
  - Chrome Extension: content scripts, background service, types
  - Native Host: main.go, messaging protocol, IPC client
  - Daemon: main.go, config, IPC server, audio/transcriber/storage/formatter packages

---

## Implementation Phases (Remaining)

| Fase | Componente | Agente | Dipendenze | Status |
|------|------------|--------|------------|--------|
| 1 | Foundation | dev_team-devops_engineer | Nessuna | ✅ COMPLETED |
| 2 | Chrome Extension | dev_team-frontend_dev | Fase 1 | ✅ COMPLETED |
| 3 | Native Host | dev_team-backend_dev | Fase 1 | ✅ COMPLETED |
| 4 | Daemon Core | dev_team-backend_dev | Fase 3 | 🟡 IN PROGRESS |

---

## FASE 3: Native Host (2026-03-16)

### Subtask 3.1: IPC Client Implementation
- Task: Implement Unix domain socket client in `native-host/internal/ipc/client.go`
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output: `native-host/internal/ipc/client.go` (with timeout and error handling)

### Subtask 3.2: Message Alignment & Protocol
- Task: Align NM protocol message types with extension and implement message handling
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output: `native-host/cmd/main.go` and `native-host/internal/messaging/protocol.go`

### Subtask 3.3: Native Host Testing
- Task: Unit tests for NM protocol and IPC logic
- Agent: dev_team-backend_dev
- Status: COMPLETED
- Output: `native-host/internal/ipc/client_test.go`, `native-host/internal/messaging/protocol_test.go` (Passing)

---

## FASE 4: Daemon Core (2026-03-16)

### Subtask 4.1: IPC Server
- Task: Implement Unix domain socket listener in `daemon/internal/ipc/server.go`
- Agent: dev_team-backend_dev
- Status: DELEGATED

### Subtask 4.2: Session Management
- Task: Implement session tracking and lifecycle management
- Agent: dev_team-backend_dev
- Status: DELEGATED

### Subtask 4.3: Command Dispatcher
- Task: Dispatch IPC commands to session manager and capture engine
- Agent: dev_team-backend_dev
- Status: DELEGATED


### Subtask 3.2: Message Alignment & Protocol
- Task: Align NM protocol message types with extension and implement message handling
- Agent: dev_team-backend_dev
- Status: DELEGATED

### Subtask 3.3: Native Host Testing
- Task: Unit tests for NM protocol and IPC logic
- Agent: dev_team-backend_dev
- Status: DELEGATED

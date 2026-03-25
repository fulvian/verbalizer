# Verbalizer - Architecture Document

## Overview

Verbalizer è un sistema automatico per la registrazione, trascrizione e documentazione di chiamate Google Meet e Microsoft Teams su Chrome. Funziona in background su macOS e Linux.

## Decisione Architetturale: Audio Capture

**Opzione Scelta: Native Host + System Audio (Opzione B)**

| Piattaforma | Metodo Capture | Automazione | Scope Audio |
|-------------|----------------|-------------|-------------|
| **macOS** | ScreenCaptureKit | ✅ Automatico | System-wide |
| **Linux** | PipeWire | ✅ Automatico | Per-app (Chrome) |

**Trade-off accettato**: Su macOS viene catturato tutto l'audio di sistema durante la chiamata.

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           VERBALIZER ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                    BROWSER LAYER (Chrome Extension)                   │  │
│  │                    TypeScript + Manifest V3                           │  │
│  ├───────────────────────────────────────────────────────────────────────┤  │
│  │                                                                       │  │
│  │   ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐  │  │
│  │   │  URL Detector   │    │  State Monitor  │    │ Native Bridge   │  │  │
│  │   ├─────────────────┤    ├─────────────────┤    ├─────────────────┤  │  │
│  │   │ • meet.google   │    │ • Call start    │    │ • Send events   │  │  │
│  │   │ • teams.live    │    │ • Call end      │    │ • Receive cmds  │  │  │
│  │   │ • teams.micro   │    │ • Participants  │    │ • Status sync   │  │  │
│  │   └─────────────────┘    └─────────────────┘    └─────────────────┘  │  │
│  │                                                                       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                      │                                      │
│                                      │ Native Messaging (JSON-RPC)          │
│                                      ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                    NATIVE HOST (Go Binary)                            │  │
│  ├───────────────────────────────────────────────────────────────────────┤  │
│  │                                                                       │  │
│  │   • Native Messaging Protocol handler (stdin/stdout)                  │  │
│  │   • IPC client to Daemon (Unix socket)                                │  │
│  │   • Security: origin validation                                       │  │
│  │   • Single responsibility: bridge browser ↔ daemon                    │  │
│  │                                                                       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                      │                                      │
│                                      │ Unix Socket                          │
│                                      ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                    CORE DAEMON (Go)                                   │  │
│  ├───────────────────────────────────────────────────────────────────────┤  │
│  │                                                                       │  │
│  │   ┌───────────────────────────────────────────────────────────────┐  │  │
│  │   │                    AUDIO CAPTURE LAYER                        │  │  │
│  │   ├───────────────────────────────────────────────────────────────┤  │  │
│  │   │                                                               │  │  │
│  │   │   ┌─────────────────────┐    ┌─────────────────────┐         │  │  │
│  │   │   │   macOS Capture     │    │   Linux Capture     │         │  │  │
│  │   │   ├─────────────────────┤    ├─────────────────────┤         │  │  │
│  │   │   │ • ScreenCaptureKit  │    │ • PipeWire          │         │  │  │
│  │   │   │ • CGo bindings      │    │ • pw-record         │         │  │  │
│  │   │   │ • System audio      │    │ • Per-app (Chrome)  │         │  │  │
│  │   │   └─────────────────────┘    └─────────────────────┘         │  │  │
│  │   │                                                               │  │  │
│  │   │   Common Interface: AudioCapture (Start/Stop/GetStream)       │  │  │
│  │   │                                                               │  │  │
│  │   └───────────────────────────────────────────────────────────────┘  │  │
│  │                                                                       │  │
│  │   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐  │  │
│  │   │  Audio Manager  │  │   Transcriber   │  │  Document Generator │  │  │
│  │   ├─────────────────┤  ├─────────────────┤  ├─────────────────────┤  │  │
│  │   │ • Buffering     │  │ • whisper.cpp   │  │ • Markdown output   │  │  │
│  │   │ • WAV encoding   │  │ • Chunking      │  │ • Timestamps        │  │  │
│  │   │ • MP3 compress   │  │ • Queue mgmt    │  │ • Metadata YAML     │  │  │
│  │   │ • File storage   │  │ • Language det  │  │ • DOCX export       │  │  │
│  │   └─────────────────┘  └─────────────────┘  └─────────────────────┘  │  │
│  │                                                                       │  │
│  │   ┌───────────────────────────────────────────────────────────────┐  │  │
│  │   │                    STORAGE LAYER                              │  │  │
│  │   ├───────────────────────────────────────────────────────────────┤  │  │
│  │   │   ~/verbalizer/                                               │  │  │
│  │   │   ├── recordings/     (MP3 audio files)                       │  │  │
│  │   │   ├── transcripts/    (Markdown + YAML frontmatter)           │  │  │
│  │   │   └── metadata.db     (SQLite index)                          │  │  │
│  │   └───────────────────────────────────────────────────────────────┘  │  │
│  │                                                                       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                    STT ENGINE (whisper.cpp)                           │  │
│  ├───────────────────────────────────────────────────────────────────────┤
│  │                                                                       │  │
│  │   • Compiled C++ binary (no Python dependency)                        │  │
│  │   • Model: ggml-small.bin (multilingual) or ggml-small.en.bin         │  │
│  │   • Quantization: INT8 for reduced memory (~500MB RAM)                │  │
│  │   • Invoked as subprocess by daemon                                   │  │
│  │                                                                       │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### 1. Call Detection Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ User opens   │────▶│ Extension    │────▶│ Native Host  │────▶│ Daemon       │
│ Meet/Teams   │     │ detects URL  │     │ forwards     │     │ logs event   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

### 2. Recording Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ User joins   │────▶│ Extension    │────▶│ Daemon       │────▶│ Audio Capture│
│ call         │     │ sends START  │     │ starts rec   │     │ begins       │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                                                                      │
                                                                      ▼
                                                              ┌──────────────┐
                                                              │ WAV buffer   │
                                                              │ → MP3 file   │
                                                              └──────────────┘
```

### 3. Transcription Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ User leaves  │────▶│ Extension    │────▶│ Daemon       │────▶│ Audio file   │
│ call         │     │ sends STOP   │     │ stops rec    │     │ finalized    │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                                                                      │
                                                                      ▼
                                                              ┌──────────────┐
                                                              │ whisper.cpp  │
                                                              │ transcribes  │
                                                              └──────────────┘
                                                                      │
                                                                      ▼
                                                              ┌──────────────┐
                                                              │ MD output    │
                                                              │ generated    │
                                                              └──────────────┘
```

---

## Component Specifications

### Chrome Extension

| Aspetto | Dettaglio |
|---------|-----------|
| Manifest Version | V3 (required) |
| Permissions | `nativeMessaging`, `tabs`, `storage` |
| Content Scripts | Meet detector, Teams detector v2 |
| Background | Service Worker |
| Communication | chrome.runtime.sendNativeMessage() |

#### Teams Web Detector v2 Architecture

The Teams Web detector uses a multi-signal scoring approach with state machine:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    TEAMS DETECTOR v2                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │  Selector    │───▶│  Evaluator   │───▶│  State Machine       │  │
│  │  Registry    │    │  (signals)   │    │  (phase transitions)  │  │
│  └──────────────┘    └──────────────┘    └──────────────────────┘  │
│         │                   │                      │                 │
│         ▼                   ▼                      ▼                 │
│  queryAny/All         collectSignals        idle/prejoin/         │
│  (fallback)           (multi-signal)        in_call/ending        │
│                                                                       │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │  Debounced   │    │  Smart       │    │  Structured          │  │
│  │  MutationObs │    │  Polling     │    │  Logging             │  │
│  └──────────────┘    └──────────────┘    └──────────────────────┘  │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

**Key Components:**

| Component | File | Purpose |
|-----------|------|---------|
| Selector Registry | `teams-selectors.ts` | Versioned DOM selectors with fallback |
| Evaluator | `teams-evaluator.ts` | Multi-signal scoring, confidence calculation |
| State Machine | `teams-evaluator.ts` | Phase transitions with stabilization |
| Detector | `teams.ts` | Orchestrates observation, polling, notifications |

**Signal Weights:**

| Signal | Weight | Description |
|--------|--------|-------------|
| `callContainer` | 0.25 | Primary call presence indicator |
| `callActive` | 0.20 | Teams internal call state |
| `hangupVisible` | 0.20 | Hangup button (G1 fix: = ACTIVE call) |
| `callControls` | 0.15 | Call control toolbar |
| `videoCount` | 0.10 | Active video elements |
| `audioCount` | 0.10 | Active audio elements |
| `prejoin` | -0.30 | Penalty when prejoin visible |

**State Machine Phases:**
- `idle` → `prejoin` → `in_call` → `ending` → `idle`
- START_THRESHOLD: 0.7 (need 70% confidence to enter `in_call`)
- END_THRESHOLD: 0.3 (need <30% confidence to enter `ending`)
- STABLE_MS: 2000 (must maintain threshold for 2s before transition)

### Native Host

| Aspetto | Dettaglio |
|---------|-----------|
| Language | Go 1.21+ |
| Binary | Single executable, no dependencies |
| Protocol | Chrome Native Messaging (length-prefixed JSON) |
| IPC | Unix domain socket to daemon |
| Install | ~/.config/google-chrome/NativeMessagingHosts/ |

### Core Daemon

| Aspetto | Dettaglio |
|---------|-----------|
| Language | Go 1.21+ |
| IPC | Unix domain socket |
| Audio (macOS) | ScreenCaptureKit via CGo |
| Audio (Linux) | PipeWire via FFmpeg (pulse) |
| Storage | SQLite + filesystem |
| Service | systemd (Linux), launchd (macOS) |

### STT Engine

| Aspetto | Dettaglio |
|---------|-----------|
| Engine | whisper.cpp |
| Model | ggml-small.bin (multilingual) |
| Memory | ~500MB with INT8 quantization |
| Speed | ~3x realtime on modern CPU |
| Invocation | Subprocess with JSON output |
| Languages | Multilingual (Italian, English, etc.) |

---

## File Formats

### Audio Output

```
~/verbalizer/recordings/2024-03-16_14-30-00_google-meet.mp3
```

- Format: MP3 (128kbps)
- Naming: `{date}_{time}_{platform}.mp3`

### Transcript Output

```markdown
---
title: "Google Meet - Project Review"
date: 2024-03-16T14:30:00+01:00
platform: google-meet
duration: 45:32
participants: 4
audio_file: ../recordings/2024-03-16_14-30-00_google-meet.mp3
---

# Transcript

## [00:00] Introduction
Speaker 1: Good afternoon everyone, let's start the project review...

## [02:15] Status Update
Speaker 2: The backend is now 80% complete...

## [15:30] Discussion
Speaker 1: What about the timeline?
Speaker 3: We're on track for the March deadline...
```

---

## Security Considerations

1. **Native Messaging**: Only accepts connections from extension with matching ID
2. **File Permissions**: Audio/transcripts stored in user home directory
3. **No Network**: Daemon doesn't expose network ports (except for OAuth callback)
4. **Local Processing**: All transcription happens locally, no cloud
5. **Cloud Sync**: Optional Google Drive backup with OAuth 2.0 + PKCE

---

## Cloud Sync Architecture (Google Drive)

### Overview

Cloud sync is implemented as an optional feature that backs up transcripts to the user's Google Drive. The sync happens automatically after transcript generation with robust retry logic.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CLOUD SYNC ARCHITECTURE                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                      CLOUD SYNC LAYER (Daemon)                      │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │                                                                       │   │
│   │   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │   │
│   │   │  OAuth Manager  │  │  Drive Client   │  │  Sync Queue     │    │   │
│   │   ├─────────────────┤  ├─────────────────┤  ├─────────────────┤    │   │
│   │   │ • PKCE flow     │  │ • files.create  │  │ • Outbox DB     │    │   │
│   │   │ • Token refresh  │  │ • Multipart     │  │ • Retry worker  │    │   │
│   │   │ • Revocation    │  │ • Folder ops    │  │ • Backoff       │    │   │
│   │   └─────────────────┘  └─────────────────┘  └─────────────────┘    │   │
│   │                                                                       │   │
│   │   ┌───────────────────────────────────────────────────────────────┐  │   │
│   │   │                    SECRET STORE                               │  │   │
│   │   ├───────────────────────────────────────────────────────────────┤  │   │
│   │   │   macOS: Keychain     │     Linux: Secret Service API        │  │   │
│   │   │   (fallback: encrypted file with machine-derived key)        │  │   │
│   │   └───────────────────────────────────────────────────────────────┘  │   │
│   │                                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                      │
│                                      │ HTTPS (OAuth API)                     │
│                                      ▼                                      │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                    GOOGLE CLOUD                                      │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │   • OAuth 2.0 (accounts.google.com)                                │   │
│   │   • Drive API (drive.google.com)                                    │   │
│   │   • Scope: drive.file (app-created files only)                      │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| OAuth Manager | `daemon/internal/auth/googleoauth/` | PKCE flow, token exchange, refresh, revocation |
| Drive Client | `daemon/internal/cloud/driveclient/` | Drive API operations (upload, folder listing) |
| Sync Queue | `daemon/internal/cloud/syncqueue/` | Job scheduling, retry with backoff, dead-letter handling |
| Secret Store | `daemon/internal/secrets/` | Secure credential storage (Keychain/Secret Service) |

### OAuth Flow

1. User initiates auth from extension UI
2. Daemon generates PKCE verifier/challenge + state
3. Browser opens Google auth URL with PKCE
4. User grants access, Google redirects to loopback
5. Daemon receives auth code, exchanges for tokens
6. Refresh token stored in OS secret store
7. Access token used for API calls, refreshed as needed

### Sync Queue States

| State | Description |
|-------|-------------|
| `pending` | Job created, waiting for worker |
| `uploading` | Currently uploading |
| `synced` | Successfully uploaded |
| `failed` | Temporary error, will retry |
| `permanent_failed` | Non-recoverable error |

### Retry Policy

- Base delay: 30 seconds
- Max delay: 2 hours
- Max attempts: 20
- Backoff: Exponential with jitter
- Retryable errors: 5xx, 429, network timeout
- Permanent failures: 401 (invalid grant), 403 (insufficient scope), 404 (folder deleted)

---

## Platform-Specific Notes

### macOS

- Requires "Screen Recording" permission for ScreenCaptureKit
- Install via launchd user agent (~/.config/verbalizer/)
- Notarization required for distribution

### Linux

- Requires PipeWire (standard on modern distros)
- Install via systemd user service
- Works with PulseAudio compatibility layer

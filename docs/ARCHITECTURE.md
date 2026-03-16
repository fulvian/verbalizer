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
| Content Scripts | Meet detector, Teams detector |
| Background | Service Worker |
| Communication | chrome.runtime.sendNativeMessage() |

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
| Audio (Linux) | PipeWire via pw-record subprocess |
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
3. **No Network**: Daemon doesn't expose network ports
4. **Local Processing**: All transcription happens locally, no cloud

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

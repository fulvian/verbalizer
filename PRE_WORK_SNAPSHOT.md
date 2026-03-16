# Verbalizer - Pre-Work Snapshot

**Data**: 2026-03-16
**Versione**: 1.0
**Stato**: In attesa di conferma utente

---

## 1. Obiettivo del Progetto

Sistema automatico per:
- Intercettare chiamate Google Meet e Microsoft Teams su Chrome
- Registrare l'audio automaticamente in background
- Trascrivere l'audio con modello STT locale
- Generare output in formato Markdown

**Piattaforme**: macOS + Linux
**Automazione**: Completa, nessun intervento utente richiesto

---

## 2. Architettura Scelta

**Opzione B: Native Host + System Audio**

| Piattaforma | Metodo | Scope Audio |
|-------------|--------|-------------|
| macOS | ScreenCaptureKit | System-wide (accettato) |
| Linux | PipeWire | Per-app (Chrome) |

---

## 3. Stack Tecnologico

| Componente | Tecnologia | Motivazione |
|------------|------------|-------------|
| Chrome Extension | TypeScript + Manifest V3 | Required per Chrome moderno |
| Native Host | Go 1.21+ | Singolo binario, cross-compile |
| Core Daemon | Go 1.21+ | Performance, basso overhead |
| STT Engine | whisper.cpp | C++ nativo, veloce su CPU |
| STT Model | ggml-small.bin | ~500MB, multilingua |
| Database | SQLite | Embedded, zero-config |
| Audio Format | MP3 (128kbps) | Compatto, universale |
| Output Format | Markdown + YAML | Leggibile, versionabile |

---

## 4. Struttura Progetto

```
verbalizer/
├── extension/                      # Chrome Extension
│   ├── src/
│   │   ├── content/
│   │   │   ├── index.ts           # Entry point
│   │   │   ├── detectors/
│   │   │   │   ├── meet.ts        # Google Meet detection
│   │   │   │   └── teams.ts       # MS Teams detection
│   │   │   └── observer.ts        # DOM state monitoring
│   │   ├── background/
│   │   │   ├── index.ts           # Service worker entry
│   │   │   └── native-bridge.ts   # Native messaging
│   │   └── types/
│   │       └── messages.ts        # Message types
│   ├── manifest.json
│   ├── package.json
│   ├── tsconfig.json
│   └── webpack.config.js          # Build config
│
├── native-host/                    # Native Messaging Host
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── messaging/
│   │   │   └── protocol.go        # Chrome NM protocol
│   │   └── ipc/
│   │       └── client.go          # Daemon communication
│   ├── go.mod
│   └── go.sum
│
├── daemon/                         # Core Service
│   ├── cmd/
│   │   └── verbalizerd/
│   │       └── main.go
│   ├── internal/
│   │   ├── audio/
│   │   │   ├── capture.go         # Interface
│   │   │   ├── capture_darwin.go  # macOS (ScreenCaptureKit)
│   │   │   └── capture_linux.go   # Linux (PipeWire)
│   │   ├── transcriber/
│   │   │   ├── whisper.go         # whisper.cpp wrapper
│   │   │   └── chunker.go         # Audio chunking
│   │   ├── storage/
│   │   │   ├── database.go        # SQLite operations
│   │   │   └── filesystem.go      # File management
│   │   ├── formatter/
│   │   │   ├── markdown.go        # MD generation
│   │   │   └── metadata.go        # YAML frontmatter
│   │   ├── ipc/
│   │   │   └── server.go          # Unix socket server
│   │   └── config/
│   │       └── config.go          # Configuration
│   ├── pkg/
│   │   └── api/
│   │       └── types.go           # Shared types
│   ├── go.mod
│   └── go.sum
│
├── whisper/                        # STT Engine
│   └── whisper.cpp/               # Git submodule
│
├── scripts/
│   ├── build.sh                   # Build all components
│   ├── install.sh                 # Linux installer
│   ├── install-macos.sh           # macOS installer
│   └── download-model.sh          # Download whisper model
│
├── test/
│   ├── extension/
│   ├── native-host/
│   └── daemon/
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── INSTALLATION.md
│   └── USAGE.md
│
├── Makefile
├── README.md
└── .gitignore
```

---

## 5. Piano di Implementazione

### FASE 1: Foundation (Setup)
- [ ] Inizializzazione repository git
- [ ] Setup struttura directory
- [ ] Makefile per build automation
- [ ] Configurazione whisper.cpp come submodule

### FASE 2: Chrome Extension
- [x] Manifest V3 con permessi necessari
- [x] Content script per rilevamento URL
- [x] Detector per Google Meet (DOM monitoring)
- [x] Detector per MS Teams (DOM monitoring)
- [x] Native messaging bridge
- [x] Build con webpack

### FASE 3: Native Host
- [ ] Protocollo Native Messaging (stdin/stdout)
- [ ] IPC client per daemon
- [ ] Installazione in Chrome NativeMessagingHosts

### FASE 4: Daemon Core
- [ ] Unix socket server
- [ ] Command dispatcher
- [ ] Session management

### FASE 5: Audio Capture
- [ ] Interfaccia AudioCapture
- [ ] Implementazione macOS (ScreenCaptureKit via CGo)
- [ ] Implementazione Linux (PipeWire via pw-record)
- [ ] Audio encoding (WAV → MP3)

### FASE 6: STT Integration
- [ ] Build whisper.cpp
- [ ] Download modello
- [ ] Wrapper Go per invocazione
- [ ] Audio chunking per file lunghi

### FASE 7: Output Generation
- [ ] Markdown formatter
- [ ] YAML frontmatter generation
- [ ] Metadata extraction

### FASE 8: Storage & Database
- [ ] SQLite schema
- [ ] File management
- [ ] Index per ricerca

### FASE 9: Service Installation
- [ ] systemd unit file (Linux)
- [ ] launchd plist (macOS)
- [ ] Installer scripts

### FASE 10: Testing & Polish
- [ ] Unit tests
- [ ] Integration tests
- [ ] E2E test con chiamate reali
- [ ] Documentation

---

## 6. Dipendenze Esterne

| Dipendenza | Versione | Scopo |
|------------|----------|-------|
| Go | 1.21+ | Native host + daemon |
| Node.js | 18+ | Extension build |
| whisper.cpp | latest | STT engine |
| FFmpeg | 5.0+ | Audio encoding (sistema) |
| PipeWire | 0.3+ | Audio capture Linux |

---

## 7. Rischi e Mitigazioni

| Rischio | Probabilità | Mitigazione |
|---------|-------------|-------------|
| DOM changes in Meet/Teams | Alta | Selector versioning, fallback detection |
| ScreenCaptureKit permissions | Media | Clear UX, permission guide |
| Whisper performance on low-end | Media | Chunking, async processing, model selection |
| PipeWire not available | Bassa | Fallback to PulseAudio |

---

## 8. Output Attesi

### Directory Utente

```
~/verbalizer/
├── recordings/
│   ├── 2026-03-16_09-30-00_google-meet.mp3
│   └── 2026-03-16_14-00-00_ms-teams.mp3
├── transcripts/
│   ├── 2026-03-16_09-30-00_google-meet.md
│   └── 2026-03-16_14-00-00_ms-teams.md
└── metadata.db
```

### Esempio Output Markdown

```markdown
---
title: "Google Meet - Project Sync"
date: 2026-03-16T09:30:00+01:00
platform: google-meet
duration: 32:15
audio_file: ../recordings/2026-03-16_09-30-00_google-meet.mp3
---

# Transcript

## [00:00] Introduction
Speaker 1: Good morning everyone...

## [05:30] Status Update
Speaker 2: The sprint is progressing well...
```

---

## 9. Conferma Richiesta

**Confermi questo piano per procedere con l'implementazione?**

- [ ] Architettura approvata
- [ ] Stack tecnologico approvato
- [ ] Struttura progetto approvata
- [ ] Piano di implementazione approvato

In caso di modifiche, indicare cosa cambiare.

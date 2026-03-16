# Prompt per Continuazione FASE 2 - Verbalizer

## Contesto del Progetto

**Progetto**: Verbalizer - Sistema automatico per registrazione, trascrizione e documentazione di chiamate Google Meet e Microsoft Teams.

**Repository GitHub**: https://github.com/fulvian/verbalizer

**Architettura scelta**: Native Host + System Audio
- macOS: ScreenCaptureKit (system-wide audio)
- Linux: PipeWire (per-app audio)

**Stack Tecnologico**:
| Componente | Tecnologia |
|------------|------------|
| Chrome Extension | TypeScript + Manifest V3 |
| Native Host | Go 1.21+ |
| Core Daemon | Go 1.21+ |
| STT Engine | whisper.cpp (git submodule) |
| Database | SQLite |
| Audio Format | MP3 (128kbps) |
| Output Format | Markdown + YAML |

---

## Stato Attuale (FASE 1 Completata)

La FASE 1 (Foundation) è stata completata con:
- ✅ Repository GitHub creato e pushato
- ✅ Git inizializzato con whisper.cpp come submodule
- ✅ Struttura directory completa creata
- ✅ Scripts di build e installazione
- ✅ Go modules per native-host e daemon
- ✅ Package.json, tsconfig, webpack per extension
- ✅ File placeholder per tutti i componenti
- ✅ Manifest V3 per Chrome extension

**Commit iniziale**: `6dbc01b` - "[orchestrator] FASE 1: Foundation complete"

---

## FASE 2: Chrome Extension - Specifiche

### Obiettivo
Implementare completamente la Chrome Extension con:
1. Content script per rilevamento URL
2. Detector per Google Meet (DOM monitoring)
3. Detector per MS Teams (DOM monitoring)
4. Background service worker completo
5. Native messaging bridge funzionante
6. Build system webpack configurato e testato

### Struttura Extension da Completare

```
extension/
├── src/
│   ├── content/
│   │   ├── index.ts           # Entry point content script
│   │   ├── observer.ts        # CallStateObserver class
│   │   └── detectors/
│   │       ├── meet.ts        # Google Meet detection
│   │       └── teams.ts       # MS Teams detection
│   ├── background/
│   │   ├── index.ts           # Service worker entry
│   │   └── native-bridge.ts   # Native messaging
│   └── types/
│       └── messages.ts        # TypeScript interfaces
├── manifest.json              # Manifest V3
├── package.json
├── tsconfig.json
├── webpack.config.js
└── jest.config.js
```

### Requisiti Funzionali

#### Content Script
- Iniettato automaticamente su `meet.google.com/*`, `teams.microsoft.com/*`, `teams.live.com/*`
- Rileva quando l'utente entra in una chiamata
- Rileva quando l'utente lascia una chiamata
- Estrae metadati (titolo meeting, partecipanti se possibile)
- Comunica con background via `chrome.runtime.sendMessage()`

#### Detector Google Meet
- Selettori DOM per:
  - Meeting room container
  - Indicatori chiamata attiva
  - Pannello partecipanti
  - Titolo meeting
- Gestione cambiamenti UI (Google può cambiare i selettori)

#### Detector MS Teams
- Selettori DOM per:
  - Call container
  - Indicatori chiamata attiva
  - Pannello partecipanti
  - Titolo meeting
- Supporto per teams.microsoft.com e teams.live.com

#### Background Service Worker
- Riceve messaggi dai content script
- Gestisce stato chiamate attive
- Comunica con native host via `chrome.runtime.sendNativeMessage()`
- Gestisce errori e retry

#### Native Bridge
- Invia comandi al native host:
  - `START_RECORDING` con payload: platform, callId, title
  - `STOP_RECORDING` con payload: callId
  - `GET_STATUS` per health check
- Gestisce risposte e errori

### Requisiti Tecnici

1. **TypeScript strict mode**: Nessun `any`, tipi espliciti
2. **Error handling**: Try-catch in tutte le operazioni asincrone
3. **Logging**: Console logs con prefisso `[Verbalizer]`
4. **Testing**: Unit tests con Jest, coverage > 80%
5. **Build**: Webpack production-ready

### Messaggi Types

```typescript
// Content -> Background
type: 'CALL_DETECTED'  | payload: { platform, url, title? }
type: 'CALL_STARTED'   | payload: { platform, callId, title?, participants? }
type: 'CALL_ENDED'     | payload: { platform, callId, duration }

// Background -> Native Host
type: 'START_RECORDING' | payload: { platform, callId, title? }
type: 'STOP_RECORDING'  | payload: { callId }
type: 'GET_STATUS'      | payload: {}
```

### Criteri di Accettazione

1. ✅ Extension si carica in Chrome senza errori
2. ✅ Content script iniettato su Meet e Teams
3. ✅ Rileva ingresso in chiamata Google Meet
4. ✅ Rileva uscita da chiamata Google Meet
5. ✅ Rileva ingresso in chiamata MS Teams
6. ✅ Rileva uscita da chiamata MS Teams
7. ✅ Background riceve e processa messaggi
8. ✅ Native bridge invia messaggi formattati correttamente
9. ✅ Build webpack completa senza errori
10. ✅ Unit tests passano con coverage > 80%

---

## Piano di Implementazione FASE 2

### Batch 1: Core Extension (parallelo)
| Subtask | Files | Descrizione |
|---------|-------|-------------|
| 2.1 | `src/types/messages.ts` | Tipi TypeScript completi |
| 2.2 | `src/content/observer.ts` | CallStateObserver class |
| 2.3 | `src/content/detectors/meet.ts` | Google Meet detector |
| 2.4 | `src/content/detectors/teams.ts` | MS Teams detector |

### Batch 2: Integration (seriale)
| Subtask | Files | Descrizione |
|---------|-------|-------------|
| 2.5 | `src/content/index.ts` | Content script entry point |
| 2.6 | `src/background/native-bridge.ts` | Native messaging bridge |
| 2.7 | `src/background/index.ts` | Background service worker |

### Batch 3: Build & Test
| Subtask | Files | Descrizione |
|---------|-------|-------------|
| 2.8 | `*.test.ts` | Unit tests |
| 2.9 | `webpack.config.js` | Build optimization |
| 2.10 | `manifest.json` | Final manifest |

---

## File Esistenti da Consultare

Prima di iniziare, leggi questi file per il contesto:
1. `docs/ARCHITECTURE.md` - Architettura completa del sistema
2. `PRE_WORK_SNAPSHOT.md` - Piano di implementazione completo
3. `orchestration_log.md` - Log di avanzamento
4. `extension/manifest.json` - Manifest V3 corrente
5. `extension/src/types/messages.ts` - Tipi esistenti
6. `extension/src/content/detectors/meet.ts` - Detector esistente
7. `extension/src/content/detectors/teams.ts` - Detector esistente

---

## Comandi Utili

```bash
# Clona il repository
git clone https://github.com/fulvian/verbalizer.git
cd verbalizer

# Installa dipendenze extension
cd extension && npm install

# Build development
npm run dev

# Build production
npm run build

# Run tests
npm test

# Type check
npm run typecheck
```

---

## Istruzioni per l'Agente

1. Leggi tutti i file di contesto elencati sopra
2. Segui il protocollo UMP per la pianificazione
3. Implementa i subtask nell'ordine dei batch
4. Scrivi test prima del codice (TDD)
5. Aggiorna `orchestration_log.md` dopo ogni subtask completato
6. Fai commit dopo ogni batch completato
7. Push su GitHub alla fine della FASE 2

---

## Prompt per Avviare la Sessione

```
Continua lo sviluppo del progetto Verbalizer.

Repository: https://github.com/fulvian/verbalizer

FASE 1 (Foundation) è completata. Procedi con la FASE 2: Chrome Extension.

Leggi i file di contesto:
- PRE_WORK_SNAPSHOT.md
- docs/ARCHITECTURE.md
- orchestration_log.md
- extension/manifest.json
- extension/src/types/messages.ts

Obiettivo FASE 2: Implementare completamente la Chrome Extension con:
1. Content script per rilevamento URL Meet/Teams
2. Detector per Google Meet (DOM monitoring)
3. Detector per MS Teams (DOM monitoring)
4. Background service worker completo
5. Native messaging bridge funzionante
6. Build webpack e test con coverage > 80%

Segui TDD: scrivi test prima del codice.
Usa TypeScript strict mode.
Aggiorna orchestration_log.md dopo ogni subtask.
Fai commit dopo ogni batch completato.
```

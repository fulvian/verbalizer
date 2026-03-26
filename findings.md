# Findings — Google Drive OAuth Planning

## Repository & Runtime Findings
- Architettura attuale a 3 livelli: Extension (TS MV3) → Native Host (Go) → Daemon (Go).
- Trigger start/stop recording già funzionante per Meet e in test su Teams.
- Trascrizione avviene in background (`session.Manager.StopRecording` -> goroutine).
- Persistenza attuale locale: Markdown in `cfg.TranscriptsDir` + SQLite (`sessions`).

## Integration Points Identified
- **Daemon**: punto principale per upload cloud post-generazione transcript (`manager.go` dopo write markdown + DB update).
- **Config**: `internal/config/config.go` non carica ancora YAML (TODO). Va esteso per sezione cloud sync.
- **Storage DB**: schema minimale (`sessions`). Necessarie tabelle/colonne sync status.
- **API/IPC**: oggi comandi `START_RECORDING`, `STOP_RECORDING`, `GET_STATUS`; servono comandi auth/sync.
- **Extension**: manca UI impostazioni OAuth/stato cartella Drive.
- **Installer**: crea config base; dovrà includere placeholder cloud settings e percorso secret storage.

## Code Quality / Technical Debt Relevant
- `capture_*` usa dir hardcoded `recordings` in alcuni punti (non sempre `cfg.RecordingsDir`).
- `config.Load()` non implementato: blocca configurabilità reale senza intervento.
- Alcuni placeholder/assunzioni in audio capture macOS.
- Stato pipeline transcript non ha coda resiliente per sync esterna.

## Git/History Findings
- Evoluzione per fasi già strutturata (fase-1..fase-10).
- Ultimo commit recente ha aggiornato audio Linux + docs.
- Branch unico `main`, repo ahead di 1 commit, untracked `recordings/`.

## Google Docs Findings (official)
- OAuth desktop: loopback redirect `127.0.0.1`/`::1`, PKCE S256 raccomandato, custom schemes/OOB deprecati per casi rilevanti.
- Token endpoint: `oauth2.googleapis.com/token`; revoke: `oauth2.googleapis.com/revoke`.
- Scope Drive consigliata: `drive.file` (least privilege), `drive.appdata` per dati app privati.
- Upload: `files.create` con `multipart` per file piccoli + metadata; `resumable` consigliato per robustezza generale.

## Teams Web Findings (initial codebase)
- Teams è supportato a livello routing/manifest/types/pipeline (`ms-teams` end-to-end).
- Detector Teams corrente (`extension/src/content/detectors/teams.ts`) usa polling + MutationObserver ma con euristiche fragili.
- Bug logico critico: presenza `hangup-button` interpretata come call ended (in pratica spesso è indicatore call attiva).
- Selettori Teams hardcoded (`data-tid`/classi) senza fallback/versioning strutturato.
- Manca state machine esplicita con stabilizzazione temporale anti-flap.
- Test extension passano (84/84) ma coprono soprattutto DOM sintetico (jsdom), non scenario Teams reale.
- Best practice 2026 per detector robusti: multi-signal evaluation, debounce MutationObserver, hysteresis, event idempotency.

# Verbalizer — Piano di Implementazione Fase Nuova
## Salvataggio trascrizioni in Google Drive con OAuth 2.0 (profilo personale utente)

**Data:** 2026-03-25  
**Stato:** ✅ COMPLETATO (tutte le milestone completate)
**Ultimo commit:** `afd06bc` (Milestone E - Hardening)

---

## 1) Obiettivo e criteri di successo

### Obiettivo
Consentire a Verbalizer di salvare automaticamente le trascrizioni `.md` in una directory Google Drive scelta dall’utente, autenticandosi via OAuth 2.0 sul profilo Google personale dell’utente, mantenendo la pipeline locale esistente.

### Criteri di successo (business + tecnici)
1. L’utente connette il proprio account Google una sola volta (salvo revoca/scadenza).
2. L’utente seleziona una cartella Drive target.
3. Ogni trascrizione completata viene caricata in Drive con retry robusti.
4. In assenza rete/API, il sistema non perde file: queue locale + retry.
5. Stato sync osservabile (pending/success/failed) e tracciabile nel DB.
6. Sicurezza: token protetti localmente, scope minimi necessari.

---

## 2) Stato attuale della codebase (as-is)

## 2.1 Architettura runtime
- `extension/` (TypeScript, MV3): rilevamento call Meet/Teams, messaggi al native host.
- `native-host/` (Go): bridge NM stdin/stdout ↔ daemon su Unix socket.
- `daemon/` (Go): session management, capture audio, trascrizione whisper.cpp, output markdown, SQLite.

## 2.2 Flusso già operativo
1. Extension invia `CALL_STARTED` → `START_RECORDING`.
2. Daemon registra audio.
3. Extension invia `CALL_ENDED` → `STOP_RECORDING`.
4. Daemon finalizza audio, trascrive in background, genera `.md` in `TranscriptsDir`, aggiorna DB sessione.

## 2.3 Evidenze tecniche rilevanti
- Punto di integrazione naturale upload: `daemon/internal/session/manager.go` subito dopo salvataggio markdown e update DB.
- Config attuale incompleta: `config.Load()` è TODO (no parse YAML reale).
- DB attuale (`sessions`) non contiene stato sync cloud.
- IPC espone solo `START_RECORDING`, `STOP_RECORDING`, `GET_STATUS`.
- Extension non ha UI settings OAuth/cloud.

---

## 3) Requisiti della nuova fase (to-be)

## 3.1 Funzionali
1. Connessione account Google (OAuth desktop).
2. Selezione cartella Google Drive destinazione.
3. Upload automatico di ogni transcript `.md` completato.
4. Retry automatici con backoff in caso errori transitori.
5. Visibilità stato sincronizzazione per sessione.
6. Disconnessione account (revoca token + cleanup locale credenziali).

## 3.2 Non funzionali
- **Affidabilità:** nessuna perdita transcript in caso offline.
- **Sicurezza:** cifratura/secret-store locale per refresh token.
- **Minimo privilegio:** scope OAuth non-restricted quando possibile.
- **Osservabilità:** log strutturati e stato sync interrogabile.
- **Compatibilità:** Linux + macOS.

---

## 4) Analisi documentazione ufficiale Google (vincoli e best practice)

## 4.1 OAuth desktop
- Redirect raccomandato: **loopback IP** `http://127.0.0.1:{port}` o `http://[::1]:{port}`.
- PKCE raccomandato con `S256`.
- `state` raccomandato per protezione CSRF.
- Endpoint:
  - Authorization: `https://accounts.google.com/o/oauth2/v2/auth`
  - Token: `https://oauth2.googleapis.com/token`
  - Revoke: `https://oauth2.googleapis.com/revoke`

## 4.2 Scope Drive
- Preferenza per `https://www.googleapis.com/auth/drive.file` (least privilege, non-sensitive).
- `drive` pieno è restricted e aumenta impatto compliance/verification.
- `drive.appdata` utile solo per dati app privati, non per cartella utente visibile.

## 4.3 Upload Drive
- Per file `.md` piccoli con metadata: `files.create` + `uploadType=multipart`.
- Per robustezza generale e resume: `uploadType=resumable` consigliato.
- Gestione errori:
  - `5xx` retry/resume
  - `403 rateLimitExceeded` backoff
  - `4xx` in resumable spesso richiede nuova session URI

---

## 5) Gap analysis

| Area | Stato attuale | Gap | Impatto |
|------|---------------|-----|---------|
| OAuth flow | Assente | Manca autenticazione Google | Bloccante |
| Token storage | Assente | Manca secure credential store | Bloccante |
| Folder selection | Assente | Manca selezione cartella Drive | Alto |
| Upload engine | Assente | Manca client Drive + retry/outbox | Bloccante |
| DB schema | Solo sessions | Manca tracking sync cloud | Alto |
| Config | Parsing YAML TODO | Manca config cloud persistente | Alto |
| IPC/API | Comandi minimi | Mancano comandi auth/status cloud | Medio |
| Extension UX | Nessuna UI cloud | Manca setup guidato utente | Medio |

---

## 6) Proposta architetturale target

## 6.1 Decisione di posizionamento
**OAuth e upload nel daemon** (non nella extension).

Motivazioni:
- daemon è processo persistente e già proprietario del lifecycle transcript.
- evita dipendenza dal lifecycle effimero service-worker MV3.
- centralizza segreti/token in componente Go nativa, più idonea a secure storage OS.

## 6.2 Nuovi sottosistemi (daemon)
1. **auth/googleoauth**
   - genera PKCE verifier/challenge
   - avvia listener loopback temporaneo
   - apre browser di sistema URL auth
   - scambia code→token, refresh token, revoca

2. **cloud/driveclient**
   - `files.create` (multipart/resumable)
   - refresh access token automatico
   - mapping errori in categorie retryable/non-retryable

3. **cloud/syncqueue**
   - outbox locale persistente (DB)
   - worker retry con exponential backoff + jitter
   - dead-letter status per errori permanenti

4. **secrets/store**
   - macOS: Keychain
   - Linux: Secret Service (fallback cifrato file con key locale, se necessario)

## 6.3 Estensioni IPC/API
Nuovi comandi proposti (native-host ↔ daemon):
- `GOOGLE_AUTH_START`
- `GOOGLE_AUTH_STATUS`
- `GOOGLE_AUTH_DISCONNECT`
- `GOOGLE_DRIVE_SET_FOLDER`
- `GOOGLE_DRIVE_GET_FOLDER`
- `GOOGLE_DRIVE_SYNC_STATUS`
- `GOOGLE_DRIVE_SYNC_RETRY`

## 6.4 Estensioni extension
- Pagina options/settings minima per:
  - Connect/Disconnect Google
  - Selezione/visualizzazione cartella destinazione
  - stato ultimo sync

---

## 7) Data model proposto

## 7.1 Tabelle nuove

### `cloud_accounts`
- `provider` (google)
- `account_email`
- `scopes`
- `connected_at`
- `status`

### `cloud_sync_jobs`
- `id`
- `session_call_id`
- `local_transcript_path`
- `provider`
- `target_folder_id`
- `state` (`pending|uploading|synced|failed|permanent_failed`)
- `attempt_count`
- `next_retry_at`
- `last_error_code`
- `last_error_message`
- `remote_file_id`
- `created_at`, `updated_at`

## 7.2 Estensione `sessions`
- `cloud_sync_state`
- `cloud_remote_file_id`
- `cloud_last_sync_at`

---

## 8) Strategia OAuth dettagliata

1. Daemon genera `state` + PKCE (`verifier`, `challenge S256`).
2. Daemon apre browser su endpoint auth con scope `drive.file`.
3. Listener loopback riceve `code`.
4. Daemon valida `state`, scambia code su token endpoint.
5. Daemon salva refresh token in secret store OS.
6. Access token in memoria (o cache temporanea con expiry).
7. Su scadenza token: refresh automatico.
8. Disconnect: revoke + wipe secret + invalidazione account in DB.

### Error handling OAuth
- `redirect_uri_mismatch`: configurazione client OAuth errata.
- `invalid_grant`: refresh token revocato/scaduto → richiede reconnect.
- `access_denied`: utente nega consenso.

---

## 9) Strategia upload e sync affidabile

## 9.1 Trigger
In `session.Manager` dopo write markdown riuscita:
- inserire job `pending` in `cloud_sync_jobs`.

## 9.2 Worker
- Goroutine dedicata con polling leggero.
- Prende job `pending` o `failed` con `next_retry_at <= now`.
- Esegue upload.
- Aggiorna stato + `remote_file_id`.

## 9.3 Upload mode
- Default v1: `multipart` (file piccoli .md).
- Feature flag per `resumable` in v1.1 se richiesto.

## 9.4 Idempotenza
- nome file deterministico: `{timestamp}_{platform}_{callId}.md`.
- prima di creare, opzionale query per nome/callId property (fase successiva ottimizzazione).

## 9.5 Retry policy
- Errori retryable: network timeout, 5xx, 429/403 rate limit.
- Backoff: esponenziale con jitter (es. 30s, 2m, 10m, 30m, 2h max).
- Errori permanenti: 401/invalid grant, 404 folder non accessibile, scope insufficienti.

---

## 10) Configurazione proposta

```yaml
cloud:
  provider: google_drive
  enabled: true
  oauth_client_id: "..."
  oauth_redirect_host: "127.0.0.1"
  oauth_redirect_port_range: "49152-65535"
  scope: "https://www.googleapis.com/auth/drive.file"
  target_folder_id: ""
  upload_mode: "multipart" # multipart|resumable
  retry:
    max_attempts: 20
    base_delay_seconds: 30
    max_delay_seconds: 7200
```

Nota: `config.Load()` va implementato prima di attivare realmente questa configurazione.

---

## 11) Sicurezza e compliance

1. Least privilege scope (`drive.file`).
2. Refresh token mai in chiaro su file di log/config.
3. Secret store nativo OS.
4. Redaction log per token/headers auth.
5. Supporto revoca utente e cleanup locale immediato.
6. Documento privacy aggiornato: ora c’è trasferimento cloud opzionale.

---

## 12) Piano implementativo per milestone

## Milestone A — Fondazione cloud (backend only)
- Implementare `config.Load()` + sezione `cloud`.
- Migrazioni DB (`cloud_accounts`, `cloud_sync_jobs`, estensione `sessions`).
- Strato interfacce `AuthProvider`, `CloudUploader`, `SyncQueue`.

## Milestone B — OAuth daemon
- Implementazione OAuth desktop loopback + PKCE.
- Secret storage macOS/Linux.
- Comandi IPC minimi: start/status/disconnect auth.

## Milestone C — Upload pipeline
- Drive client `files.create` multipart.
- Worker queue + retry/backoff.
- Hook in `session.Manager` post-transcript.

## Milestone D — UX extension
- Options page: connect/disconnect/set folder/status.
- Messaging background↔native host per nuovi comandi.

## Milestone E — Hardening
- Error taxonomy completa.
- Telemetria/log operativi.
- Test end-to-end manuali su Meet/Teams.

---

## 13) Test plan dettagliato

## 13.1 Unit test
- PKCE generation/validation.
- OAuth state validation.
- Token refresh logic.
- Drive upload request builder.
- Retry scheduler/backoff.

## 13.2 Integration test
- Mock OAuth endpoints (auth/token/revoke).
- Mock Drive API responses (200, 401, 403, 429, 5xx).
- DB migration + job lifecycle.

## 13.3 E2E test manuale
1. Connect account Google.
2. Selezione cartella Drive.
3. Join/leave Meet → verifica file in Drive.
4. Disconnessione rete durante upload → verifica retry e success post-rete.
5. Revoke token su account.google.com → verifica richiesta reconnect.

## 13.4 Criteri di accettazione
- 100% transcript generate localmente anche con cloud down.
- >=99% upload success entro finestra retry in condizioni rete normali.
- Zero leak di token in log.

---

## 14) Rischi e mitigazioni

| Rischio | Probabilità | Impatto | Mitigazione |
|--------|-------------|---------|-------------|
| OAuth setup complesso per utente | Media | Alta | Wizard guidato + doc setup OAuth client |
| Token storage su Linux frammentato | Media | Media | Secret Service + fallback cifrato |
| Rate limiting Drive API | Media | Media | Backoff + queue persistente |
| Regressione pipeline transcript | Bassa | Alta | Hook non bloccante + fallback locale invariato |
| Scope verification Google | Media | Alta | Usare `drive.file` non-sensitive |

---

## 15) Stima effort (ordine di grandezza)

- Milestone A: 2–3 gg
- Milestone B: 3–4 gg
- Milestone C: 3–4 gg
- Milestone D: 2–3 gg
- Milestone E: 2–3 gg

**Totale stimato:** 12–17 giorni lavorativi (1 engineer full-time), esclusi tempi esterni di eventuale verifica OAuth app.

---

## 16) Piano di rilascio suggerito

1. **Alpha locale** (feature flag cloud off di default).
2. **Beta interna** con utenti test e cartelle dedicate.
3. **GA** con cloud on-by-default solo dopo stabilità e UX confermata.

Rollback semplice: disattivare `cloud.enabled` mantenendo pipeline locale identica.

---

## 17) Decisioni richieste in approvazione formale

1. Conferma uso scope `drive.file` come default.
2. Conferma OAuth gestito nel daemon (non extension).
3. Conferma upload strategy iniziale `multipart` + queue retry.
4. Conferma introduzione settings UI in extension (pagina options).
5. Conferma milestone e priorità indicate.

---

## 18) Punti di integrazione concreti (file-level map)

### Da estendere (priorità alta)
- `daemon/internal/session/manager.go` (enqueue sync job post transcript)
- `daemon/internal/config/config.go` (load YAML + cloud block)
- `daemon/internal/storage/database.go` (+ migrazioni/schema sync)
- `daemon/pkg/api/types.go` (nuovi command/response types)
- `daemon/cmd/verbalizerd/handler.go` (nuovi handler auth/sync)
- `native-host/cmd/main.go` (forward nuovi comandi)
- `extension/src/background/native-bridge.ts` (metodi auth/sync)
- `extension/src/background/index.ts` (routing nuovi messaggi)

### Nuovi moduli proposti
- `daemon/internal/auth/googleoauth/*`
- `daemon/internal/cloud/driveclient/*`
- `daemon/internal/cloud/syncqueue/*`
- `daemon/internal/secrets/*`
- `extension/src/options/*` (UI configurazione)

---

## 19) Conclusione

La codebase attuale è adatta ad aggiungere sincronizzazione Google Drive senza stravolgere l’architettura. L’integrazione va costruita come estensione del daemon con OAuth desktop PKCE, queue persistente e upload robusto. Il piano sopra minimizza rischio regressioni, mantiene privacy-by-default (trascrizione locale invariata), e aggiunge cloud sync come capacità opzionale e controllata dall’utente.

# Verbalizer — Piano di implementazione completo
## Remediation Teams Web su Linux (rilevazione + registrazione automatica)

**Data:** 2026-03-26  
**Stato:** Piano esecutivo (nessuna modifica codice in questa fase)  
**Target:** rendere il sistema affidabile sul PC Linux dell’utente

---

## 1) Obiettivo e Definition of Done

## Obiettivo
Rendere affidabile e deterministico il flusso:

Teams call (web) → detector stabile → CALL_STARTED unico → registrazione audio corretta → CALL_ENDED unico → trascrizione Markdown generata.

## Definition of Done (DoD)

Il sistema è considerato “completamente funzionante” quando tutti i seguenti criteri sono soddisfatti:

1. **Rilevazione call affidabile** in 3 scenari: join diretto, join da prejoin, reconnect.
2. **Eventi idempotenti**: max 1 `CALL_STARTED` e 1 `CALL_ENDED` per sessione.
3. **Audio utile**: file registrato contiene audio call Teams (non solo microfono/silenzio).
4. **Trascrizione prodotta** in `TranscriptsDir` senza errori path/model.
5. **Tracciabilità completa**: ogni call ha correlation-id e timeline cross-layer.
6. **Suite test**: unit + integration + smoke real browser pass.

---

## 2) Strategia di esecuzione

Approccio in 6 fasi, con gate di qualità tra una fase e l’altra.

- **Fase 0**: Baseline e osservabilità minima
- **Fase 1**: Hardening detector Teams
- **Fase 2**: Hardening lifecycle eventi
- **Fase 3**: Hardening audio capture Linux
- **Fase 4**: Hardening daemon runtime/transcription
- **Fase 5**: Verifica end-to-end e runbook operativo

---

## 3) Piano dettagliato per fase

## Fase 0 — Baseline e diagnostica iniziale (P0)

### Obiettivo
Stabilire una baseline riproducibile prima delle modifiche.

### Task
1. Definire schema log strutturato con campi minimi:
   - `ts`, `layer` (content/background/native/daemon), `platform`, `callId`, `event`, `state`, `confidence`, `reason`, `errorCode`.
2. Aggiungere log di correlazione a ogni hop.
3. Preparare checklist test manuale Teams Linux (join/prejoin/reconnect/end).
4. Congelare baseline test esistente (snapshot risultati).

### Acceptance
- Esiste una timeline unificata per una call completa.
- Ogni errore è attribuibile a un layer specifico in < 2 minuti.

---

## Fase 1 — Hardening detector Teams (P0)

### Obiettivo
Eliminare falsi trigger e rendere robusta la classificazione dello stato call.

### Task
1. **Ricalibrare scoring/thresholds** (approccio conservativo):
   - introdurre soglie separate per start/end con hysteresis ampia,
   - evitare start con singolo segnale debole.
2. **Correggere precedence phase evaluation**:
   - prejoin deve dominare quando i segnali call non sono sufficienti,
   - ending/idle gestiti da perdita segnali persistente, non istantanea.
3. **Refactor `isElementVisible` production-safe**:
   - rimuovere shortcut “data-tid => visible”,
   - usare check robusti (display/visibility/opacity/size/intersection/aria-hidden).
4. **Pulizia selector set**:
   - rimuovere selettori invalidi,
   - classificare selector per affidabilità (high/medium/low),
   - mantenere registry versionato con fallback ordinato.
5. **Dedup su query multiple** (per video/audio count).

### Acceptance
- Nessun `CALL_STARTED` con solo segnali deboli.
- Nessun false end in call stabile > 30 min.
- Test matrice segnali/fasi completo e verde.

---

## Fase 2 — Hardening lifecycle eventi extension/background (P0)

### Obiettivo
Garantire determinismo e idempotenza eventi start/stop.

### Task
1. Introdurre macchina stati evento in content detector (`idle/prejoin/in_call/ending`).
2. Applicare dedupe temporal + semantic (stesso callId/stesso evento).
3. In background, aggiungere guardie contro doppi `START_RECORDING`/`STOP_RECORDING` ravvicinati.
4. Standardizzare handling errori `sendNativeMessage` con codici espliciti:
   - `NATIVE_HOST_UNREACHABLE`,
   - `DAEMON_UNREACHABLE`,
   - `DAEMON_REJECTED`.
5. Surface stato diagnostico in extension (console + opzionale badge/diagnostic panel).

### Acceptance
- Un solo start/stop per call anche con burst mutazioni DOM.
- Errori native host leggibili e classificati.

---

## Fase 3 — Hardening audio capture Linux (P0 bloccante)

### Obiettivo
Registrare effettivamente l’audio della call Teams nel tuo ambiente Linux.

### Task
1. Implementare **source discovery** PipeWire/Pulse:
   - enumerazione sources/sinks/monitor,
   - selezione automatica monitor preferito + fallback configurabile.
2. Sostituire `-i default` con input deterministico (monitor selected source).
3. Aggiungere preflight check capture:
   - verifica ffmpeg,
   - verifica source disponibile,
   - test breve VU/energia segnale.
4. Gestire gracefully start/stop process (interrupt/timeout/kill fallback) con logging esito.
5. Correggere messaggi errore fuorvianti.

### Acceptance
- In 10 chiamate test, 10/10 file con audio call udibile.
- Nessun “successo” con file vuoto/silenzio non rilevato.

---

## Fase 4 — Hardening daemon runtime/transcription (P0)

### Obiettivo
Rendere affidabile produzione transcript indipendentemente dalla cwd del servizio.

### Task
1. Eliminare path relativi hardcoded per whisper binary/model.
2. Rendere path completamente config-driven e verificati in startup.
3. Correggere `RecordingPath` di risposta start (niente placeholder).
4. Validare presenza modello/binario all’avvio daemon con errore early-fail chiaro.
5. Uniformare directory output (`RecordingsDir`, `TranscriptsDir`) e ownership permessi.

### Acceptance
- Daemon startup fail-fast se prerequisiti mancanti.
- Transcript generato sempre nel path configurato.

---

## Fase 5 — QA finale, runbook, hardening operativo (P1)

### Obiettivo
Chiudere il ciclo con validazione reale e playbook operativo.

### Task
1. Test plan real-browser Teams:
   - tenant work/school,
   - tenant personal,
   - call con condivisione schermo,
   - reconnect rete breve.
2. Test regressione Meet (no break).
3. Runbook di troubleshooting per Linux:
   - detector fail,
   - native messaging fail,
   - daemon/capture fail,
   - transcription fail.
4. Aggiornamento `docs/INSTALLATION.md` con sezione Linux capture tuning.

### Acceptance
- Checklist QA firmata con evidenze.
- MTTR diagnostico < 10 minuti su errori comuni.

---

## 4) Work breakdown structure (WBS) sintetica

1. **Telemetry contract** (cross-layer JSON logs)
2. **Detector reliability** (scoring + visibility + selector hygiene)
3. **Event idempotency** (content/background)
4. **Linux capture reliability** (source discovery + deterministic input)
5. **Daemon runtime correctness** (paths/config/fail-fast)
6. **E2E QA + docs**

---

## 5) Piano test completo

## A) Unit tests
- evaluator signals→confidence→phase matrix
- visibility helper con casi hidden/aria/zero-size
- selector parser safety + invalid selector handling

## B) Integration tests (extension)
- sequenze temporali: idle→prejoin→in_call→ending→idle
- mutation burst + debounce
- dedupe start/stop

## C) Integration tests (daemon)
- start/stop con callId corretto/errato
- unavailable source
- missing model/binario

## D) Smoke/E2E manuali Linux
- 10 chiamate Teams reali
- verifica audio registrato + transcript generato

---

## 6) Rischi residui e mitigazioni

| Rischio | Probabilità | Impatto | Mitigazione |
|---|---:|---:|---|
| UI Teams cambia DOM | Alta | Alta | selector registry versionato + confidence multi-signal |
| Source audio cambia a runtime | Media | Alta | source re-discovery + fallback automatico |
| Race start/stop su burst eventi | Media | Alta | idempotenza e state lock per callId |
| Regressione Meet | Media | Media | test regressione obbligatori in CI |

---

## 7) Deliverable finali previsti

1. Detector Teams hardened e testato.
2. Pipeline eventi deterministica con error taxonomy.
3. Capture Linux robusta e configurabile.
4. Daemon runtime allineato a config assoluta.
5. Runbook troubleshooting Linux + Teams.
6. Report QA finale con evidenze E2E.

---

## 8) Sequenza di implementazione raccomandata (ordine esecuzione)

1. **Fase 0** (strumentazione)
2. **Fase 3** (audio Linux bloccante)
3. **Fase 1** (detector Teams)
4. **Fase 2** (idempotenza eventi)
5. **Fase 4** (runtime/transcription)
6. **Fase 5** (QA + docs)

Motivo: massimizza capacità diagnostica e rimuove prima il blocker audio, evitando falsi positivi durante il tuning detector.

---

## 9) Criterio di rilascio sul tuo PC

Rilascio accettato solo se:

- 10/10 call Teams test con start/stop corretto,
- 10/10 file audio con contenuto utile,
- 10/10 transcript generati,
- 0 incidenti di doppio start/stop,
- log cross-layer completi e coerenti per ogni callId.

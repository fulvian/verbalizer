# Verbalizer — Microsoft Teams su Linux (Chrome Extension)
## Analisi approfondita delle criticità (rilevazione call + registrazione automatica)

**Data:** 2026-03-26  
**Ambiente target:** Linux desktop, Chrome extension MV3, Microsoft Teams Web  
**Scope:** analisi tecnica senza modifiche codice runtime

---

## 1) Executive summary

La codebase contiene già un detector Teams v2 e una pipeline end-to-end (content script → background → native host → daemon). Tuttavia lo stato attuale presenta **criticità strutturali e implementative** che spiegano perché la rilevazione/registrazione automatica su Teams Linux può risultare non funzionante o intermittente.

Le cause principali individuate:

1. **Fragilità del detector Teams lato segnali DOM**, con soglie e regole che possono produrre falsi positivi/falsi negativi.
2. **Heuristics di visibilità non affidabili in produzione** (`isElementVisible`) che possono trattare elementi nascosti come “visibili”.
3. **Assunzioni deboli nella cattura audio Linux** (`ffmpeg -f pulse -i default`) che spesso non garantiscono audio della call Teams.
4. **Debolezza architetturale osservability/debug end-to-end**: errori di bridge/daemon non propagati in modo operativo per diagnosi rapida.
5. **Mismatch tra path/config runtime del daemon e contesto reale installato** (recordings/transcriber paths hardcoded o relativi).

Questa combinazione rende il sistema vulnerabile soprattutto su Teams Web (DOM mutabile + audio routing Linux variabile).

---

## 2) Nota sui log forniti

Nel materiale disponibile in workspace non è presente il dump testuale dei log console della sessione Teams citata.  
Di conseguenza, l’analisi è stata effettuata in modalità:

- **code-driven root cause analysis** (sorgenti extension/native-host/daemon),
- **consistency analysis** con i runbook/documenti già presenti,
- **failure mode inference** sui punti critici della pipeline.

Impatto: le conclusioni sono tecnicamente solide ma non “fingerprintate” su una specifica timeline di log runtime della tua sessione.

---

## 3) Perimetro analizzato

### Extension
- `extension/manifest.json`
- `extension/src/content/index.ts`
- `extension/src/content/observer.ts`
- `extension/src/content/detectors/teams.ts`
- `extension/src/content/detectors/teams-evaluator.ts`
- `extension/src/content/detectors/teams-selectors.ts`
- test relativi `teams*.test.ts`

### Bridge/Backend
- `extension/src/background/index.ts`
- `extension/src/background/native-bridge.ts`
- `native-host/cmd/main.go`
- `native-host/internal/ipc/client.go`
- `daemon/cmd/verbalizerd/handler.go`
- `daemon/internal/session/manager.go`
- `daemon/internal/audio/capture_linux.go`
- `daemon/pkg/api/types.go`

### Docs/runbook
- `docs/ARCHITECTURE.md`
- `docs/INSTALLATION.md`
- `docs/runbooks/teams-selector-hotfix.md`

---

## 4) Criticità principali (prioritizzate)

## C1 — **Detector Teams: soglia START troppo bassa e semantica di fase ambigua**
**File:** `teams-evaluator.ts`  
**Evidenza:** `START_THRESHOLD = 0.25`, `END_THRESHOLD = 0.10`, logica `evaluatePhase` con branch non perfettamente coerenti.

### Impatto
- Possibili start prematuri in condizioni non realmente “in_call”.
- Possibili transizioni spurie (`ending` vs `idle` vs `in_call`) in fasi di prejoin/reattach DOM.

### Gravità
**Alta** (affidabilità core della rilevazione).

---

## C2 — **`isElementVisible` non production-safe (falso positivo strutturale)**
**File:** `teams-selectors.ts`  
**Evidenza:** per elementi con `data-tid` (non video/audio) ritorna `true` quasi immediatamente, bypassando visibilità reale.

### Impatto
- Elementi presenti ma hidden/stale possono essere interpretati come segnali attivi.
- Incremento di confidenza non meritato → trigger CALL_START non affidabile.

### Gravità
**Alta**.

---

## C3 — **Strategia selettori Teams ancora fragile nonostante registry**
**File:** `teams-selectors.ts`  
**Evidenza:** alcuni selettori generici e uno invalido (`[class*="call-"]-container`, ignorato silenziosamente).

### Impatto
- Fragilità a variazioni UI Teams.
- Ridotta capacità di diagnosi perché selector invalid viene silenziato.

### Gravità
**Media-Alta**.

---

## C4 — **Gap osservability end-to-end tra detector e registrazione**
**File:** `teams.ts`, `background/index.ts`, `native-bridge.ts`  
**Evidenza:** logging presente ma non normalizzato su correlation-id cross-layer; esiti errore native host non elevati a stato operativo consistente.

### Impatto
- Difficile distinguere: detector fail vs native host fail vs daemon fail.
- Debug in campo lento e non deterministico.

### Gravità
**Alta operativa**.

---

## C5 — **Cattura audio Linux non robusta per Teams reale**
**File:** `daemon/internal/audio/capture_linux.go`  
**Evidenza:** input fisso `ffmpeg -f pulse -i default`.

### Impatto
- Su Linux `default` spesso punta al microfono o a source non monitor corretto.
- Possibile “registrazione riuscita” ma file senza audio call (silenzio/parziale).

### Gravità
**Bloccante funzionale** lato “registrazione automatica utile”.

---

## C6 — **Path runtime e output non allineati al deployment service**
**File:** `capture_linux.go`, `manager.go`, `handler.go`  
**Evidenza:**
- capture scrive in `recordings` relativo,
- whisper path hardcoded relativi (`whisper/whisper.cpp/main`, model path),
- `handler` restituisce `RecordingPath` placeholder `/tmp/...`.

### Impatto
- Fallimenti dipendenti dalla working directory del daemon (systemd user service).
- Diagnostica fuorviante sul path reale file.

### Gravità
**Alta**.

---

## C7 — **Incoerenza architetturale tra documentazione e implementazione testuale**

La documentazione descrive robustezza Teams v2 e criteri più restrittivi; parte del codice usa parametri o comportamenti meno conservativi (soglie, visibilità, path placeholder).

### Impatto
- Aspettativa utente/QA diversa dal comportamento reale.

### Gravità
**Media**.

---

## 5) Errori architetturali / anti-pattern rilevati

1. **UI-driven detection senza contratto stabilizzato di fallback runtime**: troppo affidamento su pattern DOM volatili.
2. **Separation of concerns incompleta**: detector valuta stato, ma non produce evento diagnostico strutturato standard con causale machine-readable condivisa.
3. **Linux audio capture coupling debole al contesto utente**: manca discovery/selection esplicita source monitor.
4. **Path relativi in componenti daemonici**: anti-pattern classico per servizi lanciati da init manager.
5. **Placeholder in response path (`/tmp/...`)**: rompe fiducia osservability.

---

## 6) Bug tecnici specifici (con evidenza)

1. **Selector invalido**: `[class*="call-"]-container` (invalido CSS) in `callContainer`.
2. **Branch di fase potenzialmente contro-intuitivo** in `evaluatePhase` (precedenze START/PREJOIN/ENDING non ottimali).
3. **Heuristic visibility**: ritorno `true` su `data-tid` senza verifiche complete.
4. **`queryAll` aggrega match multipli senza dedup** → conteggi media possibili duplicati.
5. **Messaggio errore fuorviante** in capture Linux (`failed to start pw-record` ma usa ffmpeg).
6. **Output path placeholder** in `handler.go` per START_RECORDING response.

---

## 7) Failure modes osservabili lato utente

1. **Nessun start registrazione** nonostante call attiva (falso negativo detector o bridge failure).
2. **Start/stop multipli o prematuri** durante transizioni UI Teams.
3. **Registrazione creata ma audio vuoto/errato** per source PulseAudio/PipeWire non appropriata.
4. **Trascrizione non prodotta** per path whisper/model non risolti in servizio.
5. **Log Chrome “apparentemente ok” ma file mancanti/incoerenti** per mismatch path runtime.

---

## 8) Compliance SOTA 2026 (gap rispetto a baseline attesa)

Per uno stack locale browser-extension + native capture “production-grade 2026”, ci si aspetta:

- detector resilienti con confidence governance conservativa,
- telemetria strutturata cross-layer con correlation-id,
- source audio selection deterministica su Linux,
- path assoluti/config-driven per tutti gli artefatti,
- test E2E browser reali (non solo jsdom).

Lo stato attuale è **parzialmente conforme** ma non ancora “fully reliable”.

---

## 9) Matrice severità / priorità

| ID | Area | Severità | Priorità intervento |
|---|---|---|---|
| C5 | Linux audio capture source | Bloccante | P0 |
| C1 | Detector soglie/fasi | Alta | P0 |
| C2 | isElementVisible | Alta | P0 |
| C6 | Path runtime/placeholder | Alta | P0 |
| C4 | Observability end-to-end | Alta operativa | P1 |
| C3 | Selector quality | Media-Alta | P1 |
| C7 | Allineamento doc/impl | Media | P2 |

---

## 10) Conclusione

La pipeline è presente ma non ancora robusta per affidabilità “always-on” su Teams Linux.  
Il problema non è singolo: è un insieme di fragilità nel detector Teams + assunzioni Linux audio capture + gap di osservabilità e deployment runtime.

Per rendere il sistema **completamente funzionante sul tuo PC Linux** è necessario un remediation plan multi-fase con priorità P0 sui punti C5/C1/C2/C6, seguito da hardening P1/P2.

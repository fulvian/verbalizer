# Verbalizer — Teams Web Support (Chrome) 
## Analisi profonda codebase, gap di implementazione, criticità e piano tecnico 2026-compliant

**Data:** 2026-03-25  
**Stato:** Draft per approvazione formale (nessuna implementazione in questa fase)

---

## 1) Executive summary

Lo stato attuale mostra che il supporto Microsoft Teams **esiste a livello strutturale** (routing URL, detector dedicato, tipi, manifest, pipeline start/stop), ma l’implementazione del detector Teams è oggi **fragile e in parte logicamente errata** per uso reale su Teams Web moderno.

Conclusione operativa:
- Il sistema è pronto per un hardening Teams senza refactor architetturale radicale.
- È necessario introdurre un detector Teams v2 basato su **segnali multipli + state machine + stabilizzazione temporale**, evitando dipendenze da un singolo selettore DOM.
- La validazione attuale è quasi tutta unit-test jsdom; manca copertura real-browser su Teams reale.

---

## 2) Perimetro analizzato (codebase iniziale)

### Extension (MV3)
- `extension/manifest.json`
- `extension/src/content/index.ts`
- `extension/src/content/observer.ts`
- `extension/src/content/detectors/teams.ts`
- `extension/src/content/detectors/meet.ts`
- test correlati `*.test.ts`

### Bridge e backend
- `extension/src/background/index.ts`
- `extension/src/background/native-bridge.ts`
- `native-host/cmd/main.go`
- `native-host/internal/messaging/protocol.go`
- `daemon/cmd/verbalizerd/handler.go`
- `daemon/pkg/api/types.go`
- `daemon/internal/session/manager.go`

### Documentazione e baseline progetto
- `docs/ARCHITECTURE.md`
- `docs/INSTALLATION.md`

---

## 3) Implementazione effettiva Teams: cosa c’è davvero oggi

## 3.1 Routing e abilitazione piattaforma

1. **Manifest host permissions** include:
   - `https://teams.microsoft.com/*`
   - `https://teams.live.com/*`
2. **Content script** è iniettato anche su Teams (document_idle).
3. `detectPlatform()` in `content/index.ts` mappa correttamente i due domini a `ms-teams`.

Stato: ✅ presente.

## 3.2 Detector Teams corrente

File: `extension/src/content/detectors/teams.ts`

Approccio attuale:
- polling ogni 1s + `MutationObserver` su `document.body`.
- euristiche basate su selettori `data-tid`/classi.
- trigger verso observer: `notifyCallDetected`, `notifyCallStarted`, `notifyCallEnded`.

Stato: ⚠️ presente ma fragile.

## 3.3 Pipeline end-to-end dopo detection

1. `CALL_STARTED` inviato al background.
2. Background chiama native host con `START_RECORDING`.
3. Native host inoltra al daemon.
4. Daemon registra, poi su `CALL_ENDED` stoppa e trascrive.

Stato: ✅ pipeline uniforme Meet/Teams già pronta.

## 3.4 Test esistenti

- `teams.test.ts` presente e passing.
- Tutta la suite extension è verde (84 test passati).

Stato: ✅ buona base unit, ❌ mancano test browser reali su Teams live.

---

## 4) Gap analysis dettagliata Teams

## 4.1 Gap bloccanti / alta severità

### G1 — Logica “call ended” invertita sul bottone hangup
**Evidenza codice:** `isMSTeamsActive()` ritorna `false` se trova `hangup-button`.  
**Problema:** in UI meeting reali il pulsante hangup è tipicamente presente durante call attiva.  
**Effetto:** falsi negativi, recording che può non partire o interrompersi erroneamente.

### G2 — Dipendenza eccessiva da selettori DOM potenzialmente instabili
**Evidenza:** uso hardcoded di `data-tid` e classi legacy (`.ts-*`) senza fallback robusti.  
**Effetto:** regressioni immediate a ogni update Teams frontend.

### G3 — Nessun controllo visibilità/stato reale degli elementi
**Evidenza:** check su `querySelector` (esistenza nodo) senza verificare visibilità/aria state/attributi semantici.  
**Effetto:** falsi stati quando elementi persistono nascosti nel DOM.

## 4.2 Gap medi

### G4 — Nessuna stabilizzazione anti-flap
Polling + mutation possono oscillare su transizioni (joining/leaving/reconnect). Manca hysteresis o finestra minima di conferma.

### G5 — `notifyCallDetected()` e `notifyCallStarted()` rigenerano callId
`observer.ts` genera due callId distinti nelle due notify. Non blocca la pipeline, ma riduce tracciabilità e correlazione eventi.

### G6 — Segnali definiti ma non usati
`activeCall` e `callControls` sono definiti nei selectors Teams ma non partecipano alla decisione.

## 4.3 Gap di quality engineering

### G7 — Test Teams non realistici
I test validano il comportamento attuale, ma su DOM sintetico jsdom; non proteggono da cambi UI reali.

### G8 — Mancanza observability specifica detector
Mancano metriche/telemetria locali su: motivo transizione stato, selector hit-rate, false-start sospetti.

---

## 5) Criticità operative e rischi

1. **Rischio regressione silenziosa alto**: Teams cambia frontend frequentemente.
2. **Rischio mancata registrazione**: call reale non rilevata o end prematuro.
3. **Rischio manutenzione costosa**: senza catalogo selector versionato, hotfix reattivi continui.
4. **Rischio compliance funzionale**: comportamento percepito non deterministico lato utente.

---

## 6) Pattern e best practice (2026) da fonti tecniche

## 6.1 Chrome Extensions MV3
- Content scripts su host espliciti, ciclo vita coerente con `document_idle`.
- Native messaging solo in extension contexts (non direttamente da content script).
- Con `sendNativeMessage`, Chrome avvia un processo host per richiesta: preferire pochi eventi di dominio “stabili” e non stream rumoroso.

## 6.2 DOM observation
- `MutationObserver` va usato con callback leggera, filtrata e con debounce/batching.
- Decisione di stato non basata su singolo nodo: usare segnali multipli e punteggio/confidenza.

## 6.3 Selettori DOM
- `querySelector` è fragile con selettori invalidi/dinamici; introdurre wrapper safe e fallback.
- Evitare coupling su classi “styling-only”; privilegiare attributi semantici quando disponibili.

Nota: non esiste un contratto pubblico stabile Microsoft per i selettori DOM interni di Teams Web; la strategia corretta è progettare detector resiliente ai cambi UI, non selector statico singolo.

---

## 7) Obiettivo tecnico target (Teams v2)

Realizzare un detector Teams con:
1. **State machine esplicita**: `idle -> prejoin -> in_call -> ending -> idle`
2. **Multi-signal scoring** (non regola binaria singola)
3. **Debounce/hysteresis** (es. 1500–2500 ms di conferma)
4. **Determinismo eventi**: singolo `CALL_STARTED` per sessione, singolo `CALL_ENDED`.
5. **Observer cleanup robusto** e uso CPU controllato.

---

## 8) Design tecnico proposto

## 8.1 Nuovo contratto detector

Creare un evaluator puro:

```ts
type CallPhase = 'idle' | 'prejoin' | 'in_call' | 'ending';

interface SignalSample {
  hasCallContainer: boolean;
  hasPrejoin: boolean;
  hasActiveMedia: boolean;
  hasCallControls: boolean;
  hasHangupControlVisible: boolean;
  meetingTitle?: string;
}

interface EvaluationResult {
  phase: CallPhase;
  confidence: number; // 0..1
  reasons: string[];
}
```

Il detector usa `EvaluationResult` + finestra di stabilità prima di notificare start/end.

## 8.2 State machine con stabilizzazione

```ts
const START_THRESHOLD = 0.7;
const END_THRESHOLD = 0.3;
const STABLE_MS = 2000;

// Se confidence > START_THRESHOLD per STABLE_MS => CALL_STARTED
// Se confidence < END_THRESHOLD per STABLE_MS => CALL_ENDED
```

Questo elimina i “flap” durante join/reconnect/mini refresh SPA.

## 8.3 Strategia segnali Teams

Ordine consigliato (priorità alta -> bassa):
1. indicatori container call attivo
2. controlli call visibili + azione hangup disponibile
3. media elements attivi (`video`/`audio` con `readyState` utile)
4. URL pattern meeting-specific (solo come supporto)

Regola chiave: **hangup visibile è segnale positivo di call attiva**, non di call terminata.

## 8.4 Selector registry versionato

Introdurre struttura:

```ts
interface SelectorSet {
  id: string;            // es. teams-web-v2-2026q1
  selectors: {
    callContainer: string[];
    prejoin: string[];
    callControls: string[];
    hangup: string[];
    participants: string[];
    title: string[];
  };
}
```

Permette rollout/rollback rapido senza riscrivere logica core.

## 8.5 Event correlation fix (observer)

Stabilizzare callId lifecycle:
- `CALL_DETECTED` può creare callId una sola volta.
- `CALL_STARTED` riusa stesso callId.
- `CALL_ENDED` chiude stesso callId.

---

## 9) Piano di implementazione dettagliato (senza coding ora)

## Fase 0 — Baseline e safety net
**Output:** baseline tecnica prima refactor.

Task:
1. Congelare comportamento attuale con test snapshot dei detector.
2. Aggiungere casi di test che riproducono bug hangup-as-ended.
3. Documentare matrice stati attesi (idle/prejoin/in_call/ended).

Acceptance:
- test baseline presenti e ripetibili.

---

## Fase 1 — Refactor detector in evaluator + state machine
**File target (previsti):**
- `extension/src/content/detectors/teams.ts`
- nuovo `extension/src/content/detectors/teams-evaluator.ts`
- nuovo `extension/src/content/detectors/teams-selectors.ts`

Task:
1. Estrarre raccolta segnali in funzione pura testabile.
2. Implementare state machine con soglie/stabilizzazione.
3. Correggere semantica hangup.
4. Aggiungere guardie per elementi nascosti/non attivi.

Acceptance:
- nessun doppio START/END per singola sessione simulata.
- bug hangup risolto da test specifico.

---

## Fase 2 — Hardening observer e performance
**Task:**
1. Debounce callback MutationObserver.
2. Limitare `attributeFilter` e scope dei trigger.
3. Introdurre tick polling più intelligente (es. 1s idle, 500ms durante transizioni).

Acceptance:
- riduzione eventi superflui.
- nessun leak su cleanup.

---

## Fase 3 — Affidabilità messaging
**Task:**
1. Correggere lifecycle callId in `observer.ts`.
2. Aggiungere idempotenza lato background (ignora start/stop duplicati ravvicinati con stesso callId).

Acceptance:
- coerenza callId end-to-end nei log.

---

## Fase 4 — Test strategy completa
**Task:**
1. Unit test evaluator (matrice segnali -> stato/confidenza).
2. Integration test detector con sequenze temporali (join, reconnect, leave).
3. Smoke test manuale su Teams Web reale (account work/school e personal).

Acceptance:
- checklist QA firmata con evidenze real-browser.

---

## Fase 5 — Observability e runbook
**Task:**
1. Log strutturati detector (`phase`, `confidence`, `reasons`).
2. Documento runbook “selector rotto: come fare hotfix”.
3. Aggiornamento docs installazione/verifica per Teams.

Acceptance:
- troubleshooting rapido in produzione locale.

---

## 10) Snippet tecnici di riferimento (base implementazione)

## 10.1 Debounce mutation handler

```ts
function debounce<T extends (...args: any[]) => void>(fn: T, ms: number): T {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return ((...args: any[]) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => fn(...args), ms);
  }) as T;
}
```

## 10.2 Safe selector query helper

```ts
function queryAny(selectors: string[]): Element | null {
  for (const s of selectors) {
    try {
      const el = document.querySelector(s);
      if (el) return el;
    } catch {
      // selector invalido: ignora e continua fallback
    }
  }
  return null;
}
```

## 10.3 Stabilizzazione transizione di stato

```ts
interface StableState {
  candidate: CallPhase;
  since: number;
}

function isStable(state: StableState, now: number, stableMs: number): boolean {
  return now - state.since >= stableMs;
}
```

## 10.4 Correlazione callId unica

```ts
// pseudo: creare solo se null
if (!this.currentCallId) this.currentCallId = this.generateCallId(this._platform);
// riuso in CALL_STARTED e CALL_ENDED
```

---

## 11) Piano QA e criteri di accettazione finali

Il supporto Teams si considera “ready” solo se:
1. start/stop recording affidabili in 3 scenari: join diretto, join da prejoin, reconnect.
2. nessun doppio start/stop in sessioni > 30 min.
3. nessun falso end alla sola presenza pulsante hangup.
4. callId coerente lungo tutta la sessione.
5. test automatici aggiornati + smoke test manuale documentato.

---

## 12) Rischi residui e mitigazioni

| Rischio | Mitigazione |
|---|---|
| UI Teams cambia senza preavviso | selector registry versionato + evaluator multi-signal |
| falsi positivi su media elements | richiedere almeno 2 segnali concordi + stabilizzazione |
| regressione Meet durante refactor condiviso | isolamento codice Teams in moduli dedicati + test Meet invarianti |
| manutenzione eccessiva | runbook e test fixture aggiornabili rapidamente |

---

## 13) Scope esplicitamente escluso da questa fase

- Nessuna modifica codice runtime.
- Nessun refactor del daemon audio/transcription.
- Nessuna introduzione di nuove piattaforme oltre Teams Web su Chrome.

---

## 14) Deliverable di questa attività

1. Analisi profonda implementazione corrente Teams (questo documento).
2. Gap analysis e criticità prioritarie.
3. Piano tecnico completo, con fasi, acceptance criteria e snippet base.

Documento pronto per approvazione formale prima della fase implementativa.

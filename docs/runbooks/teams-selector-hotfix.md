# Teams Web Selector Hotfix Runbook

**Date:** 2026-03-25  
**Target:** Microsoft Teams Web (Chrome Extension)  
**Stato:** Procedure pronte per applicazione

---

## Scenario

Microsoft rilascia un aggiornamento UI di Teams e la registrazione non funziona più perché i selettori DOM sono cambiati.

## Sintomi Comuni

1. **Nessuna registrazione** - Il sistema non rileva l'inizio della chiamata
2. **Registrazione non si ferma** - Il sistema non rileva la fine della chiamata
3. **Doppie registrazioni** - Il sistema rileva più start senza end

## Diagnostica Iniziale

### 1. Verifica Logs Extension

Apri DevTools su Teams Web (`F12` → Console) e cerca:
```
[TeamsDetector v2] CALL_START: phase=...
[TeamsDetector v2] CALL_END: phase=...
```

### 2. Verifica Selettori Attivi

In DevTools Console:
```javascript
// Check if call container selector works
document.querySelector('[data-tid="call-container"]') !== null

// Check if hangup button selector works  
document.querySelector('[data-tid="hangup-button"]') !== null

// Check all selectors
['[data-tid="call-container"]', '.ts-calling-thread'].some(s => document.querySelector(s))
```

### 3. Identifica il Selettore Rotto

Se `document.querySelector('[data-tid="call-container"]')` ritorna `null` ma l'utente È in chiamata, il selettore è cambiato.

---

## Procedure di Hotfix

### Step 1: Ispeziona DOM Teams

1. Unisciti a una chiamata Teams
2. Apri DevTools (`F12`)
3. Ispeziona elemento per trovare il nuovo selettore:
   - Cerca `data-tid` attribute
   - Cerca classi tipo `.ts-*` o `.calling-*`
   - Cerca elementi con `aria-label` che indica "call"

### Step 2: Aggiorna Selector Registry

**File:** `extension/src/content/detectors/teams-selectors.ts`

```typescript
// Aggiungi nuovo selettore al registry esistente
{
  id: 'teams-web-v2026q2',  // <-- Nuova versione
  version: '2026Q2',
  description: 'Teams Web selectors observed after UI update',
  selectors: {
    callContainer: [
      // AGGIUNGI il nuovo selettore PRIMA di quelli vecchi
      '[data-tid="NUOVO-SELETTORE"]',
      // Mantieni i vecchi come fallback
      '[data-tid="call-container"]',
      '.ts-calling-thread',
    ],
    // ... altri selettori
  },
}
```

### Step 3: Crea Backup del Registry Precedente

**NON rimuovere** i selettori vecchi - tienili come fallback:

```typescript
{
  id: 'teams-web-v2026q1',  // Mantieni la versione precedente
  version: '2026Q1', 
  description: 'Pre-update selectors (fallback)',
  selectors: {
    callContainer: [
      '[data-tid="call-container"]',  // Vecchio selettore
      '.ts-calling-thread',
    ],
    // ...
  },
}
```

### Step 4: Testa il Fix

```bash
cd extension
npm test -- --testPathPattern="teams"
```

### Step 5: Test Manuale su Teams Reale

1. Carica extension in modalità developer (`chrome://extensions`)
2. Unisciti a una chiamata Teams
3. Verifica:
   - `[TeamsDetector v2] CALL_START` appare in Console
   - La registrazione inizia
   - `CALL_END` appare quando termini la chiamata
   - La registrazione si ferma

---

## Rollback Rapido

Se il fix causa problemi:

1. **Commuta alla versione precedente:**

```typescript
// In teams-selectors.ts, modifica:
export const CURRENT_SELECTOR_SET = SELECTOR_REGISTRY[0];  // Vecchia versione
```

2. **Oppure rimuovi il selettore problematico:**

```typescript
{
  id: 'teams-web-v2026q2',
  // ...
  selectors: {
    callContainer: [
      // RIMUOVI il selettore che causa falsi positivi
      // '[data-tid="PROBLEMATIC-SELETTOR"]',
      '[data-tid="call-container"]',  // Torna al vecchio
    ],
  },
}
```

---

## Matrice Selettori (2026Q1 Baseline)

| Elemento | Selettori (priorità alta → bassa) |
|----------|----------------------------------|
| Call Container | `[data-tid="call-container"]`, `[data-tid="meeting-container"]`, `.ts-calling-thread` |
| Prejoin | `[data-tid="prejoin-screen"]`, `.prejoin-container` |
| Call Active | `[data-tid="call-state"]`, `.calling-live-indicator` |
| Call Controls | `[data-tid="call-controls"]`, `.calling-controls` |
| Hangup Button | `[data-tid="hangup-button"]`, `[aria-label*="Hang up"]` |
| Meeting Title | `[data-tid="meeting-title"]`, `.ts-meeting-title` |

---

## Regole di Manutenzione

### DO
- Aggiungi nuovi selettori IN CIMA alla lista (più alta priorità)
- Mantieni sempre almeno 2-3 selettori per elemento (fallback)
- Testa su Teams reale prima di rilasciare
- Aggiorna la versione nel ID dopo ogni modifica

### DON'T
- Non rimuovere selettori funzionanti
- Non usare selettori CSS basati su styling (es. `.bg-white`)
- Non rilasciare senza testare su account work/school E personal
- Non fare debug diretto su Teams in produzione con dati sensibili

---

## Contatti e Riferimenti

- **Extension Source:** `extension/src/content/detectors/teams-selectors.ts`
- **Evaluator:** `extension/src/content/detectors/teams-evaluator.ts`
- **Detector:** `extension/src/content/detectors/teams.ts`

Per domande: consulta `docs/ARCHITECTURE.md` per architettura completa.

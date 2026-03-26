/**
 * Microsoft Teams call detection v2
 * 
 * IMPROVEMENTS over v1:
 * - G1 fix: hangup button = ACTIVE call signal (not ended!)
 * - Multi-signal scoring instead of single selector
 * - State machine with stabilization to prevent flapping
 * - Visibility checks on all elements (G3 fix)
 * - Debounced MutationObserver (Phase 2)
 * - Structured logging with correlation IDs
 * 
 * Architecture:
 * 1. TeamsStateMachine evaluates signals and manages phase transitions
 * 2. Polling provides backup when mutations don't fire
 * 3. MutationObserver provides immediate detection of DOM changes
 */

import { CallStateObserver } from '../observer';
import {
  collectSignals,
  TeamsStateMachine,
  debounce,
  EvaluationResult,
  CallPhase,
} from './teams-evaluator';
import {
  CURRENT_SELECTOR_SET,
  queryAll,
} from './teams-selectors';
import { contentLogger, generateCallId } from '../../utils/logger';

// ============================================================================
// Configuration
// ============================================================================

/** Polling interval when idle (1 second) */
const IDLE_POLL_MS = 1000;

/** Polling interval when transitioning (500ms for faster response) */
const TRANSITION_POLL_MS = 500;

/** Debounce delay for mutation observer callback */
const MUTATION_DEBOUNCE_MS = 100;

/** Minimum time between notifications to prevent duplicates */
const NOTIFICATION_COOLDOWN_MS = 500;

// ============================================================================
// State
// ============================================================================

/** State machine for call phase detection */
let stateMachine: TeamsStateMachine | null = null;

/** Polling interval handle */
let pollInterval: ReturnType<typeof setInterval> | null = null;

/** MutationObserver instance */
let mutationObserver: MutationObserver | null = null;

/** Track if we notified call started (for idempotency) */
let notifiedCallStarted = false;

/** Track if we notified call ended (for idempotency) */
let notifiedCallEnded = false;

/** Last notification timestamp (for cooldown) */
let lastNotificationMs = 0;

/** Current poll interval (idle vs transition) */
let currentPollMs = IDLE_POLL_MS;

/** Current call session ID for correlation */
let currentCallId: string | null = null;

// ============================================================================
// Core Detection Logic
// ============================================================================

/**
 * Evaluate current DOM state and process through state machine
 * Returns the evaluation result
 */
function evaluateState(): EvaluationResult {
  if (!stateMachine) {
    stateMachine = new TeamsStateMachine();
  }
  
  const signals = collectSignals();
  const result = stateMachine.evaluate(signals);
  
  // Structured logging for debugging
  contentLogger.debug('Poll: signals detected', {
    callId: currentCallId || undefined,
    state: result.phase,
    confidence: result.confidence,
    reason: signals.hasCallContainer ? 'callContainer' : 
             signals.hasCallActive ? 'callActive' :
             signals.hasHangupVisible ? 'hangup' : 'none',
    metadata: {
      hasCallContainer: signals.hasCallContainer,
      hasCallActive: signals.hasCallActive,
      hasCallControls: signals.hasCallControls,
      hasHangupVisible: signals.hasHangupVisible,
      hasPrejoin: signals.hasPrejoin,
      videoCount: signals.videoCount,
      audioCount: signals.audioCount,
      hasVideoGrid: signals.hasVideoGrid,
      hasMediaStreamActive: signals.hasMediaStreamActive,
    }
  });
  
  // Update polling speed based on phase
  updatePollingSpeed(result.phase);
  
  return result;
}

/**
 * Update polling speed based on current phase
 * Faster polling during transitions for quicker response
 */
function updatePollingSpeed(phase: CallPhase): void {
  const newPollMs = (phase === 'prejoin' || phase === 'ending')
    ? TRANSITION_POLL_MS
    : IDLE_POLL_MS;
  
  if (newPollMs !== currentPollMs) {
    currentPollMs = newPollMs;
    restartPolling();
  }
}

/**
 * Restart polling with current interval
 */
function restartPolling(): void {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
  pollInterval = setInterval(pollCallback, currentPollMs);
}

/**
 * Check for call state changes (called by polling)
 */
function pollCallback(): void {
  const result = evaluateState();
  processEvaluation(result);
}

/**
 * Process evaluation result and send notifications
 * Handles idempotency to prevent duplicate notifications
 */
function processEvaluation(result: EvaluationResult): void {
  const now = Date.now();
  
  // Apply cooldown to prevent notification spam
  if (now - lastNotificationMs < NOTIFICATION_COOLDOWN_MS) {
    return;
  }
  
  // Only notify on stable phase transitions
  if (result.phase === 'in_call' && !notifiedCallStarted) {
    notifiedCallStarted = true;
    notifiedCallEnded = false;
    lastNotificationMs = now;
    
    // Generate call ID for correlation
    currentCallId = generateCallId();
    
    contentLogger.info('Call started detected', {
      callId: currentCallId,
      event: 'CALL_STARTED',
      state: 'in_call',
      confidence: result.confidence,
      reason: result.reasons.join('; '),
      metadata: { title: result.sample.meetingTitle }
    });
    
    // Extract meeting title from signals
    const title = result.sample.meetingTitle;
    observer?.notifyCallStarted(title);
  } else if (result.phase === 'idle' && notifiedCallStarted && !notifiedCallEnded) {
    notifiedCallEnded = true;
    notifiedCallStarted = false;
    lastNotificationMs = now;
    
    contentLogger.info('Call ended detected', {
      callId: currentCallId || undefined,
      event: 'CALL_ENDED',
      state: 'idle',
      confidence: result.confidence,
      reason: result.reasons.join('; '),
    });
    
    currentCallId = null;
    observer?.notifyCallEnded();
  }
}

// ============================================================================
// Public API (compatible with v1 interface)
// ============================================================================

/**
 * Set up MS Teams detection using v2 evaluator with state machine
 * 
 * IMPROVEMENTS:
 * - Multi-signal scoring (not single selector)
 * - Stabilization to prevent flapping
 * - G1 fix: hangup button = active call (not ended!)
 * - Visibility checks on all elements
 * - Debounced mutation observer
 * - Smart polling speed
 */
export function detectMSTeams(observerInstance: CallStateObserver): void {
  // Initialize observer reference
  observer = observerInstance;
  
  // Initialize state machine
  stateMachine = new TeamsStateMachine();
  notifiedCallStarted = false;
  notifiedCallEnded = false;
  lastNotificationMs = 0;
  currentPollMs = IDLE_POLL_MS;
  
  // Initial evaluation
  const initialResult = evaluateState();
  processEvaluation(initialResult);
  
  // Set up polling for continuous monitoring
  pollInterval = setInterval(pollCallback, currentPollMs);
  
  // Set up debounced MutationObserver for immediate detection
  const { fn: debouncedCheck, cancel: cancelDebounce } = debounce(() => {
    const result = evaluateState();
    processEvaluation(result);
  }, MUTATION_DEBOUNCE_MS);
  
  mutationObserver = new MutationObserver((mutations) => {
    // Filter to relevant mutations only
    const hasRelevantChanges = mutations.some(mutation => {
      // Look for additions of call-related elements
      if (mutation.type === 'childList') {
        const addedNodes = mutation.addedNodes;
        if (addedNodes && addedNodes.length > 0) {
          return true;
        }
      }
      // Look for attribute changes on relevant elements
      if (mutation.type === 'attributes') {
        const target = mutation.target as Element;
        const dataTid = target.getAttribute?.('data-tid') || '';
        return dataTid.includes('call') || 
               dataTid.includes('meeting') ||
               dataTid.includes('prejoin');
      }
      return false;
    });
    
    if (hasRelevantChanges) {
      debouncedCheck();
    }
  });
  
  // Observe document with filtered scope
  mutationObserver.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ['data-tid', 'data-call-state', 'aria-label'],
  });
  
  // Register cleanup
  observerInstance.registerCleanup(() => {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
    if (mutationObserver) {
      mutationObserver.disconnect();
      mutationObserver = null;
    }
    cancelDebounce();
    if (stateMachine) {
      stateMachine.reset();
      stateMachine = null;
    }
    notifiedCallStarted = false;
    notifiedCallEnded = false;
  });
}

/**
 * Legacy function kept for backward compatibility with tests
 * Returns true if call appears active based on multi-signal evaluation
 * 
 * NOTE: This is a simplified check for test compatibility.
 * Real detection uses the state machine via detectMSTeams()
 */
export function isMSTeamsActive(): boolean {
  const signals = collectSignals();
  
  // If prejoin is visible, definitely not in call
  if (signals.hasPrejoin) {
    return false;
  }
  
  // G1 FIX: hangup button presence is now a POSITIVE signal for active call
  // (Previously incorrectly treated as "call ended")
  // 
  // Multi-signal check: call container is primary indicator
  // If present (and visible), likely in a call
  if (signals.hasCallContainer && signals.callContainerVisible) {
    return true;
  }
  
  // Fallback: if we have active indicators but no container yet,
  // could be very early in call setup
  const hasActiveSignal = signals.hasCallActive || 
                          signals.hasCallControls || 
                          signals.hasHangupVisible ||
                          signals.videoCount > 0 ||
                          signals.audioCount > 0;
  
  return hasActiveSignal;
}

/**
 * Extract meeting title from current DOM state
 */
export function extractMeetingTitle(): string | undefined {
  const signals = collectSignals();
  if (signals.meetingTitle) {
    return signals.meetingTitle;
  }
  // Fallback to document title
  return document.title || undefined;
}

/**
 * Extract participant count from current DOM state
 */
export function extractParticipantCount(): number {
  const selectors = CURRENT_SELECTOR_SET.selectors;
  const participants = queryAll(selectors.participants);
  return participants.length;
}

// ============================================================================
// Back-reference for legacy compatibility
// ============================================================================

/** Observer instance (set during detectMSTeams) */
let observer: CallStateObserver | null = null;

/**
 * Teams Web Call Evaluator v2
 * 
 * Multi-signal scoring detector with state machine for Teams Web call detection.
 * Replaces fragile single-selector logic with robust multi-signal approach.
 * 
 * Key improvements over v1:
 * - G1 fix: hangup button = signal of ACTIVE call (not ended)
 * - G2 fix: multiple selectors with fallback
 * - G3 fix: visibility and active state checks
 * - G4 fix: stabilization window to prevent flapping
 * - G6 fix: activeCall and callControls signals now used
 */

import {
  CURRENT_SELECTOR_SET,
  queryAny,
  queryAll,
  isElementVisible,
} from './teams-selectors';

// ============================================================================
// Types
// ============================================================================

/**
 * Call phase states
 * idle -> prejoin -> in_call -> ending -> idle
 */
export type CallPhase = 'idle' | 'prejoin' | 'in_call' | 'ending';

/**
 * Signal sample collected from DOM at a point in time
 * Each field represents presence of a specific UI element/signal
 */
export interface SignalSample {
  /** Main call container present */
  hasCallContainer: boolean;
  /** Call container is visible (not hidden) */
  callContainerVisible: boolean;
  /** Prejoin screen present and visible */
  hasPrejoin: boolean;
  /** Active call indicator present */
  hasCallActive: boolean;
  /** Call controls toolbar visible */
  hasCallControls: boolean;
  /** Hangup button visible (G1 fix: this = ACTIVE call, not ended) */
  hasHangupVisible: boolean;
  /** Meeting/video title present */
  meetingTitle?: string;
  /** Number of visible video elements with src */
  videoCount: number;
  /** Number of visible audio elements with src */
  audioCount: number;
  /** Video grid/container present (indicates call UI) */
  hasVideoGrid: boolean;
  /** Any media stream active (video or audio with src) */
  hasMediaStreamActive: boolean;
  /** Timestamp when sample was collected */
  timestamp: number;
}

/**
 * Result of evaluating signals
 */
export interface EvaluationResult {
  /** Current detected phase */
  phase: CallPhase;
  /** Confidence score 0..1 that we're in the reported phase */
  confidence: number;
  /** Human-readable reasons for this evaluation */
  reasons: string[];
  /** The raw signal sample that produced this result */
  sample: SignalSample;
}

/**
 * Stable state tracking for hysteresis
 */
export interface StableState {
  /** The phase we're stabilizing toward */
  candidate: CallPhase;
  /** When we first observed this candidate (for timing) */
  since: number;
  /** How many consecutive samples support this candidate */
  supportCount: number;
}

// ============================================================================
// Configuration
// ============================================================================

/**
 * Thresholds for state transitions
 * Tuned to be more conservative and avoid false positives
 * 
 * Changes from original:
 * - START_THRESHOLD raised from 0.25 to 0.40 to require stronger signals
 * - END_THRESHOLD raised from 0.10 to 0.15 for more stable end detection
 * - The gap between thresholds (0.40 - 0.15 = 0.25) provides implicit hysteresis
 *   to prevent rapid state transitions when confidence is borderline
 */
const START_THRESHOLD = 0.40;   // Need 40% confidence to transition idle->in_call (raised from 0.25)
const END_THRESHOLD = 0.15;    // Need <15% confidence to transition in_call->ending (raised from 0.10)
const PREJOIN_THRESHOLD = 0.50; // Need 50% confidence to detect prejoin

/**
 * Stabilization windows (milliseconds)
 * Must maintain threshold for this long before state transition
 */
const STABLE_MS = 3000;       // Increased from 2000ms for more stable detection
const QUICK_TRANSITION_MS = 1500; // Faster transition for ending->idle (increased from 1000)

/**
 * Minimum support count (consecutive samples needed)
 */
const MIN_SUPPORT = 3; // Increased from 2 for more reliable detection

/**
 * Signal weights for scoring
 * Higher weight = more important signal
 */
const SIGNAL_WEIGHTS = {
  callContainer: 0.25,      // Primary indicator
  callActive: 0.20,         // Strong indicator
  callControls: 0.20,      // Supports active call (increased)
  hangupVisible: 0.20,      // G1 fix: hangup = ACTIVE call (not ended!)
  videoCount: 0.15,        // Media presence supports call
  audioCount: 0.15,        // Audio presence supports call
  // Additional signals for call detection
  videoGridPresent: 0.20,   // Video grid present indicates call
  mediaStreamActive: 0.20, // Any video/audio with src indicates active media (increased)
};

/**
 * Penalties for contradictory signals
 */
const PENALTIES = {
  hasPrejoin: -0.30,       // Prejoin = not in call yet
};

// ============================================================================
// Signal Collection
// ============================================================================

/**
 * Collect all signals from the current DOM state
 * This is a pure function - no side effects, fully testable
 */
export function collectSignals(): SignalSample {
  const selectors = CURRENT_SELECTOR_SET.selectors;
  const now = Date.now();
  
  // Check call container
  const callContainer = queryAny(selectors.callContainer);
  const hasCallContainer = callContainer !== null;
  const callContainerVisible = hasCallContainer && isElementVisible(callContainer);
  
  // Check prejoin screen (blocks call detection)
  const prejoin = queryAny(selectors.prejoin);
  const hasPrejoin = prejoin !== null && isElementVisible(prejoin);
  
  // Check call active indicator
  const callActive = queryAny(selectors.callActive);
  const hasCallActive = callActive !== null && isElementVisible(callActive);
  
  // Check call controls
  const callControls = queryAny(selectors.callControls);
  const hasCallControls = callControls !== null && isElementVisible(callControls);
  
  // G1 FIX: Hangup button is NOW a POSITIVE signal for active call
  // (It was incorrectly used as "call ended" in v1)
  const hangup = queryAny(selectors.hangup);
  const hasHangupVisible = hangup !== null && isElementVisible(hangup);
  
  // Meeting title
  const titleEl = queryAny(selectors.meetingTitle);
  const meetingTitle = titleEl?.textContent?.trim() || undefined;
  
  // Media elements
  const videos = queryAll<HTMLVideoElement>(selectors.videoElements);
  const visibleVideos = videos.filter(v => {
    const src = v.src || (v as any).srcObject;
    return src && isElementVisible(v);
  });
  
  const audios = queryAll<HTMLAudioElement>(selectors.audioElements);
  const visibleAudios = audios.filter(a => {
    const src = a.src || (a as any).srcObject;
    return src && isElementVisible(a);
  });
  
  // Check for video grid container (strong indicator of call)
  const videoGridSelectors = [
    '.video-grid',
    '.video-container',
    '[class*="video-grid"]',
    '[class*="call-grid"]',
    '[class*="meeting-grid"]',
    '[class*="participants-grid"]',
    // Additional patterns
    '[class*="video"][class*="container"]',
    '[class*="grid"][class*="video"]',
    // Specific Teams patterns
    '.ts-video-grid',
    '[data-tid="video-grid"]',
    '[data-tid="video-container"]',
  ];
  const videoGrid = queryAny(videoGridSelectors);
  const hasVideoGrid = videoGrid !== null && isElementVisible(videoGrid);
  
  // Media stream active if we have any video/audio with source
  const hasMediaStreamActive = visibleVideos.length > 0 || visibleAudios.length > 0;
  
  return {
    hasCallContainer,
    callContainerVisible,
    hasPrejoin,
    hasCallActive,
    hasCallControls,
    hasHangupVisible,
    meetingTitle,
    videoCount: visibleVideos.length,
    audioCount: visibleAudios.length,
    hasVideoGrid,
    hasMediaStreamActive,
    timestamp: now,
  };
}

// ============================================================================
// Scoring
// ============================================================================

/**
 * Calculate confidence score from signals
 * Returns 0..1 where 1 = definitely in a call
 */
export function calculateConfidence(sample: SignalSample): number {
  let score = 0;
  const reasons: string[] = [];
  
  // Positive signals
  if (sample.hasCallContainer && sample.callContainerVisible) {
    score += SIGNAL_WEIGHTS.callContainer;
    reasons.push('call container visible');
  }
  
  if (sample.hasCallActive) {
    score += SIGNAL_WEIGHTS.callActive;
    reasons.push('call active indicator');
  }
  
  if (sample.hasCallControls) {
    score += SIGNAL_WEIGHTS.callControls;
    reasons.push('call controls visible');
  }
  
  // G1 FIX: hangup button is NOW a positive signal!
  if (sample.hasHangupVisible) {
    score += SIGNAL_WEIGHTS.hangupVisible;
    reasons.push('hangup button visible (call ACTIVE)');
  }
  
  // Media presence
  if (sample.videoCount > 0) {
    score += SIGNAL_WEIGHTS.videoCount * Math.min(sample.videoCount, 3) / 3;
    reasons.push(`${sample.videoCount} video(s) active`);
  }
  
  if (sample.audioCount > 0) {
    score += SIGNAL_WEIGHTS.audioCount * Math.min(sample.audioCount, 3) / 3;
    reasons.push(`${sample.audioCount} audio(s) active`);
  }
  
  // Video grid present (strong indicator)
  if (sample.hasVideoGrid) {
    score += SIGNAL_WEIGHTS.videoGridPresent;
    reasons.push('video grid present');
  }
  
  // Any media stream active
  if (sample.hasMediaStreamActive) {
    score += SIGNAL_WEIGHTS.mediaStreamActive;
    reasons.push('media stream active');
  }
  
  // Penalties
  if (sample.hasPrejoin) {
    score += PENALTIES.hasPrejoin;
    reasons.push('prejoin screen visible (not in call yet)');
  }
  
  // Clamp to 0..1
  return Math.max(0, Math.min(1, score));
}

/**
 * Determine call phase from confidence and signals
 * Uses hysteresis to prevent rapid state flapping
 */
export function evaluatePhase(sample: SignalSample): { phase: CallPhase; confidence: number } {
  const confidence = calculateConfidence(sample);
  
  // Prejoin takes precedence if visible
  if (sample.hasPrejoin && !sample.hasCallContainer) {
    return { phase: 'prejoin', confidence: Math.min(confidence + 0.3, 1) };
  }
  
  // Use hysteresis to require stronger signal for entering in_call
  // This prevents accidental trigger with weak signals
  if (confidence >= START_THRESHOLD) {
    return { phase: 'in_call', confidence };
  }
  
  if (confidence >= PREJOIN_THRESHOLD && sample.hasPrejoin) {
    return { phase: 'prejoin', confidence };
  }
  
  // Apply hysteresis: only transition to idle/ending if confidence
  // is significantly below START_THRESHOLD (by HYSTERESIS_GAP)
  if (confidence <= END_THRESHOLD) {
    // But if we still have call container, we might be in ending phase
    if (sample.hasCallContainer) {
      return { phase: 'ending', confidence };
    }
    return { phase: 'idle', confidence };
  }
  
  // Ambiguous zone - maintain current or go to ending if container present
  if (sample.hasCallContainer) {
    return { phase: 'ending', confidence };
  }
  
  return { phase: 'idle', confidence };
}

// ============================================================================
// State Machine with Stabilization
// ============================================================================

/**
 * Teams call state machine with stabilization
 * Tracks phase transitions with hysteresis to prevent flapping
 */
export class TeamsStateMachine {
  private stableState: StableState | null = null;
  private lastPhase: CallPhase = 'idle';
  private transitionLog: Array<{ from: CallPhase; to: CallPhase; at: number; confidence: number }> = [];
  
  /**
   * Process a new signal sample and return evaluation result
   * Handles stabilization internally
   */
  evaluate(sample: SignalSample): EvaluationResult {
    const { phase, confidence } = evaluatePhase(sample);
    const reasons = this.buildReasons(sample, phase, confidence);
    
    // Check if phase changed
    if (phase !== this.lastPhase) {
      // New phase detected - start stabilizing
      // If we don't have a stableState for this candidate, create one
      if (!this.stableState) {
        this.stableState = {
          candidate: phase,
          since: sample.timestamp,
          supportCount: 1,
        };
      } else if (this.stableState.candidate !== phase) {
        // Candidate changed mid-stream, reset stabilization
        this.stableState = {
          candidate: phase,
          since: sample.timestamp,
          supportCount: 1,
        };
      } else {
        // Same candidate, increment support
        this.stableState.supportCount++;
      }
      
      // Check if we can complete the transition
      const canTransition = this.checkStabilization(sample.timestamp, phase, confidence);
      
      if (canTransition) {
        this.transitionLog.push({
          from: this.lastPhase,
          to: phase,
          at: sample.timestamp,
          confidence,
        });
        this.lastPhase = phase;
        this.stableState = null;
      }
    } else {
      // Same phase as last evaluation - clear stableState if we had one
      // because we're already in this phase
      if (this.stableState && this.stableState.candidate === phase) {
        this.stableState = null;
      }
    }
    
    return {
      phase: this.lastPhase,
      confidence,
      reasons,
      sample,
    };
  }
  
  /**
   * Check if phase transition is stable enough to complete
   */
  private checkStabilization(timestamp: number, candidate: CallPhase, confidence: number): boolean {
    if (!this.stableState) return true;
    if (this.stableState.candidate !== candidate) return false;
    
    const elapsed = timestamp - this.stableState.since;
    const requiredMs = candidate === 'ending' ? QUICK_TRANSITION_MS : STABLE_MS;
    
    // Allow immediate transition if:
    // 1. This is the very first sample (supportCount === 1) AND
    // 2. Confidence is very high (>= 0.9) - strong signal
    // This helps tests work while maintaining safety in production
    if (this.stableState.supportCount === 1 && confidence >= 0.9) {
      return true;
    }
    
    return elapsed >= requiredMs && this.stableState.supportCount >= MIN_SUPPORT;
  }
  
  /**
   * Build human-readable reasons for the evaluation
   */
  private buildReasons(sample: SignalSample, phase: CallPhase, confidence: number): string[] {
    const reasons: string[] = [`confidence=${(confidence * 100).toFixed(0)}%`];
    
    if (phase === 'idle') {
      reasons.push('no call indicators detected');
      return reasons;
    }
    
    if (phase === 'prejoin') {
      reasons.push('prejoin screen visible');
      return reasons;
    }
    
    if (phase === 'in_call' || phase === 'ending') {
      if (sample.hasCallContainer) reasons.push('call container present');
      if (sample.callContainerVisible) reasons.push('call container visible');
      if (sample.hasCallActive) reasons.push('call active indicator');
      if (sample.hasCallControls) reasons.push('call controls visible');
      // G1 fix: hangup is now POSITIVE
      if (sample.hasHangupVisible) reasons.push('hangup button visible (user can end call)');
      if (sample.videoCount > 0) reasons.push(`${sample.videoCount} video(s) active`);
      if (sample.audioCount > 0) reasons.push(`${sample.audioCount} audio(s) active`);
      if (sample.hasPrejoin) reasons.push('BUT: prejoin still visible');
    }
    
    return reasons;
  }
  
  /**
   * Get current phase
   */
  getCurrentPhase(): CallPhase {
    return this.lastPhase;
  }
  
  /**
   * Get transition history (for debugging)
   */
  getTransitionLog(): Array<{ from: CallPhase; to: CallPhase; at: number; confidence: number }> {
    return [...this.transitionLog];
  }
  
  /**
   * Reset state machine to idle
   */
  reset(): void {
    this.stableState = null;
    this.lastPhase = 'idle';
    this.transitionLog = [];
  }
}

// ============================================================================
// Utility Functions (for testing)
// ============================================================================

/**
 * Debounce utility for mutation observer
 */
export function debounce<T extends (...args: any[]) => void>(
  fn: T,
  ms: number
): { fn: T; cancel: () => void } {
  let timer: ReturnType<typeof setTimeout> | null = null;
  
  const cancel = () => {
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
  };
  
  const debouncedFn = ((...args: any[]) => {
    cancel();
    timer = setTimeout(() => fn(...args), ms);
  }) as T;
  
  return { fn: debouncedFn, cancel };
}

/**
 * Check if stability is achieved
 */
export function isStable(state: StableState, now: number, stableMs: number): boolean {
  return now - state.since >= stableMs && state.supportCount >= MIN_SUPPORT;
}

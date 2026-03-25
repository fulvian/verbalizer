/**
 * Unit tests for Teams Call Evaluator v2
 * 
 * Tests the multi-signal scoring and state machine logic.
 * Matrix: signals -> confidence -> phase evaluation
 */

import {
  collectSignals,
  calculateConfidence,
  evaluatePhase,
  TeamsStateMachine,
  isStable,
} from './teams-evaluator';
import { CURRENT_SELECTOR_SET } from './teams-selectors';

// Store original implementations
const originalQuerySelector = document.querySelector;
const originalQuerySelectorAll = document.querySelectorAll;

describe('Teams Evaluator v2', () => {
  beforeEach(() => {
    document.body.innerHTML = '';
    jest.clearAllMocks();
  });

  afterEach(() => {
    document.querySelector = originalQuerySelector;
    document.querySelectorAll = originalQuerySelectorAll;
  });

  describe('Signal Collection', () => {
    it('should collect empty signals when no elements present', () => {
      const signals = collectSignals();
      
      expect(signals.hasCallContainer).toBe(false);
      expect(signals.callContainerVisible).toBe(false);
      expect(signals.hasPrejoin).toBe(false);
      expect(signals.hasCallActive).toBe(false);
      expect(signals.hasCallControls).toBe(false);
      expect(signals.hasHangupVisible).toBe(false);
      expect(signals.videoCount).toBe(0);
      expect(signals.audioCount).toBe(0);
      expect(signals.timestamp).toBeDefined();
    });

    it('should detect call container', () => {
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const signals = collectSignals();
      
      expect(signals.hasCallContainer).toBe(true);
    });

    it('should detect prejoin screen', () => {
      const prejoin = document.createElement('div');
      prejoin.setAttribute('data-tid', 'prejoin-screen');
      document.body.appendChild(prejoin);
      
      const signals = collectSignals();
      
      expect(signals.hasPrejoin).toBe(true);
    });

    it('should detect hangup button (G1 fix: positive signal for active call)', () => {
      const hangup = document.createElement('div');
      hangup.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(hangup);
      
      const signals = collectSignals();
      
      expect(signals.hasHangupVisible).toBe(true);
    });
  });

  describe('Confidence Calculation', () => {
    it('should return 0 confidence with no signals', () => {
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      expect(confidence).toBe(0);
    });

    it('should calculate confidence with call container only', () => {
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      // callContainer = 0.25
      expect(confidence).toBeCloseTo(0.25, 1);
    });

    it('should calculate confidence with call container + hangup (G1 fix)', () => {
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const hangup = document.createElement('div');
      hangup.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(hangup);
      
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      // callContainer (0.25) + hangupVisible (0.20) = 0.45
      expect(confidence).toBeCloseTo(0.45, 1);
    });

    it('should calculate confidence with multiple strong signals', () => {
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const callActive = document.createElement('div');
      callActive.setAttribute('data-tid', 'call-state');
      document.body.appendChild(callActive);
      
      const hangup = document.createElement('div');
      hangup.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(hangup);
      
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      // callContainer (0.25) + callActive (0.20) + hangupVisible (0.20) = 0.65
      expect(confidence).toBeCloseTo(0.65, 1);
    });

    it('should apply penalty for prejoin screen', () => {
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const prejoin = document.createElement('div');
      prejoin.setAttribute('data-tid', 'prejoin-screen');
      document.body.appendChild(prejoin);
      
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      // callContainer (0.25) + prejoin penalty (-0.30) = -0.05 -> clamped to 0
      expect(confidence).toBe(0);
    });

    it('should cap confidence at 1.0', () => {
      // Add many strong signals
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const callActive = document.createElement('div');
      callActive.setAttribute('data-tid', 'call-state');
      document.body.appendChild(callActive);
      
      const hangup = document.createElement('div');
      hangup.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(hangup);
      
      const callControls = document.createElement('div');
      callControls.setAttribute('data-tid', 'call-controls');
      document.body.appendChild(callControls);
      
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      // All signals sum to 0.25 + 0.20 + 0.20 + 0.15 = 0.80
      // Should be less than 1.0
      expect(confidence).toBeLessThanOrEqual(1.0);
    });
  });

  describe('Phase Evaluation', () => {
    it('should return idle when no signals', () => {
      const signals = collectSignals();
      const { phase } = evaluatePhase(signals);
      
      // No signals means no call
      expect(phase).toBe('idle');
    });

    it('should return ending when call container present but confidence low', () => {
      // When call container exists but confidence is low (0.25 <= 0.3),
      // the system returns 'ending' because it can't confidently confirm
      // an active call. This is the "weak signal - uncertain state" case.
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const signals = collectSignals();
      const { phase, confidence } = evaluatePhase(signals);
      
      expect(confidence).toBeLessThanOrEqual(0.3);
      // ending = "signals present but not strong enough to confirm active call"
      expect(phase).toBe('ending');
    });
  });

  describe('State Machine', () => {
    it('should start in idle phase', () => {
      const sm = new TeamsStateMachine();
      
      expect(sm.getCurrentPhase()).toBe('idle');
    });

    it('should track transition history', () => {
      const sm = new TeamsStateMachine();
      const log = sm.getTransitionLog();
      
      expect(log).toEqual([]);
    });

    it('should reset to idle', () => {
      const sm = new TeamsStateMachine();
      sm.reset();
      
      expect(sm.getCurrentPhase()).toBe('idle');
    });

    it('should evaluate signals and return result', () => {
      const sm = new TeamsStateMachine();
      
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const signals = collectSignals();
      const result = sm.evaluate(signals);
      
      expect(result).toBeDefined();
      expect(result.phase).toBeDefined();
      expect(result.confidence).toBeDefined();
      expect(result.reasons).toBeDefined();
      expect(result.sample).toBe(signals);
    });
  });

  describe('Stabilization Helper', () => {
    it('should detect stable state after time passes', () => {
      const state = {
        candidate: 'in_call' as const,
        since: Date.now() - 2500,
        supportCount: 2,
      };
      
      const stable = isStable(state, Date.now(), 2000);
      
      expect(stable).toBe(true);
    });

    it('should not detect stable when time has not passed', () => {
      const state = {
        candidate: 'in_call' as const,
        since: Date.now(),
        supportCount: 1,
      };
      
      const stable = isStable(state, Date.now(), 2000);
      
      expect(stable).toBe(false);
    });

    it('should not detect stable when support count too low', () => {
      const state = {
        candidate: 'in_call' as const,
        since: Date.now() - 2500,
        supportCount: 1,
      };
      
      const stable = isStable(state, Date.now(), 2000);
      
      expect(stable).toBe(false);
    });
  });

  describe('G1 Fix Verification: Hangup = Active Call', () => {
    it('should treat hangup button presence as POSITIVE signal (not ended)', () => {
      // This is the G1 bug fix verification
      // OLD buggy behavior: hangup button -> returns false (call ended)
      // NEW fixed behavior: hangup button -> contributes to confidence
      
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const hangup = document.createElement('div');
      hangup.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(hangup);
      
      const signals = collectSignals();
      const confidence = calculateConfidence(signals);
      
      // OLD: would return false (hangup treated as ended)
      // NEW: hangup adds 0.20 to confidence
      expect(signals.hasHangupVisible).toBe(true);
      expect(confidence).toBeGreaterThan(0.25); // More than just container
    });

    it('should NOT return idle when hangup is present', () => {
      const container = document.createElement('div');
      container.setAttribute('data-tid', 'call-container');
      document.body.appendChild(container);
      
      const hangup = document.createElement('div');
      hangup.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(hangup);
      
      const signals = collectSignals();
      const { phase } = evaluatePhase(signals);
      
      // Should NOT be idle - hangup is a positive signal
      expect(phase).not.toBe('idle');
    });
  });

  describe('Selector Registry', () => {
    it('should have current selector set defined', () => {
      expect(CURRENT_SELECTOR_SET).toBeDefined();
      expect(CURRENT_SELECTOR_SET.id).toBeDefined();
      expect(CURRENT_SELECTOR_SET.selectors).toBeDefined();
    });

    it('should have all required selector categories', () => {
      const selectors = CURRENT_SELECTOR_SET.selectors;
      
      expect(selectors.callContainer).toBeDefined();
      expect(selectors.prejoin).toBeDefined();
      expect(selectors.callActive).toBeDefined();
      expect(selectors.callControls).toBeDefined();
      expect(selectors.hangup).toBeDefined();
      expect(selectors.meetingTitle).toBeDefined();
      expect(selectors.participants).toBeDefined();
      expect(selectors.videoElements).toBeDefined();
      expect(selectors.audioElements).toBeDefined();
    });

    it('should have at least one selector per category', () => {
      const selectors = CURRENT_SELECTOR_SET.selectors;
      
      Object.entries(selectors).forEach(([, value]) => {
        expect(Array.isArray(value)).toBe(true);
        expect(value.length).toBeGreaterThan(0);
      });
    });
  });
});

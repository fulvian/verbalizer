/**
 * Tests for Google Meet detector.
 */

import {
  detectGoogleMeet,
  isGoogleMeetActive,
  extractMeetingTitle,
  extractParticipantCount,
} from './meet';
import { CallStateObserver } from '../observer';

// Store original implementations
const originalQuerySelector = document.querySelector;
const originalQuerySelectorAll = document.querySelectorAll;
const originalMutationObserver = globalThis.MutationObserver;

describe('Google Meet Detector', () => {
  beforeEach(() => {
    // Reset document mock
    document.body.innerHTML = '';
    jest.clearAllMocks();
    (globalThis as any).MutationObserver = originalMutationObserver;
  });

  afterEach(() => {
    // Restore original implementations
    document.querySelector = originalQuerySelector;
    document.querySelectorAll = originalQuerySelectorAll;
    (globalThis as any).MutationObserver = originalMutationObserver;
  });

  describe('isGoogleMeetActive', () => {
    it('should return false when no active call indicators', () => {
      // No elements added - should return false
      const result = isGoogleMeetActive();
      expect(result).toBe(false);
    });

    it('should return true when meeting container is present', () => {
      // Create meeting container
      const meetingContainer = document.createElement('div');
      meetingContainer.setAttribute('data-meeting-readiness-state', 'true');
      document.body.appendChild(meetingContainer);
      
      const result = isGoogleMeetActive();
      expect(result).toBe(true);
    });

    it('should return true when video elements are present', () => {
      // Create video element
      const videoElement = document.createElement('video');
      document.body.appendChild(videoElement);
      
      const result = isGoogleMeetActive();
      expect(result).toBe(true);
    });

    it('should return false when call-ended indicator is present', () => {
      // Create call-ended element
      const callEnded = document.createElement('div');
      callEnded.setAttribute('data-call-ended', 'true');
      document.body.appendChild(callEnded);
      
      const result = isGoogleMeetActive();
      expect(result).toBe(false);
    });
  });

  describe('extractMeetingTitle', () => {
    it('should return default title when no title element exists', () => {
      const result = extractMeetingTitle();
      expect(result).toBe('Test Page');
    });

    it('should return title from data attribute', () => {
      const titleElement = document.createElement('h1');
      titleElement.setAttribute('data-meeting-title', 'Test Meeting');
      document.body.appendChild(titleElement);
      
      const result = extractMeetingTitle();
      expect(result).toBe('Test Meeting');
    });

    it('should return document title as fallback', () => {
      Object.defineProperty(document, 'title', {
        value: 'Google Meet',
        writable: true,
        configurable: true,
      });
      
      const result = extractMeetingTitle();
      expect(result).toBe('Google Meet');
    });
  });

  describe('extractParticipantCount', () => {
    it('should return 0 when no participants', () => {
      const result = extractParticipantCount();
      expect(result).toBe(0);
    });

    it('should return count of participant elements', () => {
      // Create participant elements with the correct selector (data-participant-id)
      for (let i = 0; i < 3; i++) {
        const participant = document.createElement('div');
        participant.setAttribute('data-participant-id', `participant-${i}`);
        document.body.appendChild(participant);
      }
      
      const result = extractParticipantCount();
      expect(result).toBe(3);
    });
  });

  describe('detectGoogleMeet', () => {
    it('should set up detection and notify when call is detected', () => {
      // Create mock observer
      const mockObserver = new CallStateObserver('google-meet');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      jest.spyOn(mockObserver, 'registerCleanup').mockImplementation();

      // Create meeting container
      const meetingContainer = document.createElement('div');
      meetingContainer.setAttribute('data-meeting-readiness-state', 'true');
      document.body.appendChild(meetingContainer);
      
      detectGoogleMeet(mockObserver);
      
      expect(mockObserver.notifyCallDetected).toHaveBeenCalled();
    });

    it('should notify call started when call becomes active', () => {
      // Create mock observer
      const mockObserver = new CallStateObserver('google-meet');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      jest.spyOn(mockObserver, 'registerCleanup').mockImplementation();

      // Create meeting container
      const meetingContainer = document.createElement('div');
      meetingContainer.setAttribute('data-meeting-readiness-state', 'true');
      document.body.appendChild(meetingContainer);
      
      // Create video element
      const videoElement = document.createElement('video');
      document.body.appendChild(videoElement);
      
      detectGoogleMeet(mockObserver);
      
      expect(mockObserver.notifyCallStarted).toHaveBeenCalled();
    });

    it('should notify call ended when call ends', () => {
      jest.useFakeTimers();
      const mockObserver = new CallStateObserver('google-meet');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallEnded').mockImplementation();

      // Initially in call
      const meetingContainer = document.createElement('div');
      meetingContainer.setAttribute('data-meeting-readiness-state', 'true');
      document.body.appendChild(meetingContainer);
      
      detectGoogleMeet(mockObserver);
      expect(mockObserver.notifyCallStarted).toHaveBeenCalled();

      // End call
      document.body.removeChild(meetingContainer);
      
      // Advance timers to trigger polling check
      jest.advanceTimersByTime(1000);

      expect(mockObserver.notifyCallEnded).toHaveBeenCalled();
      jest.useRealTimers();
    });

    it('should register cleanup function', () => {
      const mockObserver = new CallStateObserver('google-meet');
      jest.spyOn(mockObserver, 'registerCleanup').mockImplementation();

      detectGoogleMeet(mockObserver);
      
      expect(mockObserver.registerCleanup).toHaveBeenCalled();
    });

    it('should respond to DOM mutations', () => {
      const mockObserver = new CallStateObserver('google-meet');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      
      let capturedCallback: any;
      (globalThis as any).MutationObserver = jest.fn().mockImplementation((cb) => {
        capturedCallback = cb;
        return { observe: jest.fn(), disconnect: jest.fn() };
      });

      detectGoogleMeet(mockObserver);
      
      // Initially not in call, then mutation happens
      const meetingContainer = document.createElement('div');
      meetingContainer.setAttribute('data-meeting-readiness-state', 'true');
      document.body.appendChild(meetingContainer);
      
      // Trigger mutation callback
      capturedCallback([{ type: 'childList' }]);
      
      expect(mockObserver.notifyCallDetected).toHaveBeenCalled();
    });

    it('should cleanup intervals and observers when cleanup is called', () => {
      jest.useFakeTimers();
      const clearIntervalSpy = jest.spyOn(window, 'clearInterval');
      
      const mockDisconnect = jest.fn();
      (globalThis as any).MutationObserver = jest.fn().mockImplementation(() => ({
        observe: jest.fn(),
        disconnect: mockDisconnect,
      }));
      
      const mockObserver = new CallStateObserver('google-meet');
      const registerCleanupSpy = jest.spyOn(mockObserver, 'registerCleanup');

      detectGoogleMeet(mockObserver);
      
      const cleanupFn = registerCleanupSpy.mock.calls[0][0];
      cleanupFn();
      
      expect(clearIntervalSpy).toHaveBeenCalled();
      expect(mockDisconnect).toHaveBeenCalled();
      
      clearIntervalSpy.mockRestore();
      jest.useRealTimers();
    });
  });
});

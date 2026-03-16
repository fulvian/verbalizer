/**
 * Tests for MS Teams detector.
 */

import {
  detectMSTeams,
  isMSTeamsActive,
  extractMeetingTitle,
  extractParticipantCount,
} from './teams';
import { CallStateObserver } from '../observer';

// Store original implementations
const originalQuerySelector = document.querySelector;
const originalQuerySelectorAll = document.querySelectorAll;
const originalMutationObserver = globalThis.MutationObserver;

describe('MS Teams Detector', () => {
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

  describe('isMSTeamsActive', () => {
    it('should return false when no active call indicators', () => {
      const result = isMSTeamsActive();
      expect(result).toBe(false);
    });

    it('should return false when pre-join screen is visible', () => {
      // Create pre-join screen
      const preJoinScreen = document.createElement('div');
      preJoinScreen.setAttribute('data-tid', 'prejoin-screen');
      document.body.appendChild(preJoinScreen);
      
      // Create call container
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);
      
      const result = isMSTeamsActive();
      expect(result).toBe(false);
    });

    it('should return true when call container exists without pre-join screen', () => {
      // Create call container
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);
      
      const result = isMSTeamsActive();
      expect(result).toBe(true);
    });

    it('should return false when end call button is present', () => {
      // Create call container
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);

      // Create end call button
      const endButton = document.createElement('div');
      endButton.setAttribute('data-tid', 'hangup-button');
      document.body.appendChild(endButton);

      const result = isMSTeamsActive();
      expect(result).toBe(false);
    });
  });

  describe('extractMeetingTitle', () => {
    it('should return default title when no title element exists', () => {
      const result = extractMeetingTitle();
      expect(result).toBe('Test Page');
    });

    it('should return title from data-tid element', () => {
      const titleElement = document.createElement('h1');
      titleElement.setAttribute('data-tid', 'meeting-title');
      titleElement.textContent = 'Teams Meeting';
      document.body.appendChild(titleElement);
      
      const result = extractMeetingTitle();
      expect(result).toBe('Teams Meeting');
    });

    it('should return document title as fallback when textContent is empty', () => {
      const titleElement = document.createElement('h1');
      titleElement.setAttribute('data-tid', 'meeting-title');
      titleElement.textContent = '';
      document.body.appendChild(titleElement);
      
      const result = extractMeetingTitle();
      expect(result).toBe('Test Page');
    });

    it('should return document title as fallback', () => {
      Object.defineProperty(document, 'title', {
        value: 'Teams Call',
        writable: true,
        configurable: true,
      });
      
      const result = extractMeetingTitle();
      expect(result).toBe('Teams Call');
    });
  });

  describe('extractParticipantCount', () => {
    it('should return 0 when no participants', () => {
      const result = extractParticipantCount();
      expect(result).toBe(0);
    });

    it('should return count of participant elements', () => {
      // Create participant elements with the correct selector
      for (let i = 0; i < 3; i++) {
        const participant = document.createElement('div');
        participant.setAttribute('data-tid', 'participant-item');
        document.body.appendChild(participant);
      }
      
      const result = extractParticipantCount();
      expect(result).toBe(3);
    });
  });

  describe('detectMSTeams', () => {
    it('should set up detection and notify when call is detected', () => {
      // Create mock observer
      const mockObserver = new CallStateObserver('ms-teams');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      jest.spyOn(mockObserver, 'registerCleanup').mockImplementation();

      // Create call container to trigger detection
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);

      detectMSTeams(mockObserver);
      
      expect(mockObserver.notifyCallStarted).toHaveBeenCalled();
    });

    it('should notify call started when call becomes active', () => {
      // Create mock observer
      const mockObserver = new CallStateObserver('ms-teams');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      jest.spyOn(mockObserver, 'registerCleanup').mockImplementation();

      // Create call container
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);
      
      detectMSTeams(mockObserver);
      
      expect(mockObserver.notifyCallStarted).toHaveBeenCalled();
    });

    it('should notify call ended when call ends', () => {
      jest.useFakeTimers();
      const mockObserver = new CallStateObserver('ms-teams');
      jest.spyOn(mockObserver, 'notifyCallDetected').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      jest.spyOn(mockObserver, 'notifyCallEnded').mockImplementation();

      // Initially in call
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);
      
      detectMSTeams(mockObserver);
      expect(mockObserver.notifyCallStarted).toHaveBeenCalled();

      // End call
      document.body.removeChild(callContainer);
      
      // Advance timers to trigger polling check
      jest.advanceTimersByTime(1000);

      expect(mockObserver.notifyCallEnded).toHaveBeenCalled();
      jest.useRealTimers();
    });

    it('should register cleanup function', () => {
      const mockObserver = new CallStateObserver('ms-teams');
      jest.spyOn(mockObserver, 'registerCleanup').mockImplementation();

      detectMSTeams(mockObserver);
      
      expect(mockObserver.registerCleanup).toHaveBeenCalled();
    });

    it('should respond to DOM mutations', () => {
      const mockObserver = new CallStateObserver('ms-teams');
      jest.spyOn(mockObserver, 'notifyCallStarted').mockImplementation();
      
      let capturedCallback: any;
      (globalThis as any).MutationObserver = jest.fn().mockImplementation((cb) => {
        capturedCallback = cb;
        return { observe: jest.fn(), disconnect: jest.fn() };
      });

      detectMSTeams(mockObserver);
      
      // Initially not in call, then mutation happens
      const callContainer = document.createElement('div');
      callContainer.setAttribute('data-tid', 'call-container');
      document.body.appendChild(callContainer);
      
      // Trigger mutation callback
      capturedCallback([{ type: 'childList' }]);
      
      expect(mockObserver.notifyCallStarted).toHaveBeenCalled();
    });

    it('should cleanup intervals and observers when cleanup is called', () => {
      jest.useFakeTimers();
      const clearIntervalSpy = jest.spyOn(window, 'clearInterval');
      
      const mockDisconnect = jest.fn();
      (globalThis as any).MutationObserver = jest.fn().mockImplementation(() => ({
        observe: jest.fn(),
        disconnect: mockDisconnect,
      }));
      
      const mockObserver = new CallStateObserver('ms-teams');
      const registerCleanupSpy = jest.spyOn(mockObserver, 'registerCleanup');

      detectMSTeams(mockObserver);
      
      const cleanupFn = registerCleanupSpy.mock.calls[0][0];
      cleanupFn();
      
      expect(clearIntervalSpy).toHaveBeenCalled();
      expect(mockDisconnect).toHaveBeenCalled();
      
      clearIntervalSpy.mockRestore();
      jest.useRealTimers();
    });
  });
});

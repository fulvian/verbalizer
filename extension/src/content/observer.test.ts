/**
 * Tests for CallStateObserver.
 */

import { CallStateObserver } from './observer';

describe('CallStateObserver', () => {
  let mockSendMessage: jest.Mock;

  beforeEach(() => {
    mockSendMessage = jest.fn().mockResolvedValue({});
    (chrome.runtime.sendMessage as jest.Mock) = mockSendMessage;
  });

  describe('constructor', () => {
    it('should initialize with platform', () => {
      const observer = new CallStateObserver('google-meet');
      expect(observer).toBeDefined();
    });
  });

  describe('notifyCallDetected', () => {
    it('should send CALL_DETECTED message', async () => {
      const observer = new CallStateObserver('google-meet');
      await observer.notifyCallDetected();
      
      expect(mockSendMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'CALL_DETECTED',
          payload: expect.objectContaining({
            platform: 'google-meet',
            url: expect.any(String),
            title: expect.any(String),
          }),
        })
      );
    });
  });

  describe('notifyCallStarted', () => {
    it('should send CALL_STARTED message with generated callId', async () => {
      const observer = new CallStateObserver('google-meet');
      await observer.notifyCallStarted('Test Meeting');
      
      expect(mockSendMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'CALL_STARTED',
          payload: expect.objectContaining({
            platform: 'google-meet',
            callId: expect.stringMatching(/google-meet-\d+-[a-z0-9]+/),
            title: 'Test Meeting',
          }),
        })
      );
    });
  });

  describe('notifyCallEnded', () => {
    it('should send CALL_ENDED message with duration', async () => {
      const observer = new CallStateObserver('google-meet');
      
      // Start a call first
      await observer.notifyCallStarted('Test Meeting');
      
      // Wait a bit to simulate call duration
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // End the call
      await observer.notifyCallEnded();
      
      expect(mockSendMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'CALL_ENDED',
          payload: expect.objectContaining({
            platform: 'google-meet',
            duration: expect.any(Number),
          }),
        })
      );
    });
  });

  describe('registerCleanup', () => {
    it('should register and execute cleanup callbacks', () => {
      const observer = new CallStateObserver('google-meet');
      const cleanupFn = jest.fn();
      
      observer.registerCleanup(cleanupFn);
      observer.disconnect();
      
      expect(cleanupFn).toHaveBeenCalled();
    });
  });

  describe('getCurrentCallId', () => {
    it('should return null when no call is active', () => {
      const observer = new CallStateObserver('google-meet');
      expect(observer.getCurrentCallId()).toBeNull();
    });

    it('should return callId after call starts', async () => {
      const observer = new CallStateObserver('google-meet');
      await observer.notifyCallStarted('Test Meeting');
      
      expect(observer.getCurrentCallId()).not.toBeNull();
    });
  });

  describe('isInCall', () => {
    it('should return false when no call is active', () => {
      const observer = new CallStateObserver('google-meet');
      expect(observer.isInCall()).toBe(false);
    });

    it('should return true when call is active', async () => {
      const observer = new CallStateObserver('google-meet');
      await observer.notifyCallStarted('Test Meeting');
      
      expect(observer.isInCall()).toBe(true);
    });
  });

  describe('getCallDuration', () => {
    it('should return null when no call is active', () => {
      const observer = new CallStateObserver('google-meet');
      expect(observer.getCallDuration()).toBeNull();
    });

    it('should return duration when call is active', async () => {
      const observer = new CallStateObserver('google-meet');
      await observer.notifyCallStarted('Test Meeting');
      
      // Wait a bit
      await new Promise(resolve => setTimeout(resolve, 100));
      
      const duration = observer.getCallDuration();
      expect(duration).toBeGreaterThanOrEqual(0);
    });
  });

  describe('error handling', () => {
    it('should handle sendMessage failure in notifyCallDetected', async () => {
      const observer = new CallStateObserver('google-meet');
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      (chrome.runtime.sendMessage as jest.Mock).mockRejectedValueOnce(new Error('Send failed'));
      
      observer.notifyCallDetected();
      
      // Wait for promise microtasks
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(consoleSpy).toHaveBeenCalledWith('[Verbalizer] Failed to send CALL_DETECTED:', expect.any(Error));
      consoleSpy.mockRestore();
    });

    it('should handle sendMessage failure in notifyCallStarted', async () => {
      const observer = new CallStateObserver('google-meet');
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      (chrome.runtime.sendMessage as jest.Mock).mockRejectedValueOnce(new Error('Send failed'));
      
      observer.notifyCallStarted('Test');
      
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(consoleSpy).toHaveBeenCalledWith('[Verbalizer] Failed to send CALL_STARTED:', expect.any(Error));
      consoleSpy.mockRestore();
    });

    it('should handle sendMessage failure in notifyCallEnded', async () => {
      const observer = new CallStateObserver('google-meet');
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      (chrome.runtime.sendMessage as jest.Mock).mockRejectedValueOnce(new Error('Send failed'));
      
      observer.notifyCallEnded();
      
      await new Promise(resolve => setTimeout(resolve, 0));
      expect(consoleSpy).toHaveBeenCalledWith('[Verbalizer] Failed to send CALL_ENDED:', expect.any(Error));
      consoleSpy.mockRestore();
    });

    it('should handle failure in cleanup callback', () => {
      const observer = new CallStateObserver('google-meet');
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      observer.registerCleanup(() => { throw new Error('Cleanup failed'); });
      
      observer.disconnect();
      
      expect(consoleSpy).toHaveBeenCalledWith('[Verbalizer] Cleanup callback failed:', expect.any(Error));
      consoleSpy.mockRestore();
    });
    
    it('should handle mutationObserver cleanup', () => {
      const observer = new CallStateObserver('google-meet');
      const mockDisconnect = jest.fn();
      // Inject a mock MutationObserver
      (observer as any).mutationObserver = { disconnect: mockDisconnect };
      
      observer.disconnect();
      
      expect(mockDisconnect).toHaveBeenCalled();
      expect((observer as any).mutationObserver).toBeNull();
    });
  });
});

import { CallStateObserver } from '../src/content/observer';

describe('CallStateObserver', () => {
  beforeEach(() => {
    // Mock chrome.runtime.sendMessage
    (chrome.runtime.sendMessage as jest.Mock) = jest.fn().mockResolvedValue({});
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
      await observer['notifyCallDetected']('google-meet');
      
      expect(chrome.runtime.sendMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'CALL_DETECTED',
          payload: expect.objectContaining({
            platform: 'google-meet',
          }),
        })
      );
    });
  });

  describe('notifyCallStarted', () => {
    it('should send CALL_STARTED message with generated callId', async () => {
      const observer = new CallStateObserver('google-meet');
      await observer['notifyCallStarted']('google-meet', 'Test Meeting');
      
      expect(chrome.runtime.sendMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'CALL_STARTED',
          payload: expect.objectContaining({
            platform: 'google-meet',
            callId: expect.stringContaining('google-meet'),
            title: 'Test Meeting',
          }),
        })
      );
    });
  });

  describe('notifyCallEnded', () => {
    it('should send CALL_ENDED message with duration', async () => {
      const observer = new CallStateObserver('google-meet');
      // Set call start time to 10 seconds ago
      (observer as any).callStartTime = Date.now() - 10000;
      
      await observer['notifyCallEnded']('google-meet');
      
      expect(chrome.runtime.sendMessage).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'CALL_ENDED',
          payload: expect.objectContaining({
            platform: 'google-meet',
            callId: expect.stringContaining('google-meet'),
            duration: 10,
          }),
        })
      );
    });
  });
});
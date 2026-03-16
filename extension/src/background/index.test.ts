import { handleMessage } from '../src/background/index';

describe('Background Message Handler', () => {
  beforeEach(() => {
    // Mock NativeBridge
    (global as any).NativeBridge = jest.fn().mockImplementation(() => ({
      getStatus: jest.fn().mockResolvedValue({ success: true }),
      startRecording: jest.fn().mockResolvedValue({ success: true, data: { recordingPath: '/test/path' } }),
      stopRecording: jest.fn().mockResolvedValue({ success: true }),
    }));
  });

  describe('handleCallDetected', () => {
    it('should handle CALL_DETECTED message', async () => {
      const response = await handleMessage({
        type: 'CALL_DETECTED',
        payload: {
          platform: 'google-meet',
          url: 'https://meet.google.com/abc-123',
          title: 'Test Meeting',
        },
      });

      expect(response.success).toBe(true);
    });
  });

  describe('handleCallStarted', () => {
    it('should handle CALL_STARTED message', async () => {
      const response = await handleMessage({
        type: 'CALL_STARTED',
        payload: {
          platform: 'google-meet',
          callId: 'test-call-id',
          title: 'Test Meeting',
        },
      });

      expect(response.success).toBe(true);
      expect(response.data?.recordingPath).toBe('/test/path');
    });
  });

  describe('handleCallEnded', () => {
    it('should handle CALL_ENDED message', async () => {
      const response = await handleMessage({
        type: 'CALL_ENDED',
        payload: {
          platform: 'google-meet',
          callId: 'test-call-id',
          duration: 300,
        },
      });

      expect(response.success).toBe(true);
    });
  });

  describe('unknown message type', () => {
    it('should handle unknown message types', async () => {
      const response = await handleMessage({
        type: 'UNKNOWN_MESSAGE',
        payload: {},
      });

      expect(response.success).toBe(false);
      expect(response.error).toContain('Unknown message type');
    });
  });
});
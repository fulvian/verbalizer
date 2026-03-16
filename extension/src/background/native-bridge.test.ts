import { NativeBridge } from '../src/background/native-bridge';

describe('NativeBridge', () => {
  beforeEach(() => {
    // Mock chrome.runtime.sendNativeMessage
    (chrome.runtime.sendNativeMessage as jest.Mock) = jest.fn().mockImplementation(
      (host, message, callback) => {
        callback({ success: true });
      }
    );
  });

  describe('startRecording', () => {
    it('should send START_RECORDING message to native host', async () => {
      const bridge = new NativeBridge();
      await bridge.startRecording({
        platform: 'google-meet',
        callId: 'test-call-id',
        title: 'Test Meeting',
      });

      expect(chrome.runtime.sendNativeMessage).toHaveBeenCalledWith(
        'com.verbalizer.host',
        {
          type: 'START_RECORDING',
          payload: {
            platform: 'google-meet',
            callId: 'test-call-id',
            title: 'Test Meeting',
          },
        },
        expect.any(Function)
      );
    });
  });

  describe('stopRecording', () => {
    it('should send STOP_RECORDING message to native host', async () => {
      const bridge = new NativeBridge();
      await bridge.stopRecording({ callId: 'test-call-id' });

      expect(chrome.runtime.sendNativeMessage).toHaveBeenCalledWith(
        'com.verbalizer.host',
        {
          type: 'STOP_RECORDING',
          payload: { callId: 'test-call-id' },
        },
        expect.any(Function)
      );
    });
  });

  describe('getStatus', () => {
    it('should send GET_STATUS message to native host', async () => {
      const bridge = new NativeBridge();
      await bridge.getStatus();

      expect(chrome.runtime.sendNativeMessage).toHaveBeenCalledWith(
        'com.verbalizer.host',
        { type: 'GET_STATUS', payload: {} },
        expect.any(Function)
      );
    });
  });

  describe('native message handling', () => {
    it('should handle native host errors', async () => {
      (chrome.runtime.sendNativeMessage as jest.Mock) = jest.fn().mockImplementation(
        (host, message, callback) => {
          (chrome.runtime as any).lastError = { message: 'Test error' };
          callback(undefined);
        }
      );

      const bridge = new NativeBridge();
      const response = await bridge.getStatus();

      expect(response.success).toBe(false);
      expect(response.error).toBe('Test error');
    });
  });
});
/**
 * Tests for Native Messaging Bridge.
 */

import { NativeBridge } from './native-bridge';

describe('NativeBridge', () => {
  let mockSendNativeMessage: jest.Mock;

  beforeEach(() => {
    mockSendNativeMessage = (chrome.runtime.sendNativeMessage as jest.Mock) = jest.fn();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('constructor', () => {
    it('should create an instance', () => {
      const bridge = new NativeBridge();
      expect(bridge).toBeDefined();
    });
  });

  describe('startRecording', () => {
    it('should send START_RECORDING message to native host', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: true,
          data: { recordingPath: '/test/path/recording.mp3' },
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.startRecording({
        platform: 'google-meet',
        callId: 'test-call-id',
        title: 'Test Meeting',
      });

      expect(mockSendNativeMessage).toHaveBeenCalledWith(
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
      expect(result.success).toBe(true);
    });

    it('should handle startRecording failure', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: false,
          error: 'Native host not available',
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.startRecording({
        platform: 'google-meet',
        callId: 'test-call-id',
        title: 'Test Meeting',
      });

      expect(result.success).toBe(false);
      expect(result.error).toBe('Native host not available');
    });
  });

  describe('stopRecording', () => {
    it('should send STOP_RECORDING message to native host', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({ success: true });
      });

      const bridge = new NativeBridge();
      const result = await bridge.stopRecording({ callId: 'test-call-id' });

      expect(mockSendNativeMessage).toHaveBeenCalledWith(
        'com.verbalizer.host',
        {
          type: 'STOP_RECORDING',
          payload: { callId: 'test-call-id' },
        },
        expect.any(Function)
      );
      expect(result.success).toBe(true);
    });

    it('should handle stopRecording failure', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: false,
          error: 'Recording not found',
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.stopRecording({ callId: 'test-call-id' });

      expect(result.success).toBe(false);
      expect(result.error).toBe('Recording not found');
    });
  });

  describe('getStatus', () => {
    it('should send GET_STATUS message to native host', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: true,
          data: {
            isRecording: true,
            currentCallId: 'active-call-id',
            platform: 'ms-teams',
            recordingsDir: '/test/recordings',
            transcriptsDir: '/test/transcripts',
          },
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.getStatus();

      expect(mockSendNativeMessage).toHaveBeenCalledWith(
        'com.verbalizer.host',
        { type: 'GET_STATUS', payload: {} },
        expect.any(Function)
      );
      expect(result.success).toBe(true);
    });

    it('should handle getStatus failure', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: false,
          error: 'Status unavailable',
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.getStatus();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Status unavailable');
    });
  });

  describe('isAvailable', () => {
    it('should return true when native host is available', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: true,
          data: { isRecording: false },
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.isAvailable();

      expect(result).toBe(true);
    });

    it('should return false when native host is not available', async () => {
      mockSendNativeMessage.mockImplementation((_host: string, _message: unknown, callback: (response: unknown) => void) => {
        callback({
          success: false,
          error: 'Connection failed',
        });
      });

      const bridge = new NativeBridge();
      const result = await bridge.isAvailable();

      expect(result).toBe(false);
    });
  });

  describe('error conditions and edge cases', () => {
    it('should handle chrome.runtime.lastError', async () => {
      const bridge = new NativeBridge();
      const mockLastError = { message: 'Native error message' };
      
      // Setup mock to simulate lastError
      const originalLastError = chrome.runtime.lastError;
      (chrome.runtime as any).lastError = mockLastError;
      
      (chrome.runtime.sendNativeMessage as jest.Mock).mockImplementationOnce((_host: string, _msg: unknown, callback: (response: unknown) => void) => {
        callback(null);
      });

      const result = await bridge.getStatus();
      
      expect(result.success).toBe(false);
      expect(result.error).toBe('Native error message');
      
      // Cleanup
      (chrome.runtime as any).lastError = originalLastError;
    });

    it('should handle chrome.runtime.lastError without message property', async () => {
      const bridge = new NativeBridge();
      
      // Setup mock to simulate lastError without message
      const originalLastError = chrome.runtime.lastError;
      (chrome.runtime as any).lastError = {}; 
      
      (chrome.runtime.sendNativeMessage as jest.Mock).mockImplementationOnce((_host: string, _msg: unknown, callback: (response: unknown) => void) => {
        callback(null);
      });

      const result = await bridge.getStatus();
      
      expect(result.success).toBe(false);
      expect(result.error).toBe('Unknown native messaging error');
      
      // Cleanup
      (chrome.runtime as any).lastError = originalLastError;
    });

    it('should handle missing response from native host', async () => {
      const bridge = new NativeBridge();
      
      (chrome.runtime.sendNativeMessage as jest.Mock).mockImplementationOnce((_host: string, _msg: unknown, callback: (response: unknown) => void) => {
        callback(null);
      });

      const result = await bridge.getStatus();
      
      expect(result.success).toBe(false);
      expect(result.error).toBe('No response from native host');
    });

    it('should handle exceptions in isAvailable', async () => {
      const bridge = new NativeBridge();
      jest.spyOn(bridge, 'getStatus').mockRejectedValueOnce(new Error('Fatal error'));
      
      const result = await bridge.isAvailable();
      
      expect(result).toBe(false);
    });
  });
});

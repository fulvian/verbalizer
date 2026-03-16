/**
 * Tests for Background Service Worker.
 */

import { BackgroundService } from './index';
import { NativeBridge } from './native-bridge';
import { CallStartedMessage, CallEndedMessage, CallDetectedMessage } from '../types/messages';

// Mock the NativeBridge class
jest.mock('./native-bridge');

describe('BackgroundService', () => {
  let backgroundService: BackgroundService;
  let mockNativeBridge: jest.Mocked<NativeBridge>;
  let mockSendResponse: jest.Mock;

  beforeEach(() => {
    // Create mock NativeBridge instance
    mockNativeBridge = new NativeBridge() as jest.Mocked<NativeBridge>;
    mockSendResponse = jest.fn();
    
    // Clear mock state
    jest.clearAllMocks();
    
    // Instantiate BackgroundService with mock
    backgroundService = new BackgroundService(mockNativeBridge);
  });

  describe('handleMessage', () => {
    it('should handle CALL_STARTED and start recording', async () => {
      const message: CallStartedMessage = {
        type: 'CALL_STARTED',
        payload: {
          platform: 'google-meet',
          callId: 'test-call-id',
          title: 'Custom Title',
          participants: ['User 1']
        }
      };

      mockNativeBridge.startRecording.mockResolvedValue({ 
        success: true, 
        data: { recordingPath: '/test/path' } 
      });

      await backgroundService.handleMessage(message, {} as any, mockSendResponse);

      expect(mockNativeBridge.startRecording).toHaveBeenCalledWith({
        platform: 'google-meet',
        callId: 'test-call-id',
        title: 'Custom Title'
      });
      expect(mockSendResponse).toHaveBeenCalledWith({
        success: true,
        data: { recordingPath: '/test/path' }
      });
    });

    it('should handle CALL_ENDED and stop recording', async () => {
      const message: CallEndedMessage = {
        type: 'CALL_ENDED',
        payload: {
          platform: 'google-meet',
          callId: 'test-call-id',
          duration: 120
        }
      };

      mockNativeBridge.stopRecording.mockResolvedValue({ success: true });

      await backgroundService.handleMessage(message, {} as any, mockSendResponse);

      expect(mockNativeBridge.stopRecording).toHaveBeenCalledWith({
        callId: 'test-call-id'
      });
      expect(mockSendResponse).toHaveBeenCalledWith({ success: true });
    });

    it('should handle CALL_DETECTED and respond with success', async () => {
      const message: CallDetectedMessage = {
        type: 'CALL_DETECTED',
        payload: {
          platform: 'ms-teams',
          url: 'https://teams.microsoft.com/l/meetup-join/...',
          callId: 'test-call-id'
        }
      };

      await backgroundService.handleMessage(message, {} as any, mockSendResponse);

      expect(mockSendResponse).toHaveBeenCalledWith({ success: true });
    });

    it('should return error for unknown message type', async () => {
      const message = { type: 'UNKNOWN_TYPE', payload: {} } as any;

      await backgroundService.handleMessage(message, {} as any, mockSendResponse);

      expect(mockSendResponse).toHaveBeenCalledWith({
        success: false,
        error: 'Unknown message type'
      });
    });

    it('should handle errors gracefully', async () => {
      const message: CallStartedMessage = {
        type: 'CALL_STARTED',
        payload: {
          platform: 'google-meet',
          callId: 'test-call-id'
        }
      };

      mockNativeBridge.startRecording.mockRejectedValue(new Error('Native error'));

      await backgroundService.handleMessage(message, {} as any, mockSendResponse);

      expect(mockSendResponse).toHaveBeenCalledWith({
        success: false,
        error: 'Native error'
      });
    });
  });

  describe('setupEventListeners', () => {
    it('should register a message listener', () => {
      expect(chrome.runtime.onMessage.addListener).toHaveBeenCalled();
    });

    it('should call handleMessage when a message is received', () => {
      // Get the listener that was registered
      const listener = (chrome.runtime.onMessage.addListener as jest.Mock).mock.calls[0][0];
      const message: CallDetectedMessage = { 
        type: 'CALL_DETECTED', 
        payload: { 
          platform: 'google-meet', 
          url: 'https://meet.google.com/abc',
          callId: 'test-id' 
        } 
      };
      const sender = {};
      const sendResponse = jest.fn();
      
      const handleMessageSpy = jest.spyOn(backgroundService, 'handleMessage').mockImplementation();
      
      listener(message, sender, sendResponse);
      
      expect(handleMessageSpy).toHaveBeenCalledWith(message, sender, sendResponse);
    });
  });
});

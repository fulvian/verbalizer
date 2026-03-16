import { NativeBridge } from './native-bridge';
import { ExtensionMessage } from '../types/messages';

/**
 * Background Service Worker for the Verbalizer extension.
 * Handles messages from content scripts and communicates with the native host.
 */
export class BackgroundService {
  private readonly nativeBridge: NativeBridge;

  constructor(nativeBridge: NativeBridge = new NativeBridge()) {
    this.nativeBridge = nativeBridge;
    this.setupEventListeners();
  }

  /**
   * Sets up Chrome event listeners.
   */
  private setupEventListeners(): void {
    if (typeof chrome !== 'undefined' && chrome.runtime && chrome.runtime.onMessage) {
      chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
        this.handleMessage(message as ExtensionMessage, sender, sendResponse);
        // Return true to indicate we will send a response asynchronously
        return true;
      });
    }
  }

  /**
   * Handles incoming messages from content scripts.
   * 
   * @param message The message received from the content script
   * @param _sender The sender of the message
   * @param sendResponse Callback function to send a response back
   */
  public async handleMessage(
    message: ExtensionMessage,
    _sender: chrome.runtime.MessageSender,
    sendResponse: (response?: any) => void
  ): Promise<void> {
    try {
      switch (message.type) {
        case 'CALL_STARTED': {
          const response = await this.nativeBridge.startRecording({
            platform: message.payload.platform,
            callId: message.payload.callId,
            title: message.payload.title || `Call on ${message.payload.platform}`
          });
          sendResponse(response);
          break;
        }

        case 'CALL_ENDED': {
          const response = await this.nativeBridge.stopRecording({
            callId: message.payload.callId
          });
          sendResponse(response);
          break;
        }

        case 'CALL_DETECTED':
        case 'PARTICIPANTS_CHANGED':
          // These are handled silently or used for UI updates in future
          sendResponse({ success: true });
          break;

        default:
          console.warn(`[Verbalizer] Unknown message type: ${(message as any).type}`);
          sendResponse({ success: false, error: 'Unknown message type' });
      }
    } catch (error) {
      console.error('[Verbalizer] Error in background message handler:', error);
      sendResponse({
        success: false,
        error: error instanceof Error ? error.message : 'Internal background error'
      });
    }
  }
}

// Initialize the service if not in a test environment
/* istanbul ignore next */
if (typeof window === 'undefined' && !(globalThis as any).__TEST__) {
  new BackgroundService();
}

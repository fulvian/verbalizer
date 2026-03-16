/**
 * Native Messaging bridge.
 * Communicates with the native host binary via Chrome Native Messaging.
 */

type Platform = 'google-meet' | 'ms-teams';

interface NativeMessage {
  type: string;
  payload: unknown;
}

interface NativeResponse {
  success: boolean;
  data?: unknown;
  error?: string;
}

// Native Messaging host name (must match manifest config)
const NATIVE_HOST_NAME = 'com.verbalizer.host';

/**
 * Native Messaging bridge for communicating with the native host binary.
 */
export class NativeBridge {
  /**
   * Send a message to the native host and wait for response.
   */
  private async send(message: NativeMessage): Promise<NativeResponse> {
    return new Promise((resolve) => {
      chrome.runtime.sendNativeMessage(
        NATIVE_HOST_NAME,
        message,
        (response: NativeResponse) => {
          if (chrome.runtime.lastError) {
            resolve({
              success: false,
              error: chrome.runtime.lastError.message,
            });
          } else {
            resolve(response);
          }
        }
      );
    });
  }
  
  /**
   * Start recording a call.
   */
  async startRecording(payload: {
    platform: Platform;
    callId: string;
    title?: string;
  }): Promise<NativeResponse> {
    return this.send({
      type: 'START_RECORDING',
      payload,
    });
  }
  
  /**
   * Stop recording a call.
   */
  async stopRecording(payload: { callId: string }): Promise<NativeResponse> {
    return this.send({
      type: 'STOP_RECORDING',
      payload,
    });
  }
  
  /**
   * Get current status from native host.
   */
  async getStatus(): Promise<NativeResponse> {
    return this.send({
      type: 'GET_STATUS',
      payload: {},
    });
  }
}

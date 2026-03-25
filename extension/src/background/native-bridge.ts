/**
 * Native Messaging Bridge.
 * Communicates with the native host via Chrome Native Messaging.
 * 
 * REASONING:
 * - The NativeBridge class provides a clean API for native messaging
 * - Handles connection errors gracefully
 * - Provides typed interfaces for all operations
 * - The native host name must match the manifest.json
 */

// Types are defined locally in this file

const NATIVE_HOST_NAME = 'com.verbalizer.host';

/**
 * Payload types for native host requests.
 */
interface StartRecordingPayload {
  platform: string;
  callId: string;
  title?: string;
}

interface StopRecordingPayload {
  callId: string;
}

/**
 * Response type from native host.
 */
interface NativeResponse {
  success: boolean;
  data?: {
    recordingPath?: string;
    isRecording?: boolean;
  };
  error?: string;
}

/**
 * Google Auth status response.
 */
interface GoogleAuthStatusResponse {
  success: boolean;
  data?: {
    enabled: boolean;
    connected: boolean;
    email?: string;
    name?: string;
    provider?: string;
    scopes?: string;
  };
  error?: string;
}

/**
 * Google Auth start response.
 */
interface GoogleAuthStartResponse {
  success: boolean;
  error?: string;
}

/**
 * Google Drive folder response.
 */
interface GoogleDriveFolderResponse {
  success: boolean;
  data?: {
    folders: Array<{
      id: string;
      name: string;
      parentId?: string;
    }>;
  };
  error?: string;
}

/**
 * NativeBridge provides a clean API for communicating with the native host.
 */
export class NativeBridge {
  private readonly nativeHostName: string = NATIVE_HOST_NAME;
  
  /**
   * Send a message to the native host and wait for response.
   * 
   * REASONING:
   * - Uses Chrome's native messaging API
   * - Wraps the callback-based API in a Promise
   * - Handles errors gracefully with typed responses
   */
  private async sendNativeMessage<T>(message: unknown): Promise<T> {
    return new Promise((resolve) => {
      chrome.runtime.sendNativeMessage(
        this.nativeHostName,
        message as object,
        (response) => {
          if (chrome.runtime.lastError) {
            // Handle runtime errors
            const error = chrome.runtime.lastError;
            resolve({
              success: false,
              error: error?.message || 'Unknown native messaging error',
            } as T);
            return;
          }
          
          if (!response) {
            resolve({
              success: false,
              error: 'No response from native host',
            } as T);
            return;
          }
          
          resolve(response as T);
        }
      );
    });
  }
  
  /**
   * Start recording a call.
   * 
   * REASONING:
   * - Sends START_RECORDING request to native host
   * - Includes call metadata for file naming
   */
  async startRecording(payload: StartRecordingPayload): Promise<NativeResponse> {
    return this.sendNativeMessage<NativeResponse>({
      type: 'START_RECORDING',
      payload,
    });
  }
  
  /**
   * Stop recording a call.
   * 
   * REASONING:
   * - Sends STOP_RECORDING request to native host
   * - Native host will finalize the recording file
   */
  async stopRecording(payload: StopRecordingPayload): Promise<NativeResponse> {
    return this.sendNativeMessage<NativeResponse>({
      type: 'STOP_RECORDING',
      payload,
    });
  }
  
  /**
   * Get current status from the native host.
   * 
   * REASONING:
   * - Used to check if native host is running
   * - Returns current recording state
   */
  async getStatus(): Promise<NativeResponse> {
    return this.sendNativeMessage<NativeResponse>({
      type: 'GET_STATUS',
      payload: {},
    });
  }
  
  /**
   * Check if the native host is available.
   * 
   * REASONING:
   * - Quick check to verify native host connectivity
   * - Used before attempting recording operations
   */
  async isAvailable(): Promise<boolean> {
    try {
      const response = await this.getStatus();
      return response.success;
    } catch {
      return false;
    }
  }

  /**
   * Start Google OAuth authentication flow.
   */
  async googleAuthStart(): Promise<GoogleAuthStartResponse> {
    return this.sendNativeMessage<GoogleAuthStartResponse>({
      type: 'GOOGLE_AUTH_START',
      payload: {},
    });
  }

  /**
   * Get Google OAuth authentication status.
   */
  async googleAuthStatus(): Promise<GoogleAuthStatusResponse> {
    return this.sendNativeMessage<GoogleAuthStatusResponse>({
      type: 'GOOGLE_AUTH_STATUS',
      payload: {},
    });
  }

  /**
   * Disconnect Google account.
   */
  async googleAuthDisconnect(): Promise<NativeResponse> {
    return this.sendNativeMessage<NativeResponse>({
      type: 'GOOGLE_AUTH_DISCONNECT',
      payload: {},
    });
  }

  /**
   * Get list of Google Drive folders.
   */
  async googleDriveGetFolders(): Promise<GoogleDriveFolderResponse> {
    return this.sendNativeMessage<GoogleDriveFolderResponse>({
      type: 'GOOGLE_DRIVE_GET_FOLDER',
      payload: {},
    });
  }

  /**
   * Set target Google Drive folder.
   */
  async googleDriveSetFolder(folderId: string): Promise<NativeResponse> {
    return this.sendNativeMessage<NativeResponse>({
      type: 'GOOGLE_DRIVE_SET_FOLDER',
      payload: { folderId },
    });
  }
}

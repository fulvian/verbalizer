/**
 * Call state observer.
 * Manages DOM observation and notifies background script of state changes.
 * 
 * REASONING:
 * - The CallStateObserver class manages all call state detection
 * - It provides methods for detecting, starting, and ending calls
 * - Handles cleanup through registered callbacks
 * - Generates unique call IDs for tracking
 */

type Platform = 'google-meet' | 'ms-teams';

type CleanupCallback = () => void;

export class CallStateObserver {
  // Platform is stored for potential future use (e.g., platform-specific behavior)
  private readonly _platform: Platform;
  private readonly cleanupCallbacks: CleanupCallback[] = [];
  private callStartTime: number | null = null;
  private currentCallId: string | null = null;
  private mutationObserver: MutationObserver | null = null;
  
  constructor(platform: Platform) {
    this._platform = platform;
  }
  
  /**
   * Notify background that a call was detected on the platform.
   */
  notifyCallDetected(): void {
    this.currentCallId = this.generateCallId(this._platform);
    
    chrome.runtime.sendMessage({
      type: 'CALL_DETECTED',
      payload: {
        platform: this._platform,
        url: window.location.href,
        title: document.title,
        callId: this.currentCallId,
      },
    }).catch((error) => {
      console.error('[Verbalizer] Failed to send CALL_DETECTED:', error);
    });
  }
  
  /**
   * Notify background that user joined a call.
   */
  notifyCallStarted(title?: string): void {
    this.callStartTime = Date.now();
    this.currentCallId = this.generateCallId(this._platform);
    
    chrome.runtime.sendMessage({
      type: 'CALL_STARTED',
      payload: {
        platform: this._platform,
        callId: this.currentCallId,
        title,
      },
    }).catch((error) => {
      console.error('[Verbalizer] Failed to send CALL_STARTED:', error);
    });
  }
  
  /**
   * Notify background that user left the call.
   */
  notifyCallEnded(): void {
    const duration = this.callStartTime 
      ? Math.floor((Date.now() - this.callStartTime) / 1000)
      : 0;
    
    chrome.runtime.sendMessage({
      type: 'CALL_ENDED',
      payload: {
        platform: this._platform,
        callId: this.currentCallId || '',
        duration,
      },
    }).catch((error) => {
      console.error('[Verbalizer] Failed to send CALL_ENDED:', error);
    });
    
    this.callStartTime = null;
    this.currentCallId = null;
  }
  
  /**
   * Register a cleanup callback.
   */
  registerCleanup(callback: CleanupCallback): void {
    this.cleanupCallbacks.push(callback);
  }
  
  /**
   * Disconnect the observer and run all cleanup callbacks.
   */
  disconnect(): void {
    for (const callback of this.cleanupCallbacks) {
      try {
        callback();
      } catch (error) {
        console.error('[Verbalizer] Cleanup callback failed:', error);
      }
    }
    this.cleanupCallbacks.length = 0;
    
    if (this.mutationObserver) {
      this.mutationObserver.disconnect();
      this.mutationObserver = null;
    }
  }
  
  /**
   * Generate a unique call ID for tracking.
   */
  private generateCallId(platform: Platform): string {
    const timestamp = Date.now();
    const randomPart = Math.random().toString(36).substring(2, 6);
    return `${platform}-${timestamp}-${randomPart}`;
  }
  
  /**
   * Get the current call ID (if any).
   */
  getCurrentCallId(): string | null {
    return this.currentCallId;
  }
  
  /**
   * Check if currently in a call.
   */
  isInCall(): boolean {
    return this.callStartTime !== null;
  }
  
  /**
   * Get the current call duration in seconds.
   */
  getCallDuration(): number | null {
    if (!this.callStartTime) {
      return null;
    }
    return Math.floor((Date.now() - this.callStartTime) / 1000);
  }
}

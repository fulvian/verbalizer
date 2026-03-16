/**
 * Call state observer.
 * Manages DOM observation and notifies background script of state changes.
 */

type Platform = 'google-meet' | 'ms-teams';

type CleanupCallback = () => void;

export class CallStateObserver {
  private readonly platform: Platform;
  private readonly cleanupCallbacks: CleanupCallback[] = [];
  private callStartTime: number | null = null;
  private mutationObserver: MutationObserver | null = null;
  
  constructor(platform: Platform) {
    this.platform = platform;
  }
  
  /**
   * Notify background that a call was detected on the platform.
   */
  notifyCallDetected(platform: Platform): void {
    chrome.runtime.sendMessage({
      type: 'CALL_DETECTED',
      payload: {
        platform,
        url: window.location.href,
        title: document.title,
      },
    }).catch((error) => {
      console.error('[Verbalizer] Failed to send CALL_DETECTED:', error);
    });
  }
  
  /**
   * Notify background that user joined a call.
   */
  notifyCallStarted(platform: Platform, title?: string): void {
    this.callStartTime = Date.now();
    
    chrome.runtime.sendMessage({
      type: 'CALL_STARTED',
      payload: {
        platform,
        callId: this.generateCallId(),
        title,
      },
    }).catch((error) => {
      console.error('[Verbalizer] Failed to send CALL_STARTED:', error);
    });
  }
  
  /**
   * Notify background that user left the call.
   */
  notifyCallEnded(platform: Platform): void {
    const duration = this.callStartTime 
      ? Math.floor((Date.now() - this.callStartTime) / 1000)
      : 0;
    
    chrome.runtime.sendMessage({
      type: 'CALL_ENDED',
      payload: {
        platform,
        callId: this.generateCallId(),
        duration,
      },
    }).catch((error) => {
      console.error('[Verbalizer] Failed to send CALL_ENDED:', error);
    });
    
    this.callStartTime = null;
  }
  
  /**
   * Register a cleanup callback to be called on disconnect.
   */
  registerCleanup(callback: CleanupCallback): void {
    this.cleanupCallbacks.push(callback);
  }
  
  /**
   * Disconnect observer and run all cleanup callbacks.
   */
  disconnect(): void {
    if (this.mutationObserver) {
      this.mutationObserver.disconnect();
      this.mutationObserver = null;
    }
    
    for (const callback of this.cleanupCallbacks) {
      try {
        callback();
      } catch (error) {
        console.error('[Verbalizer] Cleanup callback error:', error);
      }
    }
    
    this.cleanupCallbacks.length = 0;
  }
  
  /**
   * Generate a unique call ID based on timestamp.
   */
  private generateCallId(): string {
    const now = new Date();
    const timestamp = now.toISOString()
      .replace(/[-:]/g, '')
      .replace(/\.\d+/, '')
      .replace('T', '_');
    
    return `${timestamp}_${this.platform}`;
  }
}

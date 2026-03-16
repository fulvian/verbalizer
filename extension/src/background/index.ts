/**
 * Background service worker entry point.
 * Handles communication between content scripts and native host.
 */

import { NativeBridge } from './native-bridge';

// Initialize native bridge
const nativeBridge = new NativeBridge();

// Track active calls
const activeCalls = new Map<string, {
  platform: 'google-meet' | 'ms-teams';
  startTime: number;
  title?: string;
}>();

// Listen for messages from content scripts
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  handleMessage(message)
    .then(sendResponse)
    .catch((error) => {
      console.error('[Verbalizer] Message handler error:', error);
      sendResponse({ success: false, error: error.message });
    });
  
  // Return true to indicate async response
  return true;
});

/**
 * Handle incoming message from content script.
 */
export async function handleMessage(message: { type: string; payload: unknown }): Promise<{ success: boolean; data?: unknown; error?: string }> {
  console.log('[Verbalizer] Received message:', message.type);
  
  switch (message.type) {
    case 'CALL_DETECTED':
      return handleCallDetected(message.payload as { platform: 'google-meet' | 'ms-teams'; url: string; title?: string });
    
    case 'CALL_STARTED':
      return handleCallStarted(message.payload as { platform: 'google-meet' | 'ms-teams'; callId: string; title?: string });
    
    case 'CALL_ENDED':
      return handleCallEnded(message.payload as { platform: 'google-meet' | 'ms-teams'; callId: string; duration: number });
    
    default:
      return { success: false, error: `Unknown message type: ${message.type}` };
  }
}

/**
 * Handle call detected event.
 */
async function handleCallDetected(payload: { platform: 'google-meet' | 'ms-teams'; url: string; title?: string }): Promise<{ success: boolean }> {
  console.log(`[Verbalizer] Call detected on ${payload.platform}: ${payload.url}`);
  
  // Check if native host is available
  const status = await nativeBridge.getStatus();
  if (!status.success) {
    console.warn('[Verbalizer] Native host not available:', status.error);
  }
  
  return { success: true };
}

/**
 * Handle call started event.
 */
async function handleCallStarted(payload: { platform: 'google-meet' | 'ms-teams'; callId: string; title?: string }): Promise<{ success: boolean; data?: { recordingPath?: string }; error?: string }> {
  console.log(`[Verbalizer] Call started: ${payload.callId}`);
  
  // Track the call
  activeCalls.set(payload.callId, {
    platform: payload.platform,
    startTime: Date.now(),
    title: payload.title,
  });
  
  // Start recording via native host
  const response = await nativeBridge.startRecording({
    platform: payload.platform,
    callId: payload.callId,
    title: payload.title,
  });
  
  if (response.success) {
    console.log('[Verbalizer] Recording started:', response.data);
    return { success: true, data: response.data as { recordingPath?: string } };
  } else {
    console.error('[Verbalizer] Failed to start recording:', response.error);
    return { success: false, error: response.error };
  }
}

/**
 * Handle call ended event.
 */
async function handleCallEnded(payload: { platform: 'google-meet' | 'ms-teams'; callId: string; duration: number }): Promise<{ success: boolean; error?: string }> {
  console.log(`[Verbalizer] Call ended: ${payload.callId} (${payload.duration}s)`);
  
  // Remove from active calls
  activeCalls.delete(payload.callId);
  
  // Stop recording via native host
  const response = await nativeBridge.stopRecording({ callId: payload.callId });
  
  if (response.success) {
    console.log('[Verbalizer] Recording stopped');
    return { success: true };
  } else {
    console.error('[Verbalizer] Failed to stop recording:', response.error);
    return { success: false, error: response.error };
  }
}

// Log when service worker starts
console.log('[Verbalizer] Background service worker started');

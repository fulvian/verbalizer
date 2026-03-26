/**
 * Content script entry point.
 * Injected into Google Meet and Microsoft Teams pages.
 * Monitors DOM for call state changes and communicates with background script.
 */

import { detectGoogleMeet } from './detectors/meet';
import { detectMSTeams } from './detectors/teams';
import { CallStateObserver } from './observer';
import { contentLogger } from '../utils/logger';

// Platform detection
type Platform = 'google-meet' | 'ms-teams' | null;

export let currentPlatform: Platform = null;
export let observer: CallStateObserver | null = null;

// Current call session ID for correlation
export let currentCallId: string | null = null;

/**
 * Detect which platform we're on based on URL.
 */
export function detectPlatform(): Platform {
  const url = window.location.href;
  
  if (url.includes('meet.google.com')) {
    return 'google-meet';
  }
  
  if (url.includes('teams.microsoft.com') || url.includes('teams.live.com')) {
    return 'ms-teams';
  }
  
  return null;
}

/**
 * Initialize the content script for the detected platform.
 */
export function initialize(): void {
  try {
    currentPlatform = detectPlatform();
    contentLogger.info('Platform detected', { reason: currentPlatform || 'unknown' });
    contentLogger.debug('Current URL', { reason: window.location.href });
    
    if (!currentPlatform) {
      contentLogger.warn('Not a supported platform');
      return;
    }
    
    // Create observer for call state changes
    observer = new CallStateObserver(currentPlatform);
    contentLogger.info('Observer created', { reason: currentPlatform });
    
    // Set up platform-specific detection
    if (currentPlatform === 'google-meet') {
      detectGoogleMeet(observer);
    } else if (currentPlatform === 'ms-teams') {
      detectMSTeams(observer);
      contentLogger.info('Teams detector initialized');
    }
  } catch (error) {
    const errorMsg = error instanceof Error ? error.message : String(error);
    contentLogger.error('Initialization failed', { errorCode: 'INIT_ERROR', metadata: { error: errorMsg } });
  }
}

/**
 * Cleanup function for page unload.
 */
export function cleanup(): void {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
}

/**
 * Main entry point setup.
 */
export function setup(): void {
  console.log('[Verbalizer] Content script starting...');
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initialize);
  } else {
    initialize();
  }
  
  window.addEventListener('beforeunload', cleanup);
}

// Auto-initialize if not in a test environment
/* istanbul ignore next */
if (typeof window !== 'undefined' && !(window as any).__TEST__) {
  setup();
}

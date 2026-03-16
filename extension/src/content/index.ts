/**
 * Content script entry point.
 * Injected into Google Meet and Microsoft Teams pages.
 * Monitors DOM for call state changes and communicates with background script.
 */

import { detectGoogleMeet } from './detectors/meet';
import { detectMSTeams } from './detectors/teams';
import { CallStateObserver } from './observer';

// Platform detection
type Platform = 'google-meet' | 'ms-teams' | null;

export let currentPlatform: Platform = null;
export let observer: CallStateObserver | null = null;

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
    
    if (!currentPlatform) {
      console.warn('[Verbalizer] Not a supported platform');
      return;
    }
    
    // Create observer for call state changes
    observer = new CallStateObserver(currentPlatform);
    
    // Set up platform-specific detection
    if (currentPlatform === 'google-meet') {
      detectGoogleMeet(observer);
    } else if (currentPlatform === 'ms-teams') {
      detectMSTeams(observer);
    }
  } catch (error) {
    console.error('[Verbalizer] Initialization failed:', error);
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

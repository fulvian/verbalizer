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

let currentPlatform: Platform = null;
let observer: CallStateObserver | null = null;

/**
 * Detect which platform we're on based on URL.
 */
function detectPlatform(): Platform {
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
function initialize(): void {
  currentPlatform = detectPlatform();
  
  if (!currentPlatform) {
    console.warn('[Verbalizer] Not a supported platform');
    return;
  }
  
  console.log(`[Verbalizer] Initialized on ${currentPlatform}`);
  
  // Create observer for call state changes
  observer = new CallStateObserver(currentPlatform);
  
  // Set up platform-specific detection
  if (currentPlatform === 'google-meet') {
    detectGoogleMeet(observer);
  } else if (currentPlatform === 'ms-teams') {
    detectMSTeams(observer);
  }
}

// Run initialization when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initialize);
} else {
  initialize();
}

// Clean up on page unload
window.addEventListener('beforeunload', () => {
  if (observer) {
    observer.disconnect();
  }
});

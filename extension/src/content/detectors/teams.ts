/**
 * Microsoft Teams call detection.
 * Monitors DOM for call state changes in MS Teams.
 */

import { CallStateObserver } from '../observer';

// MS Teams DOM selectors (may need updates if Microsoft changes UI)
const SELECTORS = {
  // Call container
  callContainer: '[data-tid="call-container"], .ts-calling-thread',
  
  // Active call indicators
  activeCall: '[data-tid="call-state"], .calling-live-indicator',
  
  // Participants
  participants: '[data-tid="participant-item"]',
  
  // Meeting title
  meetingTitle: '[data-tid="meeting-title"], .ts-call-title',
  
  // Call controls
  callControls: '[data-tid="call-controls"]',
};

/**
 * Set up MS Teams detection.
 */
export function detectMSTeams(observer: CallStateObserver): void {
  console.log('[Verbalizer] Setting up MS Teams detection');
  
  // Watch for call container to appear
  const checkForCall = setInterval(() => {
    const callContainer = document.querySelector(SELECTORS.callContainer);
    const activeCall = document.querySelector(SELECTORS.activeCall);
    
    if (callContainer || activeCall) {
      clearInterval(checkForCall);
      onCallDetected(observer);
    }
  }, 1000);
  
  // Clean up interval after 30 seconds if no call found
  setTimeout(() => clearInterval(checkForCall), 30000);
}

/**
 * Check if an MS Teams call is currently active.
 */
export function isMSTeamsActive(): boolean {
  const activeCall = document.querySelector(SELECTORS.activeCall);
  const callContainer = document.querySelector(SELECTORS.callContainer);
  return activeCall !== null || callContainer !== null;
}

/**
 * Handle call detection.
 */
function onCallDetected(observer: CallStateObserver): void {
  console.log('[Verbalizer] MS Teams detected');
  
  // Notify background script
  observer.notifyCallDetected('ms-teams');
  
  // Extract meeting title
  const titleElement = document.querySelector(SELECTORS.meetingTitle);
  const title = titleElement?.textContent || undefined;
  
  // Watch for user joining/leaving
  watchCallState(observer, title);
}

/**
 * Watch for call state changes.
 */
function watchCallState(observer: CallStateObserver, title?: string): void {
  let isInCall = false;
  
  const checkCallState = () => {
    const active = isMSTeamsActive();
    
    if (active && !isInCall) {
      // User joined call
      isInCall = true;
      observer.notifyCallStarted('ms-teams', title);
    } else if (!active && isInCall) {
      // User left call
      isInCall = false;
      observer.notifyCallEnded('ms-teams');
    }
  };
  
  // Check every second
  const interval = setInterval(checkCallState, 1000);
  
  // Store interval for cleanup
  observer.registerCleanup(() => clearInterval(interval));
}

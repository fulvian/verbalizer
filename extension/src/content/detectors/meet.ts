/**
 * Google Meet call detection.
 * Monitors DOM for call state changes in Google Meet.
 */

import { CallStateObserver } from '../observer';

// Google Meet DOM selectors (may need updates if Google changes UI)
const SELECTORS = {
  // Meeting room container
  meetingRoom: '[data-meeting-title], [jsname="wzVsNb"]',
  
  // Active call indicators
  activeCall: '[data-is-muted], [data-is-camera-active]',
  
  // Participants panel
  participantsPanel: '[data-participant-id]',
  
  // Call controls
  callControls: '[data-call-controls]',
  
  // Meeting title
  meetingTitle: '[data-meeting-title]',
};

/**
 * Set up Google Meet detection.
 */
export function detectGoogleMeet(observer: CallStateObserver): void {
  console.log('[Verbalizer] Setting up Google Meet detection');
  
  // Watch for meeting room to appear
  const checkForMeeting = setInterval(() => {
    const meetingRoom = document.querySelector(SELECTORS.meetingRoom);
    const activeCall = document.querySelector(SELECTORS.activeCall);
    
    if (meetingRoom && activeCall) {
      clearInterval(checkForMeeting);
      onMeetingDetected(observer);
    }
  }, 1000);
  
  // Clean up interval after 30 seconds if no meeting found
  setTimeout(() => clearInterval(checkForMeeting), 30000);
}

/**
 * Check if a Google Meet call is currently active.
 */
export function isGoogleMeetActive(): boolean {
  const activeCall = document.querySelector(SELECTORS.activeCall);
  return activeCall !== null;
}

/**
 * Handle meeting detection.
 */
function onMeetingDetected(observer: CallStateObserver): void {
  console.log('[Verbalizer] Google Meet detected');
  
  // Notify background script
  observer.notifyCallDetected('google-meet');
  
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
    const active = isGoogleMeetActive();
    
    if (active && !isInCall) {
      // User joined call
      isInCall = true;
      observer.notifyCallStarted('google-meet', title);
    } else if (!active && isInCall) {
      // User left call
      isInCall = false;
      observer.notifyCallEnded('google-meet');
    }
  };
  
  // Check every second
  const interval = setInterval(checkCallState, 1000);
  
  // Store interval for cleanup
  observer.registerCleanup(() => clearInterval(interval));
}

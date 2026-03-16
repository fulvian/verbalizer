/**
 * Google Meet call detection.
 * Monitors DOM for call state changes in Google Meet.
 * 
 * REASONING:
 * Google Meet uses specific DOM structures to indicate call state:
 * - Active call has specific data attributes on elements
 * - We use MutationObserver to detect dynamic changes
 * - Selectors may change when Google updates Meet UI
 */

import { CallStateObserver } from '../observer';

// Google Meet DOM selectors
// These selectors are based on Google Meet's current DOM structure
// They may need updates when Google changes the Meet UI
const MEET_SELECTORS = {
  // Main meeting container - present when in a meeting
  meetingContainer: '[data-meeting-readiness-state]',
  
  // Active call indicators
  // These elements appear when user is actively in a call
  activeCall: '[data-call-ended="false"]',
  callEnded: '[data-call-ended="true"]',
  
  // Video/audio elements
  videoElement: 'video',
  audioElement: 'audio',
  
  // Participants panel
  participantsPanel: '[data-participant-id]',
  
  // Meeting title
  meetingTitle: '[data-meeting-title]',
  
  // Controls area
  controlsArea: '[data-control-bar]',
} as const;

/**
 * Check if Google Meet is currently active (user is in a call).
 * 
 * REASONING:
 * Multiple indicators provide redundancy for reliability
 * - data-meeting-readiness-state indicates meeting container
 * - video/audio elements indicate media streaming
 */
export function isGoogleMeetActive(): boolean {
  // Check for meeting container with readiness state
  const meetingContainer = document.querySelector(MEET_SELECTORS.meetingContainer);
  if (meetingContainer) {
    return true;
  }

  // Check for video elements (active media streaming)
  const videoElements = document.querySelectorAll(MEET_SELECTORS.videoElement);
  if (videoElements.length > 0) {
    return true;
  }

  // Check for call-ended indicator
  const callEnded = document.querySelector(MEET_SELECTORS.callEnded);
  if (callEnded) {
    return false; // Call has ended
  }

  return false;
}

/**
 * Extract meeting title from Google Meet page.
 * 
 * REASONING:
 * The meeting title can be extracted from:
 * - data-meeting-title attribute
 * - Document title as fallback
 */
export function extractMeetingTitle(): string | undefined {
  // Try to get title from data attribute
  const titleElement = document.querySelector(MEET_SELECTORS.meetingTitle);
  if (titleElement) {
    return titleElement.getAttribute('data-meeting-title') || undefined;
  }

  // Fallback to document title
  return document.title || undefined;
}

/**
 * Extract participant count from Google Meet.
 * 
 * REASONING:
 * Participant count helps with meeting metadata
 * - Count participant elements in the DOM
 */
export function extractParticipantCount(): number {
  const participants = document.querySelectorAll(MEET_SELECTORS.participantsPanel);
  return participants.length;
}

/**
 * Set up Google Meet detection.
 * 
 * REASONING:
 * - Uses polling with intervals for reliability
 * - MutationObserver for immediate DOM changes
 * - Cleanup handlers prevent memory leaks
 */
export function detectGoogleMeet(observer: CallStateObserver): void {
  // Track if we were already in a call
  let wasInCall = false;
  let callCheckInterval: ReturnType<typeof setInterval> | null = null;

  /**
   * Check for call state changes.
   * 
   * REASONING:
   * Polling provides backup to MutationObserver
   * Some DOM changes may not trigger mutations
   */
  function checkForCall(): void {
    const isInCall = isGoogleMeetActive();

    if (isInCall && !wasInCall) {
      // Call started
      wasInCall = true;
      observer.notifyCallDetected();
      observer.notifyCallStarted(extractMeetingTitle());
    } else if (!isInCall && wasInCall) {
      // Call ended
      wasInCall = false;
      observer.notifyCallEnded();
    }
  }

  // Initial check
  checkForCall();

  // Set up polling interval (backup for MutationObserver)
  // Google Meet sometimes doesn't trigger mutations for all state changes
  callCheckInterval = setInterval(checkForCall, 1000);

  // Set up MutationObserver for immediate detection
  const mutationObserver = new MutationObserver((mutations) => {
    // Check for relevant changes
    for (const mutation of mutations) {
      if (
        mutation.type === 'childList' ||
        mutation.type === 'attributes' ||
        mutation.type === 'characterData'
      ) {
        checkForCall();
        break; // Only need to check once per batch
      }
    }
  });

  // Observe the entire document for changes
  mutationObserver.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ['data-meeting-readiness-state', 'data-call-ended'],
  });

  // Register cleanup
  observer.registerCleanup(() => {
    if (callCheckInterval) {
      clearInterval(callCheckInterval);
    }
    mutationObserver.disconnect();
  });
}

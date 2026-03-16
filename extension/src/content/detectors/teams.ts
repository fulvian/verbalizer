/**
 * Microsoft Teams call detection.
 * Monitors DOM for call state changes in MS Teams.
 * 
 * REASONING:
 * MS Teams uses specific DOM structures to indicate call state:
 * - Teams uses different selectors than Meet
 * - We use MutationObserver to detect dynamic changes
 * - Selectors may change when Microsoft updates Teams UI
 */

import { CallStateObserver } from '../observer';

// MS Teams DOM selectors
// These selectors are based on MS Teams' current DOM structure
// They may need updates when Microsoft changes Teams UI
const TEAMS_SELECTORS = {
  // Main call container
  callContainer: '[data-tid="call-container"], .ts-calling-thread',
  
  // Active call indicator - present when in an active call
  activeCall: '[data-tid="call-state"], .calling-live-indicator',
  
  // Pre-join screen elements
  preJoinScreen: '[data-tid="prejoin-screen"]',
  
  // Call controls (hang up, leave, etc.)
  callControls: '[data-tid="call-controls"]',
  
  // End call button
  endCallButton: '[data-tid="hangup-button"], .ts-calling-button',

  // Meeting title
  meetingTitle: '[data-tid="meeting-title"], .ts-meeting-title',

  // Participants panel
  participantsPanel: '[data-tid="participant-item"], .ts-participant',
};

/**
 * Check if user is in an active MS Teams call.
 * 
 * REASONING:
 * MS Teams shows different UI states for:
 * - Pre-join screen: visible before joining
 * - Active call container visible during call
 * - Call controls visible during call
 * - End call button visible when call is ending
 */
export function isMSTeamsActive(): boolean {
  // Check for pre-join screen (user hasn't joined yet)
  const preJoinScreen = document.querySelector(TEAMS_SELECTORS.preJoinScreen);
  if (preJoinScreen) {
    return false;
  }

  // Check for active call container
  const callContainer = document.querySelector(TEAMS_SELECTORS.callContainer);
  if (!callContainer) {
    return false;
  }

  // Check for call-ended indicator
  const callEnded = document.querySelector(TEAMS_SELECTORS.endCallButton);
  if (callEnded) {
    return false; // Call has ended
  }

  return true;
}

/**
 * Extract meeting title from MS Teams page.
 * 
 * REASONING:
 * Meeting title is usually in the document title or
 * - May be in a specific element with data-tid
 */
export function extractMeetingTitle(): string | undefined {
  // Try to get title from specific element
  const titleElement = document.querySelector(TEAMS_SELECTORS.meetingTitle);
  const text = titleElement?.textContent?.trim();
  if (text) {
    return text;
  }

  // Fallback to document title
  return document.title || undefined;
}

/**
 * Extract participant count from MS Teams.
 * 
 * REASONING:
 * Participant count helps with meeting metadata
 * - Count participant avatars in the roster panel
 */
export function extractParticipantCount(): number {
  const participants = document.querySelectorAll(TEAMS_SELECTORS.participantsPanel);
  return participants.length;
}

/**
 * Set up MS Teams detection.
 * 
 * REASONING:
 * - Uses polling with intervals for reliability
 * - MutationObserver for immediate detection
 * - Cleanup on page unload
 */
export function detectMSTeams(observer: CallStateObserver): void {
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
    const isInCall = isMSTeamsActive();

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
        break;
      }
    }
  });

  // Observe the entire document for changes
  mutationObserver.observe(document.body, {
    childList: true,
    subtree: true,
    attributes: true,
    attributeFilter: ['data-tid', 'data-call-state'],
  });

  // Register cleanup
  observer.registerCleanup(() => {
    if (callCheckInterval) {
      clearInterval(callCheckInterval);
    }
    mutationObserver.disconnect();
  });
}

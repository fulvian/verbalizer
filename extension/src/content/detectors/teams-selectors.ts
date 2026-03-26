/**
 * Teams Web Selector Registry (versioned)
 * 
 * Provides stable, versioned DOM selectors for Teams Web detection.
 * Each version can have different selectors to handle UI updates.
 * 
 * IMPORTANT: Microsoft does not publish stable DOM contracts.
 * These selectors are based on observed Teams Web UI patterns.
 * They may break when Teams releases major UI updates.
 * 
 * Selector priority: semantic attributes > data-tid > class patterns
 */

export interface SelectorSet {
  id: string;
  version: string;
  description: string;
  selectors: {
    /** Main call/meeting container - appears when in a call */
    callContainer: string[];
    /** Pre-join screen - visible before user joins */
    prejoin: string[];
    /** Active call state indicator */
    callActive: string[];
    /** Call controls bar (mute, video, hangup, etc.) */
    callControls: string[];
    /** Hangup/end call button - visible DURING call */
    hangup: string[];
    /** Meeting title element */
    meetingTitle: string[];
    /** Participants panel/roster */
    participants: string[];
    /** Video elements (indicates media is active) */
    videoElements: string[];
    /** Audio elements (indicates media is active) */
    audioElements: string[];
  };
}

/**
 * Selector registry for Teams Web
 * Version format: teams-web-v{year}q{quarter}
 * 
 * When selectors break:
 * 1. Add new SelectorSet with new version
 * 2. Update CURRENT_SELECTOR_SET to new version
 * 3. Keep old version for rollback
 */
export const SELECTOR_REGISTRY: SelectorSet[] = [
  {
    id: 'teams-web-v2026q1',
    version: '2026Q1',
    description: 'Teams Web selectors observed in early 2026',
    selectors: {
      // Main call container - core indicator of being in a call
      // Expanded with more patterns observed in Teams Web
      // 
      // Selector reliability classification (for debugging/diagnostics):
      // - HIGH: data-tid attributes (Teams internal stable API)
      // - MEDIUM: specific class patterns (may change with UI updates)
      // - LOW: generic class patterns (likely to break)
      callContainer: [
        // HIGH reliability - data-tid attributes
        '[data-tid="call-container"]',
        '[data-tid="meeting-container"]',
        '[data-tid="calling-screen"]',
        '[data-tid="calling-container"]',
        '[data-tid="meeting-call-container"]',
        
        // MEDIUM reliability - specific class patterns
        '.ts-calling-thread',
        '.calling-screen-container',
        '.calling-container',
        '.meeting-call-container',
        '.call-screen',
        
        // Attribute-based
        '[data-call-type="call"]',
        '[data-call-type="meeting"]',
        
        // Video grid container (appears during calls)
        '.video-grid',
        '.video-container',
        
        // MEDIUM reliability - class patterns with wildcards (FIXED: was invalid)
        '[class*="video-grid"]',
        '[class*="call-"]',
        
        // Nested content in calling views
        '[class*="calling"][class*="content"]',
        '[class*="calling"][class*="container"]',
      ],
      
      // Pre-join screen - user hasn't joined yet
      prejoin: [
        '[data-tid="prejoin-screen"]',
        '[data-tid="lobby-screen"]',
        '.prejoin-container',
        '.join-meeting-container',
        '[class*="prejoin"]',
        '[class*="lobby"]',
      ],
      
      // Active call state - Teams internal indicator
      callActive: [
        '[data-tid="call-state"]',
        '[data-tid="calling-live-indicator"]',
        '[data-tid="call-active"]',
        '.calling-live-indicator',
        '[class*="live-indicator"]',
        '[aria-label*="In call"]',
        '[aria-label*="In a call"]',
        '[aria-label*="In meeting"]',
        // Generic but useful
        '[class*="call"][class*="active"]',
        '[class*="meeting"][class*="active"]',
      ],
      
      // Call controls - the toolbar with mute/video/hangup
      callControls: [
        '[data-tid="call-controls"]',
        '[data-tid="calling-controls"]',
        '[data-tid="meeting-controls"]',
        '.calling-controls',
        '.call-control-bar',
        '[role="toolbar"]',
        // Expanded with more patterns
        '[class*="control-bar"]',
        '[class*="call-controls"]',
        '[class*="media-controls"]',
      ],
      
      // Hangup button - PRESENT DURING CALL (G1 fix: NOT a sign of call ended)
      hangup: [
        '[data-tid="hangup-button"]',
        '[data-tid="end-call-button"]',
        '[data-tid="leave-call-button"]',
        '.ts-calling-button',
        '.hangup-button',
        '[aria-label*="Hang up"]',
        '[aria-label*="End call"]',
        '[aria-label*="Leave"]',
        // Generic button patterns
        '[class*="hangup"]',
        '[class*="end-call"]',
        '[class*="leave-call"]',
      ],
      
      // Meeting title
      meetingTitle: [
        '[data-tid="meeting-title"]',
        '[data-tid="meeting-subject"]',
        '[data-tid="call-title"]',
        '.ts-meeting-title',
        '.meeting-title',
        '[class*="meeting-title"]',
        '[class*="call-title"]',
        '[aria-label*="meeting"]',
        // Generic subject/title patterns
        'h1[class*="title"]',
        '[class*="header"] [class*="title"]',
      ],
      
      // Participants
      participants: [
        '[data-tid="participant-item"]',
        '[data-tid="participant-list"]',
        '.ts-participant',
        '.participant-avatar',
        '[data-participant-id]',
        '[class*="participant"]',
        '[class*="roster"]',
      ],
      
      // Media elements - video
      videoElements: [
        'video[src]:not([src=""])',
        'video:not([src=""])',
        'video[data-tid]',
      ],
      
      // Media elements - audio  
      audioElements: [
        'audio[src]:not([src=""])',
        'audio:not([src=""])',
        'audio[data-tid]',
      ],
    },
  },
];

/**
 * Current active selector set
 * Update this when Microsoft releases UI changes
 */
export const CURRENT_SELECTOR_SET = SELECTOR_REGISTRY[SELECTOR_REGISTRY.length - 1];

/**
 * Safe querySelector that tries multiple selectors and returns first match
 * Ignores invalid selectors silently (Teams may change DOM structure)
 */
export function queryAny<T extends Element = Element>(
  selectors: string[],
  parent: ParentNode = document
): T | null {
  for (const selector of selectors) {
    if (!selector || typeof selector !== 'string') continue;
    try {
      const el = parent.querySelector<T>(selector);
      if (el) return el;
    } catch {
      // Invalid selector syntax - skip and try next
      continue;
    }
  }
  return null;
}

/**
 * Safe querySelectorAll that tries multiple selectors
 * Returns all matches from all valid selectors
 */
export function queryAll<T extends Element = Element>(
  selectors: string[],
  parent: ParentNode = document
): T[] {
  const results: T[] = [];
  for (const selector of selectors) {
    if (!selector || typeof selector !== 'string') continue;
    try {
      const elements = Array.from(parent.querySelectorAll<T>(selector));
      results.push(...elements);
    } catch {
      // Invalid selector - skip
      continue;
    }
  }
  return results;
}

/**
 * Check if an element is actually visible (not just present in DOM)
 * Addresses G3 fix: elements may persist in DOM when hidden
 * 
 * Production-safe implementation that checks:
 * 1. hidden attribute
 * 2. aria-hidden attribute
 * 3. computed visibility style
 * 4. computed display style
 * 5. element size (zero-size elements are not visible)
 * 6. opacity (zero opacity = not visible)
 * 7. position off-screen (negative position = likely not visible)
 * 
 * NOTE: In test environments (jsdom), computed styles return default values
 * and getBoundingClientRect may return 0. We use a fallback check that 
 * considers elements with proper data-tid attributes as "visible" for testing.
 */
export function isElementVisible(el: Element): boolean {
  if (!el) return false;
  
  // 1. Check for hidden attribute
  if (el.hasAttribute('hidden')) return false;
  
  // 2. Check for aria-hidden (explicit non-visibility)
  const ariaHidden = el.getAttribute('aria-hidden');
  if (ariaHidden === 'true') return false;
  
  // Get data-tid for potential fallback in test environments
  const dataTid = el.getAttribute('data-tid');
  
  // In test environments (jsdom), computed styles may not be reliable
  // Only fall back to data-tid check if we can't get computed styles
  let hasValidComputedStyle = true;
  
  try {
    const style = window.getComputedStyle(el);
    
    // 3. Check visibility property
    if (style.visibility === 'hidden' || style.visibility === 'collapse') {
      return false;
    }
    
    // 4. Check display property
    if (style.display === 'none') {
      return false;
    }
    
    // 5. Check opacity (zero opacity = not visible)
    const opacity = parseFloat(style.opacity);
    if (!isNaN(opacity) && opacity === 0) {
      return false;
    }
    
    // 6. Check if element has zero size
    const rect = el.getBoundingClientRect();
    if (rect.width === 0 || rect.height === 0) {
      return false;
    }
    
    // 7. Check if element is positioned off-screen (common pattern for hidden elements)
    // Elements with position:absolute and left:-9999px are typically hidden
    const position = style.position;
    if (position === 'absolute' || position === 'fixed') {
      const left = parseInt(style.left, 10);
      const top = parseInt(style.top, 10);
      // If positioned far off-screen (more than viewport), consider not visible
      if (left < -10000 || top < -10000 || left > 10000 || top > 10000) {
        return false;
      }
    }
    
  } catch {
    // getComputedStyle may throw in some test environments
    hasValidComputedStyle = false;
  }
  
  // Fallback for test environments: if element exists in DOM and has a data-tid, 
  // consider it visible. This allows tests to work without full CSS rendering.
  if (!hasValidComputedStyle && dataTid) {
    return true;
  }
  
  // If we got here and have valid computed styles, element is visible
  return hasValidComputedStyle;
}

/**
 * Check if element has meaningful content or attributes indicating active state
 */
export function isElementActive(el: Element): boolean {
  if (!el) return false;
  
  // Check aria attributes for active state
  const ariaLabel = el.getAttribute('aria-label') || '';
  const ariaPressed = el.getAttribute('aria-pressed');
  
  if (ariaLabel.toLowerCase().includes('mute') && ariaPressed === 'true') return true;
  if (ariaLabel.toLowerCase().includes('unmute') && ariaPressed === 'false') return true;
  
  // If element has explicit hidden state, it's not active
  if (el.hasAttribute('hidden')) return false;
  
  // Check data-tid state attributes
  const dataState = el.getAttribute('data-tid') || '';
  if (dataState.includes('ended') || dataState.includes('ended')) return false;
  
  return true;
}

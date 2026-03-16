/**
 * Tests for content script entry point.
 */

import { detectPlatform, initialize, cleanup, observer, setup, currentPlatform } from './index';

describe('Content Script', () => {
  // Store original implementations
  const originalQuerySelector = document.querySelector;
  const originalQuerySelectorAll = document.querySelectorAll;

  beforeEach(() => {
    // Reset document mock
    document.body.innerHTML = '';
    
    // Mock chrome.runtime.sendMessage
    (chrome.runtime.sendMessage as jest.Mock) = jest.fn().mockResolvedValue({ success: true });
    
    // Reset readyState
    (document as any).readyState = 'complete';
    
    // Clear all mocks including spies
    jest.clearAllMocks();
  });
  
  afterEach(() => {
    // Restore original implementations
    document.querySelector = originalQuerySelector;
    document.querySelectorAll = originalQuerySelectorAll;
    jest.restoreAllMocks();
  });
  
  describe('detectPlatform', () => {
    it('should detect Google Meet', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
        configurable: true,
      });
      
      const platform = detectPlatform();
      expect(platform).toBe('google-meet');
    });

    it('should detect MS Teams (microsoft.com)', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://teams.microsoft.com/xyz-123' },
        writable: true,
        configurable: true,
      });
      
      const platform = detectPlatform();
      expect(platform).toBe('ms-teams');
    });

    it('should detect MS Teams (live.com)', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://teams.live.com/xyz-123' },
        writable: true,
        configurable: true,
      });
      
      const platform = detectPlatform();
      expect(platform).toBe('ms-teams');
    });

    it('should return null for unsupported platforms', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://example.com' },
        writable: true,
        configurable: true,
      });
      
      const platform = detectPlatform();
      expect(platform).toBeNull();
    });
  });

  describe('initialize', () => {
    it('should handle initialization errors', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
        configurable: true,
      });
      
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      
      // Force an error by mocking document.querySelector to throw
      const querySelectorSpy = jest.spyOn(document, 'querySelector').mockImplementation(() => {
        throw new Error('Query selector error');
      });
      
      initialize();
      
      expect(consoleSpy).toHaveBeenCalledWith('[Verbalizer] Initialization failed:', expect.any(Error));
      
      querySelectorSpy.mockRestore();
      consoleSpy.mockRestore();
    });

    it('should not initialize for unsupported platforms', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://example.com' },
        writable: true,
        configurable: true,
      });
      
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      initialize();
      
      expect(consoleSpy).toHaveBeenCalledWith('[Verbalizer] Not a supported platform');
      consoleSpy.mockRestore();
    });

    it('should initialize for MS Teams', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://teams.microsoft.com/abc' },
        writable: true,
        configurable: true,
      });
      
      initialize();
      
      expect(currentPlatform).toBe('ms-teams');
      expect(observer).not.toBeNull();
    });
  });

  describe('setup', () => {
    it('should initialize immediately if document is complete', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
        configurable: true,
      });
      (document as any).readyState = 'complete';
      
      const addEventListenerSpy = jest.spyOn(document, 'addEventListener');
      
      setup();
      
      expect(addEventListenerSpy).not.toHaveBeenCalledWith('DOMContentLoaded', expect.any(Function));
      expect(observer).not.toBeNull();
    });

    it('should wait for DOMContentLoaded if document is loading', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
        configurable: true,
      });
      (document as any).readyState = 'loading';
      
      const addEventListenerSpy = jest.spyOn(document, 'addEventListener');
      
      setup();
      
      expect(addEventListenerSpy).toHaveBeenCalledWith('DOMContentLoaded', expect.any(Function));
    });
  });

  describe('cleanup', () => {
    it('should disconnect observer and clear it', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
        configurable: true,
      });
      
      initialize();
      
      if (observer) {
        const disconnectSpy = jest.spyOn(observer, 'disconnect');
        cleanup();
        expect(disconnectSpy).toHaveBeenCalled();
        expect(observer).toBeNull();
      } else {
        fail('Observer should have been initialized');
      }
    });

    it('should do nothing if observer is not initialized', () => {
      cleanup();
      expect(observer).toBeNull();
    });
  });
});

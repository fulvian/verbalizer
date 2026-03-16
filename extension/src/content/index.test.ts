import { detectPlatform, initialize } from '../src/content/index';

describe('Content Script', () => {
  describe('detectPlatform', () => {
    it('should detect Google Meet', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
      });

      const platform = detectPlatform();
      expect(platform).toBe('google-meet');
    });

    it('should detect MS Teams', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://teams.microsoft.com' },
        writable: true,
      });

      const platform = detectPlatform();
      expect(platform).toBe('ms-teams');
    });

    it('should return null for unsupported platform', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://example.com' },
        writable: true,
      });

      const platform = detectPlatform();
      expect(platform).toBeNull();
    });
  });

  describe('initialize', () => {
    it('should initialize observer for Google Meet', () => {
      Object.defineProperty(window, 'location', {
        value: { href: 'https://meet.google.com/abc-123' },
        writable: true,
      });

      // Mock document.readyState and addEventListener
      Object.defineProperty(document, 'readyState', {
        value: 'complete',
        writable: true,
      });

      const mockAddEventListener = jest.fn();
      document.addEventListener = mockAddEventListener;

      initialize();

      expect(mockAddEventListener).not.toHaveBeenCalledWith('DOMContentLoaded', expect.any(Function));
    });
  });
});
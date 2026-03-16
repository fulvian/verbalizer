import { detectGoogleMeet, isGoogleMeetActive } from '../src/content/detectors/meet';

describe('Google Meet Detector', () => {
  describe('isGoogleMeetActive', () => {
    it('should return false when no active call indicators', () => {
      // Mock document.querySelector to return null
      document.querySelector = jest.fn().mockReturnValue(null);
      
      const result = isGoogleMeetActive();
      expect(result).toBe(false);
    });

    it('should return true when active call indicators are present', () => {
      // Mock document.querySelector to return a DOM element
      document.querySelector = jest.fn().mockReturnValue(document.createElement('div'));
      
      const result = isGoogleMeetActive();
      expect(result).toBe(true);
    });
  });

  describe('detectGoogleMeet', () => {
    it('should set up interval to check for meeting', () => {
      const observer = {
        notifyCallDetected: jest.fn(),
        notifyCallStarted: jest.fn(),
        notifyCallEnded: jest.fn(),
        registerCleanup: jest.fn(),
      };

      // Mock setInterval and setTimeout
      const mockInterval = setInterval as jest.Mock;
      const mockTimeout = setTimeout as jest.Mock;
      
      detectGoogleMeet(observer as any);
      
      expect(mockInterval).toHaveBeenCalled();
      expect(mockTimeout).toHaveBeenCalled();
    });
  });
});
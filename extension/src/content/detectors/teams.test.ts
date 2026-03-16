import { detectMSTeams, isMSTeamsActive } from '../src/content/detectors/teams';

describe('MS Teams Detector', () => {
  describe('isMSTeamsActive', () => {
    it('should return false when no active call indicators', () => {
      // Mock document.querySelector to return null
      document.querySelector = jest.fn().mockReturnValue(null);
      
      const result = isMSTeamsActive();
      expect(result).toBe(false);
    });

    it('should return true when active call indicators are present', () => {
      // Mock document.querySelector to return a DOM element
      document.querySelector = jest.fn().mockReturnValue(document.createElement('div'));
      
      const result = isMSTeamsActive();
      expect(result).toBe(true);
    });
  });

  describe('detectMSTeams', () => {
    it('should set up interval to check for call', () => {
      const observer = {
        notifyCallDetected: jest.fn(),
        notifyCallStarted: jest.fn(),
        notifyCallEnded: jest.fn(),
        registerCleanup: jest.fn(),
      };

      // Mock setInterval and setTimeout
      const mockInterval = setInterval as jest.Mock;
      const mockTimeout = setTimeout as jest.Mock;
      
      detectMSTeams(observer as any);
      
      expect(mockInterval).toHaveBeenCalled();
      expect(mockTimeout).toHaveBeenCalled();
    });
  });
});
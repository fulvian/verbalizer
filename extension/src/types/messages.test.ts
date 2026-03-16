import { ExtensionMessage, NativeHostRequest } from '../src/types/messages';

describe('ExtensionMessage Type', () => {
  it('should handle CALL_DETECTED message', () => {
    const message: ExtensionMessage = {
      type: 'CALL_DETECTED',
      payload: {
        platform: 'google-meet',
        url: 'https://meet.google.com/abc-123',
        title: 'Test Meeting',
      },
    };

    expect(message.type).toBe('CALL_DETECTED');
    expect(message.payload.platform).toBe('google-meet');
    expect(message.payload.url).toBe('https://meet.google.com/abc-123');
    expect(message.payload.title).toBe('Test Meeting');
  });

  it('should handle CALL_STARTED message', () => {
    const message: ExtensionMessage = {
      type: 'CALL_STARTED',
      payload: {
        platform: 'ms-teams',
        callId: 'test-call-id',
        participants: ['user1@example.com'],
      },
    };

    expect(message.type).toBe('CALL_STARTED');
    expect(message.payload.platform).toBe('ms-teams');
    expect(message.payload.callId).toBe('test-call-id');
    expect(message.payload.participants).toEqual(['user1@example.com']);
  });

  it('should handle CALL_ENDED message', () => {
    const message: ExtensionMessage = {
      type: 'CALL_ENDED',
      payload: {
        platform: 'google-meet',
        callId: 'test-call-id',
        duration: 300,
      },
    };

    expect(message.type).toBe('CALL_ENDED');
    expect(message.payload.platform).toBe('google-meet');
    expect(message.payload.callId).toBe('test-call-id');
    expect(message.payload.duration).toBe(300);
  });
});

describe('NativeHostRequest Type', () => {
  it('should handle START_RECORDING request', () => {
    const request: NativeHostRequest = {
      type: 'START_RECORDING',
      payload: {
        platform: 'google-meet',
        callId: 'test-call-id',
        title: 'Test Meeting',
      },
    };

    expect(request.type).toBe('START_RECORDING');
    expect(request.payload.platform).toBe('google-meet');
    expect(request.payload.callId).toBe('test-call-id');
    expect(request.payload.title).toBe('Test Meeting');
  });

  it('should handle STOP_RECORDING request', () => {
    const request: NativeHostRequest = {
      type: 'STOP_RECORDING',
      payload: { callId: 'test-call-id' },
    };

    expect(request.type).toBe('STOP_RECORDING');
    expect(request.payload.callId).toBe('test-call-id');
  });

  it('should handle GET_STATUS request', () => {
    const request: NativeHostRequest = {
      type: 'GET_STATUS',
      payload: {},
    };

    expect(request.type).toBe('GET_STATUS');
    expect(request.payload).toEqual({});
  });
});
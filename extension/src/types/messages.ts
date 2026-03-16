/**
 * Message types for communication between extension components and 
 * and with the native host.
 */

// ============================================================================
// Content Script -> Background Messages
// ============================================================================

/** Sent when a call is detected on a supported platform */
export interface CallDetectedMessage {
  readonly type: 'CALL_DETECTED';
  readonly payload: {
    platform: 'google-meet' | 'ms-teams';
    url: string;
    title?: string;
  };
}

/** Sent when a call starts (user joins) */
export interface CallStartedMessage{
  readonly type: 'CALL_STARTED';
  readonly payload: {
    platform: 'google-meet' | 'ms-teams';
    callId: string;
    participants?: string[];
  };
}

/** Sent when a call ends( user leaves) */
export interface CallEndedMessage{
  readonly type: 'CALL_ENDED';
  readonly payload: {
    platform: 'google-meet' | 'ms-teams';
    callId: string;
    duration: number; // seconds
  };
}

/** Sent when participant list changes */
export interface ParticipantsChangedMessage{
  readonly type: 'PARTICIPANTS_CHANGED';
  readonly payload: {
    platform: 'google-meet' | 'ms-teams';
    callId: string;
    participants: string[];
    added?: string[];
    removed?: string[];
  };
}

// ============================================================================
// Background -> Native Host Messages
// ============================================================================

/** Start recording request */
export interface StartRecordingRequest{
  readonly type: 'START_RECORDING';
  readonly payload: {
    platform: 'google-meet' | 'ms-teams';
    callId: string;
    title?: string;
  };
}

/** Stop recording request */
export interface StopRecordingRequest{
  readonly type: 'STOP_RECORDING';
  readonly payload: {
    callId: string;
  };
}

/** Get status request */
export interface GetStatusRequest{
  readonly type: 'GET_STATUS';
  readonly payload: Record<string, never>;
}

// ============================================================================
// Native Host -> Background Responses
// ============================================================================

/** Generic response */
export interface NativeHostResponse{
  readonly success: boolean;
  readonly data?: unknown;
  readonly error?: string;
}

/** Recording started response */
export interface RecordingStartedResponse{
  readonly success: true;
  readonly data: {
    callId: string;
    recordingPath: string;
  };
}

/** Status response */
export interface StatusResponse{
  readonly success: true;
  readonly data: {
    isRecording: boolean;
    currentCallId?: string;
    platform?: 'google-meet' | 'ms-teams';
    recordingsDir: string;
    transcriptsDir: string;
  };
}

// ============================================================================
// Union Types
// ============================================================================

export type ExtensionMessage = 
  | CallDetectedMessage
  | CallStartedMessage
  | CallEndedMessage
  | ParticipantsChangedMessage;

export type NativeHostRequest = 
  | StartRecordingRequest
  | StopRecordingRequest
  | GetStatusRequest;

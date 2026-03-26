/**
 * Structured Logging for Verbalizer Extension
 * 
 * Provides consistent cross-layer logging with correlation IDs.
 * Used by content script, background, and passes through native messaging.
 * 
 * @用法 Usage:
 * import { contentLogger } from '../utils/logger';
 * contentLogger.info('Event happened', { callId: 'xxx', event: 'CALL_STARTED' });
 */

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export type LogLayer = 'content' | 'background' | 'native' | 'daemon';

export interface LogEntry {
  ts: string;
  level: LogLevel;
  layer: LogLayer;
  platform: string;
  callId?: string;
  event?: string;
  state?: string;
  confidence?: number;
  reason?: string;
  errorCode?: string;
  message: string;
  metadata?: Record<string, unknown>;
}

/**
 * Generate a unique correlation ID for a call session
 */
export function generateCallId(): string {
  return `call_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

/**
 * Create a structured log entry
 */
export function createLogEntry(
  level: LogLevel,
  layer: LogLayer,
  message: string,
  options?: {
    platform?: string;
    callId?: string;
    event?: string;
    state?: string;
    confidence?: number;
    reason?: string;
    errorCode?: string;
    metadata?: Record<string, unknown>;
  }
): LogEntry {
  const entry: LogEntry = {
    ts: new Date().toISOString(),
    level,
    layer,
    platform: options?.platform || 'teams',
    message,
  };

  if (options?.callId) entry.callId = options.callId;
  if (options?.event) entry.event = options.event;
  if (options?.state) entry.state = options.state;
  if (options?.confidence !== undefined) entry.confidence = options.confidence;
  if (options?.reason) entry.reason = options.reason;
  if (options?.errorCode) entry.errorCode = options.errorCode;
  if (options?.metadata) entry.metadata = options.metadata;

  return entry;
}

/**
 * Format log entry for console output
 */
export function formatLogEntry(entry: LogEntry): string {
  const parts = [
    `[${entry.ts}]`,
    `[${entry.layer.toUpperCase()}]`,
    `[${entry.level.toUpperCase()}]`,
    entry.message,
  ];

  if (entry.callId) parts.push(`callId=${entry.callId}`);
  if (entry.event) parts.push(`event=${entry.event}`);
  if (entry.state) parts.push(`state=${entry.state}`);
  if (entry.confidence !== undefined) parts.push(`conf=${(entry.confidence * 100).toFixed(0)}%`);
  if (entry.reason) parts.push(`reason=${entry.reason}`);
  if (entry.errorCode) parts.push(`errorCode=${entry.errorCode}`);

  return parts.join(' ');
}

/**
 * Logger class with structured logging capabilities
 */
export class StructuredLogger {
  constructor(private layer: LogLayer) {}

  private shouldLog(_level: LogLevel): boolean {
    // In production, filter based on configured level
    // For now, log everything in development
    return true;
  }

  private log(entry: LogEntry): void {
    if (!this.shouldLog(entry.level)) return;

    const formatted = formatLogEntry(entry);

    switch (entry.level) {
      case 'debug':
      case 'info':
        console.log(formatted);
        break;
      case 'warn':
        console.warn(formatted);
        break;
      case 'error':
        console.error(formatted);
        break;
    }

    // Also log the structured object for machine parsing
    console.log('[VERBALIZER]', JSON.stringify(entry));
  }

  debug(message: string, options?: { callId?: string; event?: string; state?: string; confidence?: number; reason?: string; errorCode?: string; metadata?: Record<string, unknown> }): void {
    this.log(createLogEntry('debug', this.layer, message, options));
  }

  info(message: string, options?: { callId?: string; event?: string; state?: string; confidence?: number; reason?: string; errorCode?: string; metadata?: Record<string, unknown> }): void {
    this.log(createLogEntry('info', this.layer, message, options));
  }

  warn(message: string, options?: { callId?: string; event?: string; state?: string; confidence?: number; reason?: string; errorCode?: string; metadata?: Record<string, unknown> }): void {
    this.log(createLogEntry('warn', this.layer, message, options));
  }

  error(message: string, options?: { callId?: string; event?: string; state?: string; confidence?: number; reason?: string; errorCode?: string; metadata?: Record<string, unknown> }): void {
    this.log(createLogEntry('error', this.layer, message, options));
  }

  /**
   * Log call state transition
   */
  logCallState(
    callId: string,
    event: 'CALL_STARTED' | 'CALL_ENDED',
    state: string,
    confidence: number,
    reason?: string
  ): void {
    this.info(`Call ${event}`, {
      callId,
      event,
      state,
      confidence,
      reason,
    });
  }

  /**
   * Log error with error code
   */
  logError(message: string, errorCode: string, callId?: string, metadata?: Record<string, unknown>): void {
    this.error(message, {
      callId,
      errorCode,
      metadata,
    });
  }
}

/**
 * Pre-configured loggers for each layer
 */
export const contentLogger = new StructuredLogger('content');
export const backgroundLogger = new StructuredLogger('background');
export const TRACE_ID_HEADER = 'X-Trace-ID';
export const REQUEST_ID_HEADER = 'X-Request-ID';

export function createTraceId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `trace-${crypto.randomUUID()}`;
  }
  return `trace-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

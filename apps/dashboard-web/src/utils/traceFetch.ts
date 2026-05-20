import { createTraceId, REQUEST_ID_HEADER, TRACE_ID_HEADER } from './trace';

declare global {
  interface Window {
    __flyfleetTraceFetchInit?: boolean;
  }
}

export function initializeTraceFetch(): void {
  if (window.__flyfleetTraceFetchInit) {
    return;
  }

  const originalFetch = window.fetch.bind(window);

  window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const request = input instanceof Request ? input : undefined;
    const headers = new Headers(request?.headers || init?.headers || {});
    
    // Get the target URL string
    const urlString = typeof input === 'string'
      ? input
      : (input instanceof URL ? input.toString() : input.url);

    // Only inject tracing headers for same-origin requests to avoid breaking external CORS preflights (like OSRM)
    const isSameOrigin = !urlString.startsWith('http://') && !urlString.startsWith('https://')
      || urlString.startsWith(window.location.origin);

    if (isSameOrigin) {
      const traceId = createTraceId();
      if (!headers.has(TRACE_ID_HEADER)) {
        headers.set(TRACE_ID_HEADER, traceId);
      }
      if (!headers.has(REQUEST_ID_HEADER)) {
        headers.set(REQUEST_ID_HEADER, traceId);
      }
    }

    const mergedInit: RequestInit = {
      ...init,
      headers,
    };

    return originalFetch(input, mergedInit);
  };

  window.__flyfleetTraceFetchInit = true;
}

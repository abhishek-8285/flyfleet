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
    const traceId = createTraceId();

    if (!headers.has(TRACE_ID_HEADER)) {
      headers.set(TRACE_ID_HEADER, traceId);
    }
    if (!headers.has(REQUEST_ID_HEADER)) {
      headers.set(REQUEST_ID_HEADER, traceId);
    }

    const mergedInit: RequestInit = {
      ...init,
      headers,
    };

    return originalFetch(input, mergedInit);
  };

  window.__flyfleetTraceFetchInit = true;
}

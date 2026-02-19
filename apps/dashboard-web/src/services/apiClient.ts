import axios from 'axios';
import { createTraceId, REQUEST_ID_HEADER, TRACE_ID_HEADER } from '../utils/trace';

// Environment-aware base URL
const API_URL = process.env.REACT_APP_API_URL || '/api/v1';

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // For cookie-based sessions if applicable
});

// Request Interceptor: Attach Auth Token and Org ID
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    const orgId = localStorage.getItem('org_id');

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    if (orgId) {
      config.headers['X-Organization-ID'] = orgId;
    }

    const traceId = createTraceId();
    config.headers[TRACE_ID_HEADER] = traceId;
    config.headers[REQUEST_ID_HEADER] = traceId;

    return config;
  },
  (error) => Promise.reject(error)
);

// Response Interceptor: Handle Global Errors (401, 403, 429)
apiClient.interceptors.response.use(
  (response) => response.data, // Unwrap data directly for cleaner usage
  (error) => {
    const status = error.response?.status;
    const data = error.response?.data;

    // 1. Handle Token Expiry / Unauthorized
    if (status === 401) {
      // Only redirect if not already on login to avoid loops
      if (window.location.pathname !== '/login') {
        // Optional: Attempt Refresh Token flow here before logout
        localStorage.removeItem('access_token');
        window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
      }
    }

    // 2. Handle Forbidden (RBAC)
    if (status === 403) {
      console.error('Permission Denied:', data?.message);
      // Trigger global toast or redirect to 403 page
    }

    // 3. Handle Rate Limiting
    if (status === 429) {
      console.error('Rate Limit Exceeded. Please slow down.');
    }

    // Pass standardized error object downstream
    return Promise.reject({
      status,
      message: data?.error?.message || 'An unexpected error occurred',
      code: data?.error?.code || 'UNKNOWN_ERROR',
      traceId: data?.error?.trace_id
    });
  }
);

export default apiClient;

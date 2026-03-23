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
      config.headers['X-Org-ID'] = orgId;
    }

    const traceId = createTraceId();
    config.headers[TRACE_ID_HEADER] = traceId;

    // Financial Shield: Idempotency keys on all mutations
    if (config.method && ['post', 'put', 'patch', 'delete'].includes(config.method.toLowerCase())) {
        if (!config.headers['X-Request-ID']) {
            config.headers['X-Request-ID'] = (typeof crypto !== 'undefined' && crypto.randomUUID) 
                ? crypto.randomUUID() 
                : Math.random().toString(36).substring(2) + Date.now().toString(36);
        }
    } else {
        config.headers[REQUEST_ID_HEADER] = traceId;
    }

    return config;
  },
  (error) => Promise.reject(error)
);

// Token refresh state
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

// Response Interceptor: Handle Global Errors (401, 403, 429)
apiClient.interceptors.response.use(
  (response) => response.data, // Unwrap data directly for cleaner usage
  async (error) => {
    const originalRequest = error.config;
    const status = error.response?.status;
    const data = error.response?.data;

    // 1. Handle Token Expiry / Unauthorized
    if (status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise(function (resolve, reject) {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            originalRequest._retry = true; // CRITICAL: Prevent queued requests from causing an infinite loop
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return apiClient(originalRequest);
          })
          .catch((err) => {
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          // Send request directly via axios to avoid interceptor loops
          const res = await axios.post(`${API_URL}/auth/refresh`, { refresh_token: refreshToken });
          
          const newToken = res.data?.token || res.data?.access_token;
          const newRefreshToken = res.data?.refresh_token || refreshToken;
          
          if (newToken) {
            localStorage.setItem('access_token', newToken);
            localStorage.setItem('refresh_token', newRefreshToken);
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            processQueue(null, newToken);
            isRefreshing = false;
            return apiClient(originalRequest);
          }
        } catch (refreshError) {
          processQueue(refreshError, null);
          isRefreshing = false;
          if (window.location.pathname !== '/login') {
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
          }
          return Promise.reject(refreshError);
        }
      }
      
      isRefreshing = false;
      // If no refresh token or it failed, fallback to standard logout
      if (window.location.pathname !== '/login') {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
      }
    }

    // Handle 503 SRE Maintenance Global Kill-Switch
    if (status === 503) {
      if (window.location.pathname !== '/system-maintenance') {
        window.location.href = '/system-maintenance';
      }
    }

    // 2. Handle Forbidden (RBAC)
    if (status === 403) {
      console.error('Permission Denied:', data?.message);
      // Trigger global toast or redirect to 403 page
    }

    // 3. Handle Rate Limiting
    if (status === 429) {
      const url = error.config?.url || '';
      if (url.includes('/analytics') || url.includes('/reports')) {
        showSREToast('System load is high. Analytics are temporarily paused to prioritize vehicle tracking.');
      } else {
        console.error('Rate Limit Exceeded. Please slow down.');
      }
    }

    // Pass standardized error object downstream
    return Promise.reject({
      status,
      message: typeof data?.error === 'string' ? data.error : (data?.error?.message || error.message || 'An unexpected error occurred'),
      code: data?.error?.code || 'UNKNOWN_ERROR',
      traceId: data?.error?.trace_id || data?.trace_id
    });
  }
);

// Helper vanilla JS toast for SRE degrade notices without heavy libraries
function showSREToast(message: string) {
  const toast = document.createElement('div');
  toast.innerText = message;
  toast.style.position = 'fixed';
  toast.style.bottom = '20px';
  toast.style.right = '20px';
  toast.style.backgroundColor = '#fef08a'; // yellow-200
  toast.style.color = '#854d0e'; // yellow-800
  toast.style.padding = '12px 20px';
  toast.style.borderRadius = '8px';
  toast.style.boxShadow = '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)';
  toast.style.zIndex = '9999';
  toast.style.fontWeight = '500';
  toast.style.transition = 'opacity 0.4s ease';
  toast.style.fontFamily = 'ui-sans-serif, system-ui, sans-serif';
  document.body.appendChild(toast);
  
  setTimeout(() => {
    toast.style.opacity = '0';
    setTimeout(() => document.body.removeChild(toast), 400);
  }, 4000);
}

export default apiClient;

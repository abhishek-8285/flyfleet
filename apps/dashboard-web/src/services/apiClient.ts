import axios from 'axios';
import { createTraceId, REQUEST_ID_HEADER, TRACE_ID_HEADER } from '../utils/trace';
import { store } from '../store';
import { setToken, setUser } from '../store/authSlice';

// Environment-aware base URL
let API_URL = import.meta.env.VITE_API_URL || '/api/v1';
if (/^https?:\/\/[^\/]+\/?$/.test(API_URL)) {
  API_URL = API_URL.replace(/\/$/, '') + '/api/v1';
}

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Send refresh cookie automatically
});

// Request Interceptor: Attach Auth Token, User, and Org ID
apiClient.interceptors.request.use(
  (config) => {
    // Try Redux store first, fall back to localStorage (survives page refresh)
    const state = store.getState();
    let token = state.auth.token;
    if (!token) {
      token = localStorage.getItem('access_token');
      // Re-hydrate Redux state if we recovered from localStorage
      if (token) {
        store.dispatch(setToken(token));
      }
    }

    // Also recover user from localStorage if Redux doesn't have it yet
    // This is critical for race-condition-free page refresh
    if (!state.auth.user) {
      const savedUser = localStorage.getItem('user_data');
      if (savedUser) {
        try {
          const parsedUser = JSON.parse(savedUser);
          if (parsedUser?.id) {
            store.dispatch(setUser(parsedUser));
            console.log('[apiClient] Recovered user from localStorage');
          }
        } catch (e) {
          console.warn('[apiClient] Failed to parse saved user data');
        }
      }
    }

    // Get tenant_id from localStorage (user preference, not security-critical)
    let tenantId = localStorage.getItem('tenant_id');

    // If tenant_id is not in localStorage but we have user data in Redux, use that
    // Normalization: check both tenant_id and org_id
    if (!tenantId) {
      if ((state.auth.user as any)?.tenant_id) {
        tenantId = (state.auth.user as any).tenant_id;
      } else if (state.auth.user?.tenant_id) {
        tenantId = state.auth.user.tenant_id;
      }
      
      if (tenantId) {
        localStorage.setItem('tenant_id', tenantId);
      }
    }

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    if (tenantId && token) {
      config.headers['X-Tenant-ID'] = tenantId;
      console.log('[apiClient] Authenticated request with tenant_id:', tenantId, 'to:', config.url);
    } else if (tenantId) {
      console.debug('[apiClient] Public request context — suppressing X-Tenant-ID for:', config.url);
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
  (response) => {
    // Extract and store org_id from auth responses
    const responseData = response.data;
    const configUrl = response.config?.url || '';
    
    if (configUrl.includes('/auth/')) {
      // Normalization: Handle both tenant_id and org_id in both flat and nested structures
      const userPayload = responseData?.user || responseData;
      const tenantId = userPayload?.tenant_id || 
                    userPayload?.org_id || 
                    responseData?.organization?.id || 
                    responseData?.organization?.tenant_id ||
                    responseData?.organization?.org_id;

      if (tenantId) {
        localStorage.setItem('tenant_id', tenantId);
        console.log('[apiClient] ✅ Stored normalized tenant_id from', configUrl, ':', tenantId);
      }
    }
    return responseData;
  },
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

      try {
        // Refresh using the stored refresh_token from localStorage
        const refreshToken = localStorage.getItem('refresh_token');
        const tenantId = localStorage.getItem('tenant_id');
        const res = await axios.post(
          `${API_URL}/auth/refresh`,
          { refresh_token: refreshToken },
          {
            headers: {
              'Content-Type': 'application/json',
              'X-Tenant-ID': tenantId || '',
            },
          }
        );

        const newToken = res.data?.token || res.data?.access_token;

        if (newToken) {
          store.dispatch(setToken(newToken));
          // Persist for page refresh survival
          localStorage.setItem('access_token', newToken);
          if (res.data?.refresh_token) {
            localStorage.setItem('refresh_token', res.data.refresh_token);
          }

          // Update user data if returned
          if (res.data?.user) {
            store.dispatch(setUser(res.data.user));
          }

          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          processQueue(null, newToken);
          isRefreshing = false;
          return apiClient(originalRequest);
        } else {
          throw new Error('No token received');
        }
      } catch (refreshError) {
        processQueue(refreshError, null);
        isRefreshing = false;
        if (window.location.pathname !== '/login' && window.location.pathname !== '/session-expired') {
          store.dispatch(setToken(null));
          store.dispatch(setUser(null));
          // Clear cookie on refresh failure
          try {
            await axios.post(`${API_URL}/auth/logout`, {}, { withCredentials: true });
          } catch (e) {
            // Ignore logout errors
          }
          window.location.href = `/session-expired?redirect=${encodeURIComponent(window.location.pathname)}`;
        }
        return Promise.reject(refreshError);
      }
    }

    // Handle 503 SRE Maintenance Global Kill-Switch
    if (status === 503) {
      if (window.location.pathname !== '/system-maintenance') {
        window.location.href = '/system-maintenance';
      }
    }

    // Fix 10: 403 Forbidden — redirect to /403 access-denied page instead of just logging
    if (status === 403) {
      console.error('Permission Denied');
      if (window.location.pathname !== '/403') {
        window.location.href = '/403';
      }
    }

    // Fix 11: Rate Limiting — deduplicated SRE toast
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

// Fix 11: Track active SRE toast to prevent stacking multiple toasts for rapid 429s
let activeSREToast: HTMLElement | null = null;

function showSREToast(message: string) {
  // If a toast is already visible, just update its text rather than stacking
  if (activeSREToast) {
    activeSREToast.innerText = message;
    return;
  }

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
  activeSREToast = toast;

  setTimeout(() => {
    toast.style.opacity = '0';
    setTimeout(() => {
      if (document.body.contains(toast)) {
        document.body.removeChild(toast);
      }
      activeSREToast = null;
    }, 400);
  }, 4000);
}

export default apiClient;

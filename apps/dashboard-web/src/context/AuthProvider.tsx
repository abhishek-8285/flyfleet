import React, { createContext, useContext, useEffect, useState } from "react";
import axios from "axios";
import apiClient from "../services/apiClient";
import { store } from "../store";
import { setToken, setUser as setReduxUser } from "../store/authSlice";
import type { User } from "../types/index";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  authError: string | null;
  login: (token: string, refreshToken: string, userData: User) => void;
  logout: () => void;
  setOrganization: (orgId: string) => void;
  toggleTheme: () => Promise<void>;
  setLanguage: (lang: string) => Promise<void>;
  toggleSidebar: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

/** Apply the given theme to the DOM and localStorage. */
function applyTheme(theme: string) {
  if (theme === "dark") {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
  localStorage.setItem("theme", theme);
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  // Fix 6: expose an explicit authError state so callers can show a proper message
  const [authError, setAuthError] = useState<string | null>(null);

  const hydrateUser = (data: any) => {
    const userWithPermissions: User = {
      ...data.user,
      permissions: data.user.permissions || (data.permissions?.map((p: any) =>
        typeof p === "string" ? p : `${p.module}:${p.action}`
      ) || []),
    } as User;
    store.dispatch(setReduxUser(userWithPermissions));
    // Persist for page refresh survival
    localStorage.setItem('user_data', JSON.stringify(userWithPermissions));
    return userWithPermissions;
  };

  const fetchProfile = async () => {
    try {
      const profileData = await apiClient.get("/auth/profile") as any;

      // CRITICAL: Store org_id from profile response
      const profileOrgId = profileData.user?.org_id;
      if (profileOrgId) {
        localStorage.setItem('org_id', profileOrgId);
        console.log('[AuthProvider] fetchProfile stored org_id:', profileOrgId);
      }

      // Fix 2: Prefer values returned by the backend; fall back to localStorage only when absent
      const backendTheme = profileData.user?.theme;
      const savedTheme = backendTheme || localStorage.getItem("theme") || "light";
      applyTheme(savedTheme);

      const backendLang = profileData.user?.language;
      const savedLang = backendLang || localStorage.getItem("flyfleet_language") || "en";
      localStorage.setItem("flyfleet_language", savedLang);

      const backendSidebar = profileData.user?.sidebar_collapsed;
      const savedSidebar =
        backendSidebar !== undefined
          ? backendSidebar
          : localStorage.getItem("flyfleet_sidebar_collapsed") === "true";
      localStorage.setItem("flyfleet_sidebar_collapsed", String(savedSidebar));

      const userData: User = {
        ...profileData.user,
        permissions:
          profileData.permissions?.map((p: any) =>
            typeof p === "string" ? p : `${p.module}:${p.action}`
          ) || [],
        theme: savedTheme,
        language: savedLang,
        sidebar_collapsed: savedSidebar,
      };

      store.dispatch(setReduxUser(userData));
      setUser(userData);

      // Persist user data for session recovery on page reload
      localStorage.setItem('user_data', JSON.stringify(userData));

      return userData;
    } catch (err) {
      console.error("Failed to fetch full profile:", err);
      throw err;
    }
  };

  // Hydrate state on mount
  useEffect(() => {
    const initAuth = async () => {
      setAuthError(null);
      try {
        console.log('[AuthProvider] Starting auth initialization...');

        // 1. Try localStorage token recovery first (fastest path on page refresh)
        const savedToken = localStorage.getItem('access_token');
        const savedUser = localStorage.getItem('user_data');
        if (savedToken) {
          console.log('[AuthProvider] Found saved token, recovering session');
          store.dispatch(setToken(savedToken));
          if (savedUser) {
            try {
              const parsedUser = JSON.parse(savedUser);
              if (parsedUser.id) {
                store.dispatch(setReduxUser(parsedUser));
                setUser(parsedUser);
                console.log('[AuthProvider] Session recovered from localStorage');
              }
            } catch (e) {
              console.warn('[AuthProvider] Failed to parse saved user data');
            }
          }
          // Verify token is still valid by fetching profile
          try {
            await fetchProfile();
            console.log('[AuthProvider] Auth initialization complete');
            setIsLoading(false);
            return;
          } catch (profileErr) {
            console.warn('[AuthProvider] Saved token expired, attempting refresh');
          }
        }

        // 2. Token expired — try refresh
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          try {
            const response = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken });
            const token = response.data?.token || response.data?.access_token;
            const orgIdFromResponse = response.data?.user?.org_id;

            if (token) {
              store.dispatch(setToken(token));
              localStorage.setItem('access_token', token);
              if (response.data?.refresh_token) {
                localStorage.setItem('refresh_token', response.data.refresh_token);
              }
              if (orgIdFromResponse) {
                localStorage.setItem('org_id', orgIdFromResponse);
              }

              await fetchProfile();
              console.log('[AuthProvider] Session refreshed successfully');
              setIsLoading(false);
              return;
            }
          } catch (refreshErr) {
            console.warn('[AuthProvider] Refresh failed, clearing tokens');
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('user_data');
          }
        }

        // 3. Try Dev Auto-Login if in development (only once per session)
        if (import.meta.env.DEV && !sessionStorage.getItem('flyfleet_dev_autologin_attempted')) {
          try {
            sessionStorage.setItem('flyfleet_dev_autologin_attempted', 'true');
            const devEmail =
              import.meta.env.VITE_DEV_LOGIN_EMAIL || "admin@delta.io";
            console.log(`[AuthProvider] Dev Auto-Login as ${devEmail}`);
            const devResponse = await axios.post(
              "/api/v1/auth/login",
              { email: devEmail, password: "Fleet_password123", org_id: "550e8400-e29b-41d4-a716-446655440000" },
              { withCredentials: false }
            ) as any;
            const devToken = devResponse.data?.token || devResponse.data?.access_token;
            if (devToken) {
              store.dispatch(setToken(devToken));
              localStorage.setItem('access_token', devToken);
              if (devResponse.data?.refresh_token) {
                localStorage.setItem('refresh_token', devResponse.data.refresh_token);
              }
              if (devResponse.data?.user) {
                hydrateUser(devResponse.data);
                if (devResponse.data.user?.org_id) {
                  localStorage.setItem('org_id', devResponse.data.user.org_id);
                }
              }
              await fetchProfile();
              console.log('[AuthProvider] Dev auto-login complete');
              setIsLoading(false);
              return;
            }
          } catch (devErr) {
            console.error("AuthProvider: Dev Auto-Login failed:", devErr);
          }
        }

        // 4. No valid session
        console.log('[AuthProvider] No valid session');
        store.dispatch(setToken(null));
        store.dispatch(setReduxUser(null));
        setUser(null);
      } catch (err) {
        console.error("AuthProvider: Session hydration failed:", err);
        store.dispatch(setToken(null));
        store.dispatch(setReduxUser(null));
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  // Fix 4: Apply theme to the DOM immediately when login() is called manually
  const login = (token: string, refreshToken: string, userData: User) => {
    console.log('[AuthProvider] login() called with:', {
      hasToken: !!token,
      userOrgId: userData?.org_id,
      currentLocalStorageOrgId: localStorage.getItem('org_id')
    });

    store.dispatch(setToken(token));
    store.dispatch(setReduxUser(userData));
    setUser(userData);

    // CRITICAL FIX: Ensure org_id is stored in localStorage immediately
    if (userData?.org_id) {
      localStorage.setItem('org_id', userData.org_id);
      console.log('[AuthProvider] Stored org_id in localStorage:', userData.org_id);
    } else {
      console.error('[AuthProvider] login() called without org_id in userData!');
    }

    // Persist tokens for session recovery across page reloads
    localStorage.setItem('access_token', token);
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken);
    }

    if (userData.theme) {
      applyTheme(userData.theme);
    }
  };

  const logout = async () => {
    try {
      await apiClient.post("/auth/logout", {}, { withCredentials: true });
    } catch (err) {
      console.error("Logout error:", err);
    }
    store.dispatch(setToken(null));
    store.dispatch(setReduxUser(null));
    setUser(null);
    // Clear persisted tokens
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_data');
    localStorage.removeItem('org_id');
    window.location.href = "/login";
  };

  // Fix 5: Avoid full page reload — update context state and re-fetch org-scoped profile
  const setOrganization = async (orgId: string) => {
    localStorage.setItem("org_id", orgId);
    try {
      // Trigger a session refresh with the new org_id
      const refreshToken = localStorage.getItem('refresh_token');
      const refreshResponse = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken });
      const newToken = refreshResponse.data?.token || refreshResponse.data?.access_token;
      if (newToken) {
        store.dispatch(setToken(newToken));
        localStorage.setItem('access_token', newToken);
        if (refreshResponse.data?.refresh_token) {
          localStorage.setItem('refresh_token', refreshResponse.data.refresh_token);
        }
        if (refreshResponse.data?.user) {
          hydrateUser(refreshResponse.data);
        }
      }

      // Re-fetch the authoritative profile (scoped to the new org)
      await fetchProfile();
    } catch (err: any) {
      console.warn("setOrganization: org switch failed, clearing org_id and falling back:", err);
      localStorage.removeItem("org_id");

      if (err.response?.status === 403 || err.response?.status === 401) {
        window.location.href = "/";
      } else {
        window.location.reload();
      }
    }
  };

  const toggleTheme = async () => {
    if (!user) return;
    const newTheme = user.theme === "dark" ? "light" : "dark";
    applyTheme(newTheme);
    setUser({ ...user, theme: newTheme });
    try {
      await apiClient.put(`/admin/users/${user.id}`, { theme: newTheme });
    } catch (err) {
      console.error("Failed to sync theme with API:", err);
    }
  };

  const setLanguage = async (lang: string) => {
    if (!user) return;
    localStorage.setItem("flyfleet_language", lang);
    setUser({ ...user, language: lang });
    try {
      await apiClient.put(`/admin/users/${user.id}`, { language: lang });
    } catch (err) {
      console.error("Failed to sync language with API:", err);
    }
  };

  const toggleSidebar = async () => {
    if (!user) return;
    const newCollapsed = !user.sidebar_collapsed;
    localStorage.setItem("flyfleet_sidebar_collapsed", String(newCollapsed));
    setUser({ ...user, sidebar_collapsed: newCollapsed });
    try {
      await apiClient.put(`/admin/users/${user.id}`, {
        sidebar_collapsed: newCollapsed,
      });
    } catch (err) {
      console.error("Failed to sync sidebar with API:", err);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        authError,
        login,
        logout,
        setOrganization,
        toggleTheme,
        setLanguage,
        toggleSidebar,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

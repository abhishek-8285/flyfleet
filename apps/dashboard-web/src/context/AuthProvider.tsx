import React, { createContext, useContext, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import axios from "axios";
import apiClient from "../services/apiClient";
import { store, type RootState } from "../store";
import { setToken, setUser as setReduxUser, setLoading } from "../store/authSlice";
import type { User } from "../types/index";
import PreferenceService from "../services/PreferenceService";

// Redux is the single source of truth for auth state (token, user, loading, isAuthenticated).
// AuthProvider only provides action functions and persists to localStorage.

interface AuthContextType {
  login: (token: string, refreshToken: string, userData: User) => void;
  logout: () => void;
  setTenant: (tenantId: string) => void;
  toggleTheme: () => Promise<void>;
  setLanguage: (lang: string) => Promise<void>;
  toggleSidebar: () => Promise<void>;
  theme: string;
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
  const [authError, setAuthError] = useState<string | null>(null);
  const user = useSelector((state: RootState) => state.auth.user);
  const theme = user?.theme || localStorage.getItem("theme") || "light";

  // Side-effect: sync 'dark' class with document root whenever theme changes
  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  const hydrateUser = (data: any) => {
    const userPayload = data.user || data;
    const userWithPermissions: User = {
      ...userPayload,
      tenant_id: userPayload.tenant_id || userPayload.org_id || data.organization?.id,
      permissions: userPayload.permissions || (data.permissions?.map((p: any) =>
        typeof p === "string" ? p : `${p.module}:${p.action}`
      ) || []),
    } as User;
    store.dispatch(setReduxUser(userWithPermissions));
    localStorage.setItem('user_data', JSON.stringify(userWithPermissions));
    if (userWithPermissions.tenant_id) {
      localStorage.setItem('tenant_id', userWithPermissions.tenant_id);
    }
    return userWithPermissions;
  };

  const fetchProfile = async () => {
    try {
      const profileData = (await apiClient.get("/auth/profile")) as any;
      if (!profileData) {
        throw new Error("Profile synchronization failed: No data received");
      }

      // Handle both nested { user: {...} } format and flat ProfileResponse format
      const userPayload = profileData.user || profileData;
      const permissionsPayload =
        profileData.permissions || userPayload?.permissions || [];
      const organizationsPayload = profileData.organizations || [];

      if (!userPayload) {
        throw new Error("Profile synchronization failed: Missing identity payload");
      }

      const profileTenantId = userPayload.tenant_id || 
                           userPayload.org_id || 
                           userPayload.current_org?.id || 
                           profileData.organization?.id;

      if (profileTenantId) {
        localStorage.setItem("tenant_id", profileTenantId);
      }

      const backendTheme = userPayload.theme;
      const savedTheme = backendTheme || localStorage.getItem("theme") || "light";
      applyTheme(savedTheme);

      const backendLang = userPayload.language;
      const savedLang = backendLang || localStorage.getItem("flyfleet_language") || "en";
      localStorage.setItem("flyfleet_language", savedLang);

      const backendSidebar = userPayload.sidebar_collapsed;
      const savedSidebar =
        backendSidebar !== undefined
          ? backendSidebar
          : localStorage.getItem("flyfleet_sidebar_collapsed") === "true";
      localStorage.setItem("flyfleet_sidebar_collapsed", String(savedSidebar));

      const userData: User = {
        ...userPayload,
        tenant_id: profileTenantId,
        permissions: permissionsPayload.map((p: any) =>
          typeof p === "string" ? p : `${p.module}:${p.action}`,
        ),
        organizations: organizationsPayload,
        theme: savedTheme,
        language: savedLang,
        sidebar_collapsed: savedSidebar,
      };

      store.dispatch(setReduxUser(userData));
      localStorage.setItem('user_data', JSON.stringify(userData));
      return userData;
    } catch (err) {
      console.error("Failed to fetch full profile:", err);
      throw err;
    }
  };

  // Session recovery on mount — syncs localStorage → Redux
  useEffect(() => {
    const initAuth = async () => {
      setAuthError(null);
      store.dispatch(setLoading(true));
      try {
        const savedToken = localStorage.getItem('access_token');
        const savedUser = localStorage.getItem('user_data');

        if (savedToken && savedUser) {
          try {
            const parsedUser = JSON.parse(savedUser);
            if (parsedUser.id) {
              store.dispatch(setToken(savedToken));
              store.dispatch(setReduxUser(parsedUser));
            }
          } catch (e) {
            console.warn('[AuthProvider] Failed to parse saved user data');
          }

          try {
            await fetchProfile();
            return;
          } catch {
            // Token expired, fall through to refresh
          }
        }

        // Token expired — try refresh
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          try {
            const response = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken });
            const token = response.data?.token || response.data?.access_token;
            if (token) {
              store.dispatch(setToken(token));
              localStorage.setItem('access_token', token);
              if (response.data?.refresh_token) {
                localStorage.setItem('refresh_token', response.data.refresh_token);
              }
              await fetchProfile();
              return;
            }
          } catch {
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            localStorage.removeItem('user_data');
          }
        }

        // Dev auto-login
        if (import.meta.env.DEV && !sessionStorage.getItem('flyfleet_dev_autologin_attempted')) {
          try {
            sessionStorage.setItem('flyfleet_dev_autologin_attempted', 'true');
            const devEmail = import.meta.env.VITE_DEV_LOGIN_EMAIL || "admin@delta.io";
            const devResponse = await axios.post(
              "/api/v1/auth/login",
              { email: devEmail, password: "Fleet_password123", tenant_id: "550e8400-e29b-41d4-a716-446655440000" },
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
                const hydrated = hydrateUser(devResponse.data);
                if (hydrated.tenant_id) {
                  localStorage.setItem('tenant_id', hydrated.tenant_id);
                }
              }
              await fetchProfile();
              return;
            }
          } catch (devErr) {
            console.error("AuthProvider: Dev Auto-Login failed:", devErr);
          }
        }

        // No valid session
        store.dispatch(setToken(null));
        store.dispatch(setReduxUser(null));
      } catch (err) {
        console.error("AuthProvider: Session hydration failed:", err);
        store.dispatch(setToken(null));
        store.dispatch(setReduxUser(null));
      } finally {
        store.dispatch(setLoading(false));
      }
    };

    initAuth();
  }, []);

  const login = (token: string, refreshToken: string, userData: User) => {
    // Normalization: frontend and backend models now use tenant_id
    const normalizedUser = {
      ...userData,
      tenant_id: userData.tenant_id || (userData as any).org_id,
    };

    store.dispatch(setToken(token));
    store.dispatch(setReduxUser(normalizedUser));
    
    if (normalizedUser.tenant_id) {
      localStorage.setItem('tenant_id', normalizedUser.tenant_id);
    }
    
    localStorage.setItem('access_token', token);
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken);
    }
    if (normalizedUser.theme) {
      applyTheme(normalizedUser.theme);
    }
    localStorage.setItem('user_data', JSON.stringify(normalizedUser));
  };

  const logout = async () => {
    try {
      await apiClient.post("/auth/logout", {}, { withCredentials: true });
    } catch {
      // Ignore logout errors
    }
    store.dispatch(setToken(null));
    store.dispatch(setReduxUser(null));
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_data');
    localStorage.removeItem('tenant_id');
    window.location.href = "/login";
  };

  const setTenant = async (tenantId: string) => {
    localStorage.setItem("tenant_id", tenantId);
    try {
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
      await fetchProfile();
    } catch (err: any) {
      console.warn("setTenant: tenant switch failed, clearing tenant_id and falling back:", err);
      localStorage.removeItem("tenant_id");
      if (err.response?.status === 403 || err.response?.status === 401) {
        window.location.href = "/";
      } else {
        window.location.reload();
      }
    }
  };

  const toggleTheme = async () => {
    const user = store.getState().auth.user;
    if (!user) return;
    const newTheme = user.theme === "dark" ? "light" : "dark";
    applyTheme(newTheme);
    const updatedUser = { ...user, theme: newTheme };
    store.dispatch(setReduxUser(updatedUser));
    localStorage.setItem('user_data', JSON.stringify(updatedUser));
    
    // Sync with backend
    await PreferenceService.updateTheme(newTheme);
  };

  const setLanguage = async (lang: string) => {
    const user = store.getState().auth.user;
    if (!user) return;
    localStorage.setItem("flyfleet_language", lang);
    const updatedUser = { ...user, language: lang };
    store.dispatch(setReduxUser(updatedUser));
    localStorage.setItem('user_data', JSON.stringify(updatedUser));
    
    // Sync with backend
    await PreferenceService.updateLanguage(lang);
  };

  const toggleSidebar = async () => {
    const user = store.getState().auth.user;
    if (!user) return;
    const newCollapsed = !user.sidebar_collapsed;
    localStorage.setItem("flyfleet_sidebar_collapsed", String(newCollapsed));
    const updatedUser = { ...user, sidebar_collapsed: newCollapsed };
    store.dispatch(setReduxUser(updatedUser));
    localStorage.setItem('user_data', JSON.stringify(updatedUser));
    
    // Sync with backend
    await PreferenceService.updatePreferences({ sidebar_collapsed: newCollapsed });
  };

  return (
    <AuthContext.Provider
      value={{
        login,
        logout,
        setTenant,
        toggleTheme,
        setLanguage,
        toggleSidebar,
        theme,
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

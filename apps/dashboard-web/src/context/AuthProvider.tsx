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
        
        // 1. Try standard session refresh
        const response = await apiClient.post("/auth/refresh", {}, { withCredentials: true }) as any;
        const token = response.token || response.access_token;
        const orgIdFromResponse = response.user?.org_id;
        
        console.log('[AuthProvider] Refresh response:', { 
          hasToken: !!token, 
          orgIdFromResponse,
          localStorageOrgId: localStorage.getItem('org_id')
        });
        
        if (token) {
          store.dispatch(setToken(token));
        }

        // CRITICAL FIX: Store org_id from refresh response BEFORE calling fetchProfile
        if (orgIdFromResponse) {
          localStorage.setItem('org_id', orgIdFromResponse);
          console.log('[AuthProvider] Stored org_id from refresh:', orgIdFromResponse);
        } else {
          console.warn('[AuthProvider] Refresh succeeded but no org_id in response');
        }

        // Fix 3: No redundant double-hydration. hydrateUser on refresh is only
        // for quick Redux population; fetchProfile is the authoritative source.
        // We skip the intermediate hydrateUser call to avoid two renders with
        // stale/partial data.
        try {
          await fetchProfile();
          console.log('[AuthProvider] Auth initialization complete');
        } catch (profileErr) {
          // Fix 6: fetchProfile failed after a valid refresh — surface the error
          console.error("AuthProvider: Profile fetch failed after refresh:", profileErr);
          setAuthError("Could not load your profile. Please try refreshing.");
          store.dispatch(setToken(null));
          store.dispatch(setReduxUser(null));
          setUser(null);
        }
      } catch (err) {
        // 2. Refresh failed — try Dev Auto-Login if in development
        if (import.meta.env.DEV) {
          try {
            // Fix 1: Dev email comes from env var so each dev can use their own account
            const devEmail =
              import.meta.env.VITE_DEV_LOGIN_EMAIL || "org_admin@delta.io";
            console.log(`AuthProvider: Attempting Dev Auto-Login as ${devEmail}...`);
            // Use axios directly to bypass apiClient interceptors
            // Note: dev-login endpoint removed, using standard login instead
            const devResponse = await axios.post(
              "/api/v1/auth/login",
              { email: devEmail, password: "Fleet_password123", org_id: "" },
              { withCredentials: true }
            ) as any;
            const devToken = devResponse.data?.token || devResponse.data?.access_token;
            if (devToken) {
              store.dispatch(setToken(devToken));
              if (devResponse.data?.user) {
                hydrateUser(devResponse.data);
                // Store org_id from response
                if (devResponse.data.user?.org_id) {
                  localStorage.setItem('org_id', devResponse.data.user.org_id);
                }
              }
              await fetchProfile();
              return; // Success!
            }
          } catch (devErr) {
            console.error("AuthProvider: Dev Auto-Login failed:", devErr);
          }
        }

        // 3. Fallback: No valid session
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
    window.location.href = "/login";
  };

  // Fix 5: Avoid full page reload — update context state and re-fetch org-scoped profile
  const setOrganization = async (orgId: string) => {
    localStorage.setItem("org_id", orgId);
    try {
      // Step A: Trigger a session refresh with the new org_id.
      // This ensures the backend issues a new JWT with the correct org_id claim immediately.
      const refreshResponse = await apiClient.post("/auth/refresh", {}, { withCredentials: true }) as any;
      const newToken = refreshResponse.token || refreshResponse.access_token;
      if (newToken) {
        store.dispatch(setToken(newToken));
        if (refreshResponse.user) {
          hydrateUser(refreshResponse);
        }
      }
      
      // Step B: Re-fetch the authoritative profile (scoped to the new org)
      await fetchProfile();
    } catch (err: any) {
      // If the switch fails, clear the invalid org_id so we don't get stuck in a loop
      console.warn("setOrganization: org switch failed, clearing org_id and falling back:", err);
      localStorage.removeItem("org_id");
      
      // If it's a 403/401, the apiClient might have already handled it, but we'll force a reload
      // to ensure state is clean from the default organization perspective.
      if (err.status === 403 || err.status === 401) {
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

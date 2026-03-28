import React, { createContext, useContext, useEffect, useState } from "react";
import apiClient from "../services/apiClient";
import { store } from "../store";
import { setToken, setUser as setReduxUser } from "../store/authSlice";
import type { User } from "../types/index";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, refreshToken: string, userData: User) => void;
  logout: () => void;
  setOrganization: (orgId: string) => void;
  toggleTheme: () => Promise<void>;
  setLanguage: (lang: string) => Promise<void>;
  toggleSidebar: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Hydrate state on mount: try to refresh session using refresh cookie
  useEffect(() => {
    const initAuth = async () => {
      try {
        // Attempt to refresh session using refresh token cookie (no body needed)
        const response = await apiClient.post("/auth/refresh", {}, { withCredentials: true }) as any;
        const token = response.token || response.access_token;
        if (token) {
          store.dispatch(setToken(token));
        }

        // Store user if returned
        if (response.user) {
          // Transform permissions to string[] if needed
          const userWithPermissions: User = {
            ...response.user,
            permissions: response.user.permissions || (response.permissions?.map((p: any) =>
              typeof p === "string" ? p : `${p.module}:${p.action}`
            ) || []),
          } as User;
          store.dispatch(setReduxUser(userWithPermissions));
        }

        // Then fetch full profile to ensure we have latest user data
        const profileData = await apiClient.get("/auth/profile") as any;
        const userData: User = {
          ...profileData.user,
          permissions:
            profileData.permissions?.map((p: any) =>
              typeof p === "string" ? p : `${p.module}:${p.action}`
            ) || [],
        };

        // Restore UI preferences from localStorage (these are not security-critical)
        const savedTheme = localStorage.getItem("theme") || "light";
        userData.theme = savedTheme;
        if (savedTheme === "dark") {
          document.documentElement.classList.add("dark");
        } else {
          document.documentElement.classList.remove("dark");
        }

        const savedLang = localStorage.getItem("flyfleet_language") || "en";
        userData.language = savedLang;
        localStorage.setItem("flyfleet_language", savedLang);

        const savedSidebar = localStorage.getItem("flyfleet_sidebar_collapsed") === "true";
        userData.sidebar_collapsed = savedSidebar;
        localStorage.setItem("flyfleet_sidebar_collapsed", String(savedSidebar));

        store.dispatch(setReduxUser(userData));
        setUser(userData);
      } catch (err) {
        // No valid session
        console.error("Session hydration failed:", err);
        store.dispatch(setToken(null));
        store.dispatch(setReduxUser(null));
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = (token: string, refreshToken: string, userData: User) => {
    // Store token and user in Redux
    store.dispatch(setToken(token));
    store.dispatch(setReduxUser(userData));
    // Update local state
    setUser(userData);
  };

  const logout = async () => {
    try {
      await apiClient.post("/auth/logout", {}, { withCredentials: true }); // Clears cookie
    } catch (err) {
      console.error("Logout error:", err);
    }
    store.dispatch(setToken(null));
    store.dispatch(setReduxUser(null));
    setUser(null);
    window.location.href = "/login";
  };

  const setOrganization = (orgId: string) => {
    // Store in localStorage for UI persistence only - server validates org access on each request
    localStorage.setItem("org_id", orgId);
    // Reload to apply org context
    window.location.reload();
  };

  const toggleTheme = async () => {
    if (!user) return;
    const newTheme = user.theme === "dark" ? "light" : "dark";

    // Update local DOM
    if (newTheme === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }

    // Update local storage
    localStorage.setItem("theme", newTheme);

    // Optimistic UI update
    setUser({ ...user, theme: newTheme });

    // Sync with API
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

import React, { createContext, useContext, useEffect, useState } from "react";
import apiClient from "../services/apiClient";

interface User {
  id: string;
  name: string;
  email: string;
  role?: string;
  role_id?: string;
  org_id?: string;
  theme?: string;
  language?: string;
  sidebar_collapsed?: boolean;
  permissions?: string[];
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, refreshToken: string, user: User) => void;
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

  // Hydrate state from storage on mount
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem("access_token");
      if (token) {
        try {
          // Verify session effectively by fetching profile
          let profileData;
          try {
            profileData = (await apiClient.get("/auth/profile")) as any;
          } catch (e: any) {
            if (e?.response?.status === 401 || e?.response?.status === 403) {
              throw e;
            }
            console.warn("Backend unreachable, using mock session.");
            profileData = {
              user: JSON.parse(localStorage.getItem("user") || "{}"),
              permissions: ["control_tower:view", "control_tower:edit"],
            };
          }

          // profileData is { user, permissions, orgs }
          const userData: User = {
            ...profileData.user,
            permissions:
              profileData.permissions?.map((p: any) =>
                typeof p === "string" ? p : `${p.module}:${p.action}`,
              ) || [],
          };

          // Hydrate theme from DB or fallback to localStorage
          const savedTheme =
            userData.theme || localStorage.getItem("theme") || "light";
          userData.theme = savedTheme;
          localStorage.setItem("theme", savedTheme);
          if (savedTheme === "dark") {
            document.documentElement.classList.add("dark");
          } else {
            document.documentElement.classList.remove("dark");
          }

          // Hydrate language from DB or fallback to localStorage
          const savedLang =
            userData.language ||
            localStorage.getItem("flyfleet_language") ||
            "en";
          userData.language = savedLang;
          localStorage.setItem("flyfleet_language", savedLang);

          // Hydrate sidebar from DB or fallback to localStorage
          const savedSidebar =
            userData.sidebar_collapsed !== undefined
              ? userData.sidebar_collapsed
              : localStorage.getItem("flyfleet_sidebar_collapsed") === "true";
          userData.sidebar_collapsed = savedSidebar;
          localStorage.setItem(
            "flyfleet_sidebar_collapsed",
            String(savedSidebar),
          );

          setUser(userData);
        } catch (err) {
          // Token invalid
          console.error("Session hydration failed:", err);
          logout();
        }
      }
      setIsLoading(false);
    };

    initAuth();
  }, []);

  const login = (token: string, refreshToken: string, userData: User) => {
    localStorage.setItem("access_token", token);
    localStorage.setItem("refresh_token", refreshToken);
    setUser(userData);
  };

  const logout = () => {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("org_id");
    setUser(null);
    window.location.href = "/login";
  };

  const setOrganization = (orgId: string) => {
    localStorage.setItem("org_id", orgId);
    // Ideally trigger a token refresh to get Org-scoped permissions
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

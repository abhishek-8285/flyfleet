import React, { createContext, useContext, useEffect, useState } from 'react';
import apiClient from '../services/apiClient';

interface User {
  id: string;
  name: string;
  email: string;
  role_id?: string;
  org_id?: string;
  theme?: string;
  permissions?: string[];
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, refreshToken: string, user: User) => void;
  logout: () => void;
  setOrganization: (orgId: string) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Hydrate state from storage on mount
  useEffect(() => {
    const initAuth = async () => {
        const token = localStorage.getItem('access_token');
        if (token) {
            try {
                // Verify session effectively by fetching profile
                const profile = await apiClient.get('/auth/profile');
                setUser(profile as unknown as User); 
            } catch (err) {
                // Token invalid
                logout();
            }
        }
        setIsLoading(false);
    };

    initAuth();
  }, []);

  const login = (token: string, refreshToken: string, userData: User) => {
    localStorage.setItem('access_token', token);
    localStorage.setItem('refresh_token', refreshToken);
    setUser(userData);
  };

  const logout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('org_id');
    setUser(null);
    window.location.href = '/login';
  };

  const setOrganization = (orgId: string) => {
      localStorage.setItem('org_id', orgId);
      // Ideally trigger a token refresh to get Org-scoped permissions
      window.location.reload(); 
  };

  return (
    <AuthContext.Provider value={{ user, isAuthenticated: !!user, isLoading, login, logout, setOrganization }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

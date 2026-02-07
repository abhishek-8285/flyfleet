import React, { createContext, useContext } from 'react';
import { useAuth } from './AuthProvider';

interface RBACContextType {
  can: (action: string, resource: string) => boolean;
  hasRole: (role: string) => boolean;
}

const RBACContext = createContext<RBACContextType | undefined>(undefined);

export const RBACProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user } = useAuth();

  // This relies on the backend sending compiled 'permissions' list in the User object or specific endpoint
  // For now, we assume user.permissions is a list of strings like "vehicle:active", "org:read"
  
  const can = (action: string, resource: string) => {
    if (!user) return false;
    // Super Admin Bypass
    if (user.role_id === 'super_admin_uuid_or_name') return true; 

    const permissionKey = `${resource}:${action}`;
    return user.permissions?.includes(permissionKey) || false;
  };

  const hasRole = (roleName: string) => {
      // Logic depends on if role_id is UUID or name. Assuming Name or mapped ID for simplicity
      return user?.role_id === roleName; 
  };

  return (
    <RBACContext.Provider value={{ can, hasRole }}>
      {children}
    </RBACContext.Provider>
  );
};

export const useRBAC = () => {
  const context = useContext(RBACContext);
  if (!context) {
    throw new Error('useRBAC must be used within an RBACProvider');
  }
  return context;
};

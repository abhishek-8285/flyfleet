import React, { createContext, useContext } from "react";
import { useAuth } from "./AuthProvider";

interface RBACContextType {
  can: (action: string, resource: string) => boolean;
  hasRole: (role: string) => boolean;
}

const RBACContext = createContext<RBACContextType | undefined>(undefined);

export const RBACProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const { user } = useAuth();

  // This relies on the backend sending compiled 'permissions' list in the User object or specific endpoint
  // For now, we assume user.permissions is a list of strings like "vehicle:active", "org:read"

  const can = (action: string, resource: string) => {
    if (!user) {
      console.log("[RBAC] No user found");
      return false;
    }

    // Super Admin Bypass - check for name, ID, or literal string
    const isSuperAdmin =
      user.role_id === "super_admin" ||
      user.role_id === "Super Admin" ||
      user.role_id === "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" || // Known Super Admin UUID
      user.permissions?.includes("*:*");

    if (isSuperAdmin) return true;

    const permissionKey = `${resource}:${action}`;
    const hasPerm = user.permissions?.includes(permissionKey) || false;

    if (!hasPerm) {
      console.log(`[RBAC] Access denied for ${permissionKey}`, {
        userPermissions: user.permissions,
        userRole: user.role_id,
      });
    }

    return hasPerm;
  };

  const hasRole = (roleName: string) => {
    return (
      user?.role_id === roleName ||
      (roleName === "super_admin" &&
        (user?.role_id === "Super Admin" || user?.role_id === "super_admin"))
    );
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
    throw new Error("useRBAC must be used within an RBACProvider");
  }
  return context;
};

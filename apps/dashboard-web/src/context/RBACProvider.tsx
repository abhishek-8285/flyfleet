import React, { createContext, useContext } from "react";
import { useSelector } from "react-redux";
import type { RootState } from "../store";

interface RBACContextType {
  can: (action: string, resource: string) => boolean;
  hasRole: (role: string) => boolean;
}

const RBACContext = createContext<RBACContextType | undefined>(undefined);

const SUPER_ADMIN_SLUGS = ["super_admin", "superadmin", "Super Admin"];

export const RBACProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  // Redux is the single source of truth for auth state
  const user = useSelector((state: RootState) => state.auth.user);

  const isSuperAdmin = () => {
    if (!user) return false;
    return (
      SUPER_ADMIN_SLUGS.includes(user.role ?? "") ||
      SUPER_ADMIN_SLUGS.includes(user.role_id ?? "") ||
      user.permissions?.includes("*:*") === true
    );
  };

  const can = (action: string, resource: string) => {
    if (!user) {
      console.log("[RBAC] No user found");
      return false;
    }

    if (isSuperAdmin()) return true;

    const permissionKey = `${resource}:${action}`;
    const hasPerm = user.permissions?.includes(permissionKey) || false;

    if (!hasPerm) {
      console.log(`[RBAC] Access denied for ${permissionKey}`, {
        userPermissions: user.permissions,
        userRole: user.role,
        userRoleId: user.role_id,
      });
    }

    return hasPerm;
  };

  const hasRole = (roleName: string) => {
    return (
      user?.role === roleName ||
      user?.role_id === roleName ||
      (SUPER_ADMIN_SLUGS.includes(roleName) && isSuperAdmin())
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

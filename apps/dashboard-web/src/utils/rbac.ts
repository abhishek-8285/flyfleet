/**
 * Centralized RBAC utility for the FlyFleet Dashboard.
 * 
 * Standardizes how user roles match against required permissions, 
 * accounting for hierarchical roles and common admin types.
 */

const SUPER_ADMIN_SLUGS = ["super_admin", "superadmin", "super-admin"];

/**
 * Checks if a user's role satisfies any of the allowed roles.
 * 
 * @param userRole - The role of the current user (e.g., 'global_admin', 'fleet_manager').
 * @param allowedRoles - An array of role slugs that are authorized for a resource.
 * @returns true if the user is authorized, false otherwise.
 */
export const isRoleAllowed = (userRole: string | undefined, allowedRoles: string[]): boolean => {
  if (!userRole) return false;
  
  const role = userRole.toLowerCase().trim();
  const allowed = allowedRoles.map(r => r.toLowerCase().trim());

  // 1. Universal access
  if (allowed.includes('*')) return true;

  // 2. Exact match
  if (allowed.includes(role)) return true;

  // 3. Super Admin floor
  // If 'super_admin' is in the allowed list, only super admins get in.
  // If 'admin' is in the allowed list, super admins also get in.
  const isUserSuperAdmin = SUPER_ADMIN_SLUGS.includes(role);
  if (isUserSuperAdmin && (allowed.includes('admin') || allowed.some(r => SUPER_ADMIN_SLUGS.includes(r)))) {
    return true;
  }

  // 4. Inclusive Admin matching
  // If 'admin' is an allowed role, treat any role containing 'admin' as authorized.
  // This covers: global_admin, system_admin, org_admin, etc.
  if (allowed.includes('admin') && role.includes('admin')) {
    return true;
  }

  return false;
};

/**
 * Helper to extract role from user object consistently.
 */
export const extractUserRole = (user: any): string => {
  return (user?.role_id || user?.role || '').toString();
};

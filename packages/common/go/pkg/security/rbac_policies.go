package security

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// Role constants
const (
	RoleSuperAdmin   = "super_admin"
	RoleOrgAdmin     = "org_admin"
	RoleFleetManager = "fleet_manager"
	RoleDriver       = "driver"
	RoleMember       = "member"
)

// Resource constants
const (
	ResourceAuditLogs     = "audit_logs"
	ResourceAuditSettings = "audit_settings"
	ResourceOrganizations  = "organizations"
	ResourceUsers         = "users"
	ResourceVehicles      = "vehicles"
	ResourceTrips         = "trips"
	ResourceAlerts        = "alerts"
	ResourceAnalytics     = "analytics"
	ResourceDocuments     = "documents"
)

// PolicyManifest defines the required roles for each resource
// This centralizes policy definitions to prevent configuration drift
var PolicyManifest = map[string][]string{
	ResourceOrganizations:  {RoleSuperAdmin},
	ResourceAuditLogs:      {RoleSuperAdmin, RoleOrgAdmin},
	ResourceAuditSettings:  {RoleSuperAdmin, RoleOrgAdmin},
	ResourceUsers:         {RoleSuperAdmin, RoleOrgAdmin},
	ResourceVehicles:      {RoleSuperAdmin, RoleOrgAdmin, RoleFleetManager},
	ResourceTrips:         {RoleSuperAdmin, RoleOrgAdmin, RoleFleetManager},
	ResourceAlerts:        {RoleSuperAdmin, RoleOrgAdmin, RoleFleetManager},
	ResourceAnalytics:     {RoleSuperAdmin, RoleOrgAdmin},
	ResourceDocuments:     {RoleSuperAdmin, RoleOrgAdmin, RoleFleetManager},
}

// SecretProvider defines an interface for fetching security secrets
// Implementations can fetch from Env, Vault, or AWS Secrets Manager
type SecretProvider interface {
	GetJWTSecret() string
}

// EnvSecretProvider implements SecretProvider using environment variables
type EnvSecretProvider struct {
	Key string
}

func (p *EnvSecretProvider) GetJWTSecret() string {
	// In a real rotation scenario, this would check a cache or fetch from a KV store
	return p.Key
}

// IsAuthorized checks if a role is permitted to access a resource
func IsAuthorized(role string, resource string) bool {
	allowedRoles, ok := PolicyManifest[resource]
	if !ok {
		return false
	}

	for _, r := range allowedRoles {
		if r == role {
			return true
		}
	}
	return false
}

// GetPolicyVersion returns a stable hash of the current PolicyManifest
// This is used for runtime consistency checks between services
func GetPolicyVersion() string {
	// 1. Sort keys for determinism
	keys := make([]string, 0, len(PolicyManifest))
	for k := range PolicyManifest {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. Build a canonical representation
	canonical := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		roles := PolicyManifest[k]
		sort.Strings(roles)
		canonical = append(canonical, map[string]interface{}{
			"resource": k,
			"roles":    roles,
		})
	}

	data, _ := json.Marshal(canonical)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

package permissions

type Permission uint64

const (
	// Shift positions map to the permission integer scale:
	// 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, etc...
	// Shift positions map to the bit_position in the permissions table (058_iam_seed_permissions.sql)
	
	// Vehicles (0-4)
	PermCreateVehicles   Permission = 1 << 0
	PermReadVehicles     Permission = 1 << 1
	PermUpdateVehicles   Permission = 1 << 2
	PermDeleteVehicles   Permission = 1 << 3
	PermExportVehicles   Permission = 1 << 4
	
	// Drivers (5-9)
	PermCreateDrivers    Permission = 1 << 5
	PermReadDrivers      Permission = 1 << 6
	PermUpdateDrivers    Permission = 1 << 7
	PermDeleteDrivers    Permission = 1 << 8
	PermExportDrivers    Permission = 1 << 9
	
	// Trips (10-14)
	PermCreateTrips      Permission = 1 << 10
	PermReadTrips        Permission = 1 << 11
	PermUpdateTrips      Permission = 1 << 12
	PermDeleteTrips      Permission = 1 << 13
	PermExportTrips      Permission = 1 << 14
	
	// Users (15-18)
	PermCreateUsers      Permission = 1 << 15
	PermReadUsers        Permission = 1 << 16
	PermUpdateUsers      Permission = 1 << 17
	PermDeleteUsers      Permission = 1 << 18
	
	// Roles (19-22)
	PermCreateRoles      Permission = 1 << 19
	PermReadRoles        Permission = 1 << 20
	PermUpdateRoles      Permission = 1 << 21
	PermDeleteRoles      Permission = 1 << 22
	
	// Reports (23-24)
	PermReadReports      Permission = 1 << 23
	PermExportReports    Permission = 1 << 24
	
	// Audit Logs (25-26)
	PermReadAuditLogs    Permission = 1 << 25
	PermExportAuditLogs  Permission = 1 << 26
	
	// Settings (27-28)
	PermReadSettings     Permission = 1 << 27
	PermUpdateSettings   Permission = 1 << 28
	
	// Alerts (29-30)
	PermReadAlerts       Permission = 1 << 29
	PermResolveAlerts    Permission = 1 << 30
	
	// Maintenance (31-33)
	PermCreateMaint      Permission = 1 << 31
	PermReadMaint        Permission = 1 << 32
	PermUpdateMaint      Permission = 1 << 33
	
	// Fuel (34-35)
	PermReadFuel         Permission = 1 << 34
	PermCreateFuel       Permission = 1 << 35
	
	// Billing (36-37)
	PermReadBilling      Permission = 1 << 36
	PermExportBilling    Permission = 1 << 37
	
	// Analytics (38-39)
	PermReadAnalytics    Permission = 1 << 38
	PermExportAnalytics  Permission = 1 << 39
	
	// Integrations (40-41)
	PermReadIntegrations Permission = 1 << 40
	PermManageIntegrations Permission = 1 << 41
	
	// Governance (42-43)
	PermReadGovernance   Permission = 1 << 42
	PermManageGovernance Permission = 1 << 43
)

// HasPermission validates if a specific Permission constant is active inside the BigInt User Mask.
// Example: userMask (9, which is 1001 binary) & required (8, which is 1000 binary) == 8. Returns True.
func HasPermission(userMask, required Permission) bool {
	return (userMask & required) == required
}

// Assign Permission adds to the userMask
func AssignPermission(userMask, perm Permission) Permission {
	return userMask | perm
}

// Revoke Permission removes from the userMask
func RevokePermission(userMask, perm Permission) Permission {
	return userMask &^ perm
}

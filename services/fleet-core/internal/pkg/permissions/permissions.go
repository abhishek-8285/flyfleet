package permissions

type Permission uint64

const (
	// Shift positions map to the permission integer scale:
	// 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, etc...
	PermReadVehicles  Permission = 1 << iota // 1: Can view fleet vehicles
	PermWriteVehicles                        // 2: Can create/update fleet vehicles
	PermReadCosts                            // 4: Can read sensitive financial/fuel data
	PermWriteCosts                           // 8: Can update sensitive financial/fuel data
	PermReadTrips                            // 16: Can view trip assignments and progress
	PermWriteTrips                           // 32: Can dispatch or cancel trips
	PermDeleteTrips                          // 64: Can completely delete a trip record

	// Add more permissions up to 64 limit...
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

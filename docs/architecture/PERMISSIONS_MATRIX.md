# FlyFleet Permissions Matrix

This document provides the canonical map of the 64-bit integer (`Permission uint64`) used to drive the mathematical Authorization (RBAC) engine across FlyFleet.

Because we map roles to an integer bitmask, adding a permission simply shifts `1` to the left by `n` positions.

## Base Architecture
- The system evaluates permission by applying a logical `AND` (`&`).
- E.g., `if (user.PermissionMask & PERM_DELETE_VEHICLE) == PERM_DELETE_VEHICLE`
- The `permission_mask` column in PostgreSQL is a `BIGINT`.

## Bit Definitions

| Bit Position | Hex Value | Go Constant String | Sub-Domain | Description |
| :--- | :--- | :--- | :--- | :--- |
| **0** (`1 << 0`) | `0x01` | `PermissionReadVehicles` | `vehicles` | Allowed to view the list of vehicles and basic metadata. |
| **1** (`1 << 1`) | `0x02` | `PermissionWriteVehicles` | `vehicles` | Allowed to create, edit, or delete non-critical vehicle metadata. |
| **2** (`1 << 2`) | `0x04` | `PermissionReadTrips` | `trips` | Allowed to view historical and active trips/routes. |
| **3** (`1 << 3`) | `0x08` | `PermissionWriteTrips` | `trips` | Allowed to assign drivers to trips or cancel dispatch. |
| **4** (`1 << 4`) | `0x10` | `PermissionReadDrivers` | `drivers` | Allowed to view the driver catalog. |
| **5** (`1 << 5`) | `0x20` | `PermissionWriteDrivers` | `drivers` | Allowed to onboard, suspend, or terminate driver access. |
| **6** (`1 << 6`) | `0x40` | `PermissionReadTelemetry` | `telemetry` | Allowed to query the live and historical GPS/OBD telemetry layer. |
| ... | ... | *Reserved* | ... | *Positions 7-19 reserved for core entities.* |
| **20** (`1 << 20`) | `0x100000` | `PermissionInviteUsers` | `auth` | Allowed to generate `JIT` invitation codes for the current Organization. |
| **21** (`1 << 21`) | `0x200000` | `PermissionManageBilling` | `billing` | Allowed to view payment methods and invoice history. |
| **22** (`1 << 22`) | `0x400000` | `PermissionShadowAccount` | `auth` | Allowed to execute `POST /switch-context` if a `user_link` constraint exists. |
| ... | ... | *Reserved* | ... | *Positions 23-62 reserved for future expansion.* |
| **63** (`1 << 63`) | `0x8000000000000000` | `PermissionSuperAdmin` | `global` | Mathematically supersedes all checks. Has maximum access. |

## Adding a New Permission
When introducing a new permission to the Go `permissions.go` block, you **MUST**:
1. Claim an unused bit from this matrix.
2. Ensure you append `1 << X`.
3. Do not modify or shift the indices of existing permissions, as they are immutably persisted in the database payload `BIGINT`.

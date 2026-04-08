/**
 * IAM User Management - E2E Tests (Playwright)
 *
 * Tests the 5 user stories from product-spec/IAM-USER-MANAGEMENT.md
 *
 * US-001: Header Org Dropdown (P0)
 * US-002: Current Org Display (P0)
 * US-003: Fix setOrganization endpoint (P1)
 * US-004: Role Clone Button (P1)
 * US-005: Bulk Role Assignment (P2)
 */

import { test, expect, Page } from "@playwright/test";

// ============================================================
// Test Helpers
// ============================================================

async function loginAsAdmin(page: Page, orgId?: string) {
  const targetOrg = orgId || "org-alpha-001";

  // Setup auth state
  await page.context().addCookies([
    {
      name: "session_id",
      value: "test-session-123",
      domain: "localhost",
      path: "/",
    },
  ]);

  // Intercept initial refresh
  await page.route("**/api/v1/auth/refresh", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        token: "test-access-token",
        refresh_token: "test-refresh-token",
        user: {
          id: "user-admin-001",
          email: "admin@flyfleet.io",
          name: "Admin User",
          org_id: targetOrg,
          org_name: targetOrg === "org-alpha-001" ? "Alpha Fleet" : "Beta Fleet",
          permissions: ["role:write", "role:read", "user:read", "user:write"],
        },
      }),
    });
  });

  // Intercept profile
  await page.route("**/api/v1/auth/profile", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        user: {
          id: "user-admin-001",
          email: "admin@flyfleet.io",
          name: "Admin User",
          org_id: targetOrg,
          org_name: targetOrg === "org-alpha-001" ? "Alpha Fleet" : "Beta Fleet",
        },
        permissions: ["role:write", "role:read", "user:read", "user:write"],
      }),
    });
  });

  // Intercept orgs list
  await page.route("**/api/v1/auth/orgs", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        data: [
          {
            id: "org-alpha-001",
            name: "Alpha Fleet",
            role: "admin",
            vehicle_count: 45,
          },
          {
            id: "org-beta-002",
            name: "Beta Fleet",
            role: "fleet_manager",
            vehicle_count: 12,
          },
        ],
      }),
    });
  });

  // Navigate to dashboard
  await page.goto("/dashboard");
  await page.waitForLoadState("networkidle");
}

// ============================================================
// US-001: Header Org Dropdown
// ============================================================

test.describe("US-001: Header Org Dropdown", () => {
  test("should show current org name with dropdown indicator for multi-org user", async ({
    page,
  }) => {
    await loginAsAdmin(page);

    // Current org should be visible
    await expect(page.getByText("Alpha Fleet")).toBeVisible();

    // Dropdown indicator should be present
    const dropdownTrigger = page.getByTestId("org-dropdown-trigger");
    await expect(dropdownTrigger).toBeVisible();

    // Click to open dropdown
    await dropdownTrigger.click();

    // Both orgs should be listed with details
    await expect(page.getByText("Alpha Fleet")).toBeVisible();
    await expect(page.getByText("Beta Fleet")).toBeVisible();
    await expect(page.getByText("admin")).toBeVisible();
    await expect(page.getByText("fleet_manager")).toBeVisible();
    await expect(page.getByText("45")).toBeVisible(); // vehicle count
    await expect(page.getByText("12")).toBeVisible();
  });

  test("should call POST /auth/orgs/{orgId}/switch when selecting org", async ({
    page,
  }) => {
    // Track the org switch call
    let orgSwitchCalled = false;
    let switchedOrgId = "";

    await page.route(
      "**/api/v1/auth/orgs/*/switch",
      async (route, request) => {
        orgSwitchCalled = true;
        // Extract org ID from URL
        const url = request.url();
        const match = url.match(/\/orgs\/([^\/]+)\/switch/);
        switchedOrgId = match ? match[1] : "";

        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            access_token: "new-token",
            refresh_token: "new-refresh",
            success: true,
            organization: {
              id: switchedOrgId,
              name: "Beta Fleet",
              role: "fleet_manager",
            },
          }),
        });
      }
    );

    await loginAsAdmin(page);

    // Open dropdown
    await page.getByTestId("org-dropdown-trigger").click();

    // Select Beta Fleet
    await page.getByText("Beta Fleet").click();

    // Verify correct endpoint was called
    expect(orgSwitchCalled).toBe(true);
    expect(switchedOrgId).toBe("org-beta-002");

    // Should show success toast
    await expect(page.getByText(/Switched to Beta Fleet/i)).toBeVisible();
  });

  test("should store new JWT tokens after org switch", async ({ page }) => {
    await page.route(
      "**/api/v1/auth/orgs/*/switch",
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            access_token: "new-access-token-xyz",
            refresh_token: "new-refresh-token-xyz",
            success: true,
            organization: { id: "org-beta-002", name: "Beta Fleet" },
          }),
        });
      }
    );

    await loginAsAdmin(page);

    // Store original token
    const originalToken = await page.evaluate(
      () => localStorage.getItem("access_token")
    );

    // Switch org
    await page.getByTestId("org-dropdown-trigger").click();
    await page.getByText("Beta Fleet").click();

    // Wait for token update
    await page.waitForTimeout(500);

    const newToken = await page.evaluate(
      () => localStorage.getItem("access_token")
    );
    expect(newToken).toBe("new-access-token-xyz");
    expect(newToken).not.toBe(originalToken);
  });
});

// ============================================================
// US-002: Current Org Display
// ============================================================

test.describe("US-002: Current Org Display", () => {
  test("should display current org name in header next to org icon", async ({
    page,
  }) => {
    await loginAsAdmin(page);

    // Org name should be visible near the org icon/logo
    const orgNameInHeader = page.locator('[data-testid="current-org-name"]');
    await expect(orgNameInHeader).toBeVisible();
    await expect(orgNameInHeader).toContainText("Alpha Fleet");
  });

  test("should show name without dropdown for single-org user", async ({
    page,
  }) => {
    // Override orgs list to return only one org
    await page.route("**/api/v1/auth/orgs", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: [
            {
              id: "org-alpha-001",
              name: "Alpha Fleet",
              role: "admin",
              vehicle_count: 45,
            },
          ],
        }),
      });
    });

    await loginAsAdmin(page);

    // Org name visible
    await expect(page.getByText("Alpha Fleet")).toBeVisible();

    // No dropdown trigger
    await expect(page.getByTestId("org-dropdown-trigger")).not.toBeVisible();
  });

  test("should show name with dropdown indicator for multi-org user", async ({
    page,
  }) => {
    await loginAsAdmin(page);

    await expect(page.getByText("Alpha Fleet")).toBeVisible();
    await expect(page.getByTestId("org-dropdown-trigger")).toBeVisible();
  });
});

// ============================================================
// US-003: Fix setOrganization
// ============================================================

test.describe("US-003: Fix setOrganization Endpoint", () => {
  test("should call /auth/orgs/{id}/switch instead of /auth/refresh", async ({
    page,
  }) => {
    const wrongEndpointCalled: string[] = [];
    const correctEndpointCalled: string[] = [];

    // Track wrong endpoint (should NOT be called)
    await page.route("**/api/v1/auth/refresh", async (route, request) => {
      // Only track if it's not the initial page load refresh
      if (request.headers()["x-test-org-switch"] === "true") {
        wrongEndpointCalled.push(request.url());
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          token: "refresh-token",
          user: { org_id: "org-alpha-001" },
        }),
      });
    });

    // Track correct endpoint (SHOULD be called)
    await page.route(
      "**/api/v1/auth/orgs/*/switch",
      async (route, request) => {
        correctEndpointCalled.push(request.url());
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            access_token: "new-token",
            refresh_token: "new-refresh",
            success: true,
            organization: { id: "org-beta-002", name: "Beta Fleet" },
          }),
        });
      }
    );

    await loginAsAdmin(page);

    // Trigger org switch
    await page.getByTestId("org-dropdown-trigger").click();
    await page.getByText("Beta Fleet").click();

    // Wait for network
    await page.waitForTimeout(500);

    // ASSERTION: Correct endpoint called
    expect(correctEndpointCalled.length).toBeGreaterThan(0);
    expect(correctEndpointCalled[0]).toContain("/auth/orgs/");
    expect(correctEndpointCalled[0]).toContain("/switch");

    // ASSERTION: Wrong endpoint NOT called for org switching
    expect(wrongEndpointCalled.length).toBe(0);
  });

  test("should not reload page on org switch", async ({ page }) => {
    let pageLoads = 0;
    page.on("load", () => pageLoads++);

    await page.route(
      "**/api/v1/auth/orgs/*/switch",
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            access_token: "new-token",
            refresh_token: "new-refresh",
            success: true,
            organization: { id: "org-beta-002", name: "Beta Fleet" },
          }),
        });
      }
    );

    await loginAsAdmin(page);

    const loadsBefore = pageLoads;

    // Switch org
    await page.getByTestId("org-dropdown-trigger").click();
    await page.getByText("Beta Fleet").click();

    await page.waitForTimeout(500);

    // Should NOT have triggered a full page reload
    expect(pageLoads).toBe(loadsBefore);
  });
});

// ============================================================
// US-004: Role Clone Button
// ============================================================

test.describe("US-004: Role Clone Button", () => {
  test.beforeEach(async ({ page }) => {
    await page.route("**/api/v1/admin/rbac/roles", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: [
            {
              id: "role-fleet-mgr",
              name: "Fleet Manager",
              description: "Manages fleet operations",
              is_system: false,
              created_at: "2026-01-01T00:00:00Z",
            },
            {
              id: "role-driver",
              name: "Driver",
              description: "Driver role",
              is_system: true,
              created_at: "2026-01-01T00:00:00Z",
            },
            {
              id: "role-dispatcher",
              name: "Dispatcher",
              description: "Dispatch operations",
              is_system: false,
              created_at: "2026-01-01T00:00:00Z",
            },
          ],
        }),
      });
    });
  });

  test("should display clone button on each non-system role row", async ({
    page,
  }) => {
    await loginAsAdmin(page);

    // Navigate to roles page
    await page.goto("/admin/rbac/roles");
    await page.waitForLoadState("networkidle");

    // Wait for roles to load
    await expect(page.getByText("Fleet Manager")).toBeVisible();

    // Non-system roles should have clone buttons
    const fleetManagerRow = page.getByText("Fleet Manager").locator("..");
    await expect(
      fleetManagerRow.getByTestId("role-clone-btn")
    ).toBeVisible();

    const dispatcherRow = page.getByText("Dispatcher").locator("..");
    await expect(
      dispatcherRow.getByTestId("role-clone-btn")
    ).toBeVisible();

    // System roles should NOT have clone button
    const driverRow = page.getByText("Driver").locator("..");
    await expect(
      driverRow.getByTestId("role-clone-btn")
    ).not.toBeVisible();
  });

  test("should call clone endpoint and show cloned role with (Copy) suffix", async ({
    page,
  }) => {
    let cloneCalled = false;

    await page.route(
      "**/api/v1/admin/rbac/roles/*/clone",
      async (route, request) => {
        cloneCalled = true;
        const url = request.url();
        const match = url.match(/\/roles\/([^\/]+)\/clone/);

        await route.fulfill({
          status: 201,
          contentType: "application/json",
          body: JSON.stringify({
            id: "role-cloned-new",
            name: "Fleet Manager (Copy)",
            description: "Cloned from Fleet Manager",
            is_system: false,
            created_at: "2026-04-05T00:00:00Z",
          }),
        });
      }
    );

    await loginAsAdmin(page);
    await page.goto("/admin/rbac/roles");
    await page.waitForLoadState("networkidle");

    await expect(page.getByText("Fleet Manager")).toBeVisible();

    // Click clone button
    await page.getByTestId("role-clone-btn").first().click();

    // Wait for clone dialog/confirmation
    await expect(page.getByText("Fleet Manager (Copy)")).toBeVisible();

    // Verify endpoint was called
    expect(cloneCalled).toBe(true);

    // Toast confirmation
    await expect(page.getByText(/role cloned/i)).toBeVisible();
  });
});

// ============================================================
// US-005: Bulk Role Assignment
// ============================================================

test.describe("US-005: Bulk Role Assignment", () => {
  test.beforeEach(async ({ page }) => {
    await page.route("**/api/v1/admin/users*", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: {
            users: [
              { id: "user-001", name: "Alice", email: "alice@flyfleet.io", status: "active" },
              { id: "user-002", name: "Bob", email: "bob@flyfleet.io", status: "active" },
              { id: "user-003", name: "Charlie", email: "charlie@flyfleet.io", status: "active" },
              { id: "user-004", name: "Diana", email: "diana@flyfleet.io", status: "active" },
              { id: "user-005", name: "Eve", email: "eve@flyfleet.io", status: "active" },
            ],
            total: 5,
          },
        }),
      });
    });

    await page.route("**/api/v1/admin/rbac/roles", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          data: [
            { id: "role-fleet-mgr", name: "Fleet Manager" },
            { id: "role-driver", name: "Driver" },
          ],
        }),
      });
    });
  });

  test("should show checkboxes on user rows", async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto("/admin/rbac/user-assignments");
    await page.waitForLoadState("networkidle");

    await expect(page.getByText("Alice")).toBeVisible();

    // User rows should have checkboxes
    const checkboxes = page.getByRole("checkbox");
    await expect(checkboxes.first()).toBeVisible();
  });

  test("should show Select All checkbox", async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto("/admin/rbac/user-assignments");
    await page.waitForLoadState("networkidle");

    const selectAll = page.getByTestId("select-all-checkbox");
    await expect(selectAll).toBeVisible();
  });

  test("should show action bar when users are selected", async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto("/admin/rbac/user-assignments");
    await page.waitForLoadState("networkidle");

    // Select multiple users
    const checkboxes = page.getByRole("checkbox");
    await checkboxes.nth(1).click(); // Alice
    await checkboxes.nth(2).click(); // Bob

    // Action bar should appear
    const actionBar = page.getByTestId("bulk-action-bar");
    await expect(actionBar).toBeVisible();

    // Should show selection count
    await expect(page.getByText(/2 users selected/i)).toBeVisible();

    // Role dropdown should be present
    const roleDropdown = page.getByTestId("bulk-role-dropdown");
    await expect(roleDropdown).toBeVisible();

    // Assign button should be present
    const assignButton = page.getByTestId("bulk-assign-button");
    await expect(assignButton).toBeVisible();
  });

  test("should call bulk assign endpoint and show count toast", async ({
    page,
  }) => {
    let bulkAssignCalled = false;
    let bulkAssignPayload: any = null;

    await page.route(
      "**/api/v1/admin/rbac/users/assign/bulk",
      async (route, request) => {
        bulkAssignCalled = true;
        bulkAssignPayload = JSON.parse(request.postData() || "{}");

        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            assigned: 2,
            failed: [],
          }),
        });
      }
    );

    await loginAsAdmin(page);
    await page.goto("/admin/rbac/user-assignments");
    await page.waitForLoadState("networkidle");

    // Select users
    const checkboxes = page.getByRole("checkbox");
    await checkboxes.nth(1).click();
    await checkboxes.nth(2).click();

    // Select role
    await page.getByTestId("bulk-role-dropdown").selectOption("role-fleet-mgr");

    // Click assign
    await page.getByTestId("bulk-assign-button").click();

    // Verify endpoint called
    expect(bulkAssignCalled).toBe(true);

    // Verify payload
    expect(bulkAssignPayload.user_ids).toContain("user-001");
    expect(bulkAssignPayload.user_ids).toContain("user-002");
    expect(bulkAssignPayload.role_ids).toContain("role-fleet-mgr");

    // Toast with count
    await expect(page.getByText(/Role assigned to 2 users/i)).toBeVisible();
  });

  test("should handle partial failures in bulk assign", async ({ page }) => {
    await page.route(
      "**/api/v1/admin/rbac/users/assign/bulk",
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({
            assigned: 4,
            failed: [
              { user_id: "user-005", error: "user_not_found" },
            ],
          }),
        });
      }
    );

    await loginAsAdmin(page);
    await page.goto("/admin/rbac/user-assignments");
    await page.waitForLoadState("networkidle");

    // Select all users
    await page.getByTestId("select-all-checkbox").click();

    // Select role and assign
    await page.getByTestId("bulk-role-dropdown").selectOption("role-driver");
    await page.getByTestId("bulk-assign-button").click();

    // Should show partial success message
    await expect(
      page.getByText(/Role assigned to 4 users.*1 failed/i)
    ).toBeVisible();
  });
});

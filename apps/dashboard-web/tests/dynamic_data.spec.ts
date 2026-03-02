import { expect, test } from "@playwright/test";

test.describe("Dashboard Dynamic Data Validation", () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the login page
    await page.goto("http://localhost:3001/login");

    // Perform login
    await page.getByPlaceholder("name@company.com").fill("admin@flyfleet.io");
    await page.getByPlaceholder("••••••••").fill("password123");
    await page.getByRole("button", { name: "Sign In" }).click();

    // Wait for navigation to /admin
    await expect(page).toHaveURL(/.*admin/);
    await page.waitForLoadState("networkidle");
  });

  test("sidebar should load dynamic menu items from API", async ({ page }) => {
    // The sidebar items are dynamic.
    const sidebar = page.locator("aside");
    await expect(sidebar).toBeVisible();

    // Check for some main menu items from GlobalAdminSidebar menuItems API
    // Use more specific selectors to avoid strict mode violations
    await expect(
      sidebar.locator("span.flex-1").getByText("Dashboard", { exact: true }),
    ).toBeVisible({ timeout: 15000 });
    await expect(sidebar.getByText("Fleet", { exact: true })).toBeVisible();
    await expect(sidebar.getByText("Settings", { exact: true })).toBeVisible();
  });

  test("trusted devices should render data from API", async ({ page }) => {
    // In App.tsx, TrustedDevices is at /trusted-devices
    await page.goto("http://localhost:3001/trusted-devices");

    // Explicitly wait for one of the devices to appear
    await page.waitForSelector("text=Workstation DP-402", { timeout: 15000 });

    // Assert all devices are present
    await expect(page.getByText("Workstation DP-402")).toBeVisible();
    await expect(page.getByText("Mobile Fleet Pad 12")).toBeVisible();
    await expect(page.getByText("Personal iPhone Agent")).toBeVisible();
  });

  test("no static placeholders should be visible in the production-like build", async ({
    page,
  }) => {
    // Mock the API to return DIFFERENT data and check that the old static ones are NOT there.
    await page.route("**/api/v1/ui/misc/trusted_devices", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: "99",
            name: "DYNAMIC_TEST_DEVICE_SECRET",
            location: "Mars",
            lastActive: "now",
            isCurrent: true,
            type: "desktop",
          },
        ]),
      });
    });

    await page.goto("http://localhost:3001/trusted-devices");

    await page.waitForSelector("text=DYNAMIC_TEST_DEVICE_SECRET", {
      timeout: 15000,
    });
    await expect(page.getByText("DYNAMIC_TEST_DEVICE_SECRET")).toBeVisible();

    // This confirms that "Workstation DP-402" was NOT hardcoded but came from the API
    await expect(page.getByText("Workstation DP-402")).not.toBeVisible();
  });
});

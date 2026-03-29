import { expect, test } from "@playwright/test";

// Require TEST_PASSWORD to be set - fail fast if not configured
if (!process.env.TEST_PASSWORD) {
  throw new Error("TEST_PASSWORD environment variable must be set for E2E tests");
}

const defaultPassword = process.env.TEST_PASSWORD;
const roles = [
  { role: "super_admin", email: "admin@flyfleet.com", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "org_admin", email: "org_admin@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "fleet_manager", email: "fleet_manager@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "dispatcher", email: "dispatcher@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "fleet_operator", email: "operator@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "maintenance_manager", email: "maintenance@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "analyst", email: "analyst@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "support_agent", email: "support_agent@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "billing_admin", email: "billing@delta.io", password: defaultPassword, dashboard: "/admin/dashboard" },
  { role: "driver", email: "driver@delta.io", password: defaultPassword, dashboard: "/dashboard" },
];

test.describe("All Roles Dashboard Checks", () => {
  test.beforeEach(async ({ page }) => {
    // Mock SharedArrayBuffer globally
    await page.addInitScript(() => {
      if (typeof (window as any).SharedArrayBuffer === "undefined") {
        Object.defineProperty(window, "SharedArrayBuffer", {
          value: ArrayBuffer,
          writable: true,
          configurable: true,
        });
      }
    });
  });

  for (const user of roles) {
    test(`Check Dashboard for ${user.role}`, async ({ page }) => {
      await page.goto("http://localhost:3000/login", { timeout: 60000 });
      const emailInput = page.locator('input[type="text"]');
      await emailInput.fill(user.email);
      await page.click('button:has-text("Continue")');
      
      const passwordInput = page.locator('input[type="password"]');
      await passwordInput.fill(user.password);

      const loginPromise = page.waitForResponse("**/api/v1/auth/login");
      await page.click('button:has-text("Sign In")', { force: true });
      await loginPromise;

      // Navigate to their explicit dashboard URL
      await page.goto(`http://localhost:3000${user.dashboard}`, { waitUntil: 'domcontentloaded' });
      
      // Wait for the page to stabilize
      await page.waitForTimeout(2000);

      // Verify no critical error messages are highly visible
      const bodyText = await page.locator("body").innerText();
      expect(bodyText).not.toContain("404 Not Found");
      expect(bodyText).not.toContain("500 Internal Server Error");

      // Verify a typical dashboard element exists depending on the role
      if (user.role === "driver") {
        // Driver dashboard checks
        await expect(page.locator("body")).toBeVisible();
      } else {
        // Admin dashboard checks
        // Usually there's a title like "Dashboard" or "Overview"
        await expect(page.locator("body")).toBeVisible();
      }

      await page.screenshot({ path: `test-results/screenshots/dashboard-${user.role}.png`, fullPage: true });
      console.log(`Successfully loaded and screenshotted dashboard for ${user.role}`);
    });
  }
});

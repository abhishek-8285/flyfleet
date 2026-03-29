import { expect, test } from "@playwright/test";

// Require TEST_PASSWORD to be set - fail fast if not configured
if (!process.env.TEST_PASSWORD) {
  throw new Error("TEST_PASSWORD environment variable must be set for E2E tests");
}

const roles = [
  { role: "super_admin", email: "admin@flyfleet.com", password: process.env.TEST_PASSWORD },
  { role: "org_admin", email: "org_admin@delta.io", password: process.env.TEST_PASSWORD },
  { role: "fleet_manager", email: "fleet_manager@delta.io", password: process.env.TEST_PASSWORD },
  { role: "dispatcher", email: "dispatcher@delta.io", password: process.env.TEST_PASSWORD },
  { role: "fleet_operator", email: "operator@delta.io", password: process.env.TEST_PASSWORD },
  { role: "maintenance_manager", email: "maintenance@delta.io", password: process.env.TEST_PASSWORD },
  { role: "analyst", email: "analyst@delta.io", password: process.env.TEST_PASSWORD },
  { role: "support_agent", email: "support_agent@delta.io", password: process.env.TEST_PASSWORD },
  { role: "billing_admin", email: "billing@delta.io", password: process.env.TEST_PASSWORD },
  { role: "driver", email: "driver@delta.io", password: process.env.TEST_PASSWORD },
];

test.describe("Deep Dashboard Crawl For All Roles", () => {
  test.beforeEach(async ({ page }) => {
    // Mock SharedArrayBuffer globally for Carto/Mapbox GL
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
    test(`Crawl All Sidebar Links for ${user.role}`, async ({ page }) => {
      test.setTimeout(120000); // Allow 2 minutes per role iteration as they might have many links
      
      await page.goto("http://localhost:3000/login", { timeout: 60000 });
      await page.fill('input[type="text"]', user.email);
      await page.click('button:has-text("Continue")');
      await page.fill('input[type="password"]', user.password);

      const loginPromise = page.waitForResponse("**/api/v1/auth/login");
      await page.click('button:has-text("Sign In")', { force: true });
      await loginPromise;

      // Wait for navigation and sidebar to mount
      await page.waitForTimeout(3000);
      const sidebar = page.locator("aside").first();
      await expect(sidebar).toBeVisible({ timeout: 15000 });

      // Gather all links from the sidebar
      const linkElements = await sidebar.locator("a").elementHandles();
      const urlsToVisit: string[] = [];
      
      for (const link of linkElements) {
        const href = await link.getAttribute("href");
        if (href && !href.startsWith("http") && href !== "#") {
          urlsToVisit.push(href);
        }
      }

      console.log(`[${user.role}] Found ${urlsToVisit.length} sidebar links to verify:`, urlsToVisit);

      // Verify each link
      for (let i = 0; i < urlsToVisit.length; i++) {
        const url = urlsToVisit[i];
        console.log(`[${user.role}] [${i + 1}/${urlsToVisit.length}] Verifying: ${url}`);
        
        await page.goto(`http://localhost:3000${url}`, { waitUntil: 'domcontentloaded' });
        await page.waitForTimeout(1000); // Allow React to throw 404/500 if broken

        const bodyText = await page.locator("body").innerText();
        
        // Assertions: Should not crash into the generic 404 or 500 error screens
        expect(bodyText).not.toContain("404 Not Found");
        expect(bodyText).not.toContain("Internal Server Error");
        expect(bodyText).not.toContain("Oops! The page you're looking for doesn't exist.");
        
        // Basic visibility check
        await expect(page.locator("body")).toBeVisible();
      }
      
      console.log(`[${user.role}] Successfully completed deep crawl of all authorized dashboards!`);
    });
  }
});

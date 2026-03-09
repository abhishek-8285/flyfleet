import { expect, test } from "@playwright/test";

test.describe("FlyFleet Customer Portal E2E", () => {
  test("Customer can login and view shipment dashboard", async ({ page }) => {
    // 1. Login
    await page.goto("http://localhost:3000/login");
    await page.fill("#email", "enterprise@customer.com");
    await page.fill("#password", "password123");
    await page.click('button[type="submit"]');

    // 2. Dashboard landing
    await expect(page).toHaveURL("http://localhost:3000/dashboard");
    await expect(page.locator("h1")).toContainText("Fleet Health");

    // 3. View Shipments
    await page.click("nav >> text=Shipments");
    await expect(page.locator(".shipment-card")).toBeVisible();

    // 4. Download Invoice
    await page.click("text=Invoices");
    const downloadPromise = page.waitForEvent("download");
    await page.click(".download-pdf-btn");
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toContain(".pdf");
  });
});

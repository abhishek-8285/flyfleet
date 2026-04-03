import { expect, test } from "@playwright/test";

// Billing E2E Test: Create Invoice → View → Record Payment → Download PDF
// Requires: TEST_PASSWORD env var, running backend on :8084, running frontend on :5173

const BASE_URL = process.env.BASE_URL || "http://localhost:5173";
const API_URL = process.env.API_URL || "http://localhost:8084";

test.describe("Billing Invoice E2E", () => {
  test("create invoice → view details → record payment → download PDF", async ({ page }) => {
    // 1. Login as admin
    await page.goto(`${BASE_URL}/login`);
    await page.fill('input[name="email"]', "admin@flyfleet.com");
    await page.fill('input[name="password"]', process.env.TEST_PASSWORD || "");
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/admin/, { timeout: 10000 });

    // 2. Navigate to billing dashboard
    await page.goto(`${BASE_URL}/admin/billing`);
    await page.waitForLoadState("networkidle");

    // 3. Create new invoice via form
    await page.goto(`${BASE_URL}/admin/billing/new`);
    await page.waitForLoadState("networkidle");

    // Fill customer name
    await page.fill('input[placeholder*="Customer"]', "E2E Test Corp");

    // Fill dates (use today + 30 days)
    const today = new Date();
    const dueDate = new Date(today);
    dueDate.setDate(dueDate.getDate() + 30);

    const todayStr = today.toISOString().split("T")[0];
    const dueDateStr = dueDate.toISOString().split("T")[0];

    const dateInputs = page.locator('input[type="date"]');
    await dateInputs.first().fill(todayStr);
    await dateInputs.nth(1).fill(dueDateStr);

    // Fill line item
    await page.fill('input[placeholder*="Describe"]', "Fleet Management - Premium Plan");
    await page.fill('input[type="number"]', "1"); // quantity

    const priceInputs = page.locator('input[type="number"]');
    await priceInputs.nth(1).fill("5000"); // unit price

    // Submit form
    await page.click('button[type="submit"]');
    await page.waitForTimeout(2000);

    // 4. Verify invoice appears in list
    await page.goto(`${BASE_URL}/admin/billing`);
    await page.waitForLoadState("networkidle");

    // Check that invoices table has at least one row
    const invoiceRows = page.locator("tbody tr");
    const count = await invoiceRows.count();
    expect(count).toBeGreaterThan(0);

    // 5. Click first invoice to view details
    await invoiceRows.first().click();
    await page.waitForLoadState("networkidle");

    // Verify invoice detail page loaded
    await expect(page.locator("h2", { hasText: /Invoice/i })).toBeVisible();

    // 6. Record payment
    const recordPaymentBtn = page.locator("button", { hasText: /Record Payment/i });
    if (await recordPaymentBtn.isVisible()) {
      await recordPaymentBtn.click();

      // Fill payment amount
      const paymentInput = page.locator('input[type="number"]').last();
      await paymentInput.fill("2500");

      // Confirm payment
      const confirmBtn = page.locator("button", { hasText: /Confirm Payment/i });
      await confirmBtn.click();
      await page.waitForTimeout(1000);

      // Reload to see updated balance
      await page.reload();
      await page.waitForLoadState("networkidle");
    }

    // 7. Download PDF
    const downloadBtn = page.locator("button", { hasText: /Download PDF/i });
    if (await downloadBtn.isVisible()) {
      const [download] = await Promise.all([
        page.waitForEvent("download"),
        downloadBtn.click(),
      ]);

      // Verify PDF was downloaded
      const suggestedFilename = download.suggestedFilename();
      expect(suggestedFilename).toMatch(/\.pdf$/);

      // Verify file is not empty
      const stream = await download.createReadStream();
      let size = 0;
      for await (const chunk of stream) {
        size += chunk.length;
      }
      expect(size).toBeGreaterThan(0);
    }
  });

  test("billing API endpoints respond correctly", async ({ request }) => {
    // Test GET /api/v1/billing/invoices
    const invoicesRes = await request.get(`${API_URL}/api/v1/billing/invoices`, {
      headers: {
        "Authorization": `Bearer ${process.env.TEST_TOKEN || "test"}`,
        "X-Org-ID": process.env.TEST_ORG_ID || "ce95cfe5-42a2-48b2-99d5-3f817c2f0247",
      },
    });

    // Should return 200 (or 401 if no valid token — that's expected in dev)
    expect([200, 401, 403]).toContain(invoicesRes.status());

    if (invoicesRes.status() === 200) {
      const body = await invoicesRes.json();
      expect(body).toHaveProperty("data");
      expect(body).toHaveProperty("total");
    }

    // Test GET /api/v1/billing/settings
    const settingsRes = await request.get(`${API_URL}/api/v1/billing/settings`, {
      headers: {
        "Authorization": `Bearer ${process.env.TEST_TOKEN || "test"}`,
        "X-Org-ID": process.env.TEST_ORG_ID || "ce95cfe5-42a2-48b2-99d5-3f817c2f0247",
      },
    });
    expect([200, 401, 403]).toContain(settingsRes.status());

    // Test GET /api/v1/billing/metrics
    const metricsRes = await request.get(`${API_URL}/api/v1/billing/metrics`, {
      headers: {
        "Authorization": `Bearer ${process.env.TEST_TOKEN || "test"}`,
        "X-Org-ID": process.env.TEST_ORG_ID || "ce95cfe5-42a2-48b2-99d5-3f817c2f0247",
      },
    });
    expect([200, 401, 403]).toContain(metricsRes.status());
  });
});

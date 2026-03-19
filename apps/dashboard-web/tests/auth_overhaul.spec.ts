import { expect, test } from "@playwright/test";

test.describe("Auth & Registration UI Overhaul Tests", () => {
  // We use a unique timestamp to prevent email collisions during testing
  const runId = Date.now();

  test.beforeEach(async ({ page }) => {
    // Navigate straight to the registration screen
    await page.goto("http://localhost:3000/register");
  });

  test("Scenario 1: Complete Registration - Create New Organization", async ({ page }) => {
    console.log("Starting Registration Flow - Create Org...");

    // Step 1: Basic Info
    await page.fill('input[placeholder="John Doe"]', `Test Admin ${runId}`);
    await page.fill('input[type="email"]', `test_admin_${runId}@flyfleet.com`);
    await page.locator('input[type="password"]').first().fill("SecurePassword123!");
    
    // Check if confirm password exists, typically placeholders are used
    const confirmInput = page.locator('input[type="password"]').nth(1);
    if (await confirmInput.isVisible()) {
      await confirmInput.fill("SecurePassword123!");
    }

    await page.click('button:has-text("Continue to Organization")');

    // Step 2: Org Handling
    // Wait for the form transition
    await expect(page.locator('text="Organization Name"')).toBeVisible({ timeout: 5000 });

    // Select "Create Organization" if it's a toggle/radio. Since default is create, we just fill it.
    await page.fill('input[placeholder="Acme Logistics Inc."]', `Test Org ${runId}`);

    // Let's intercept the API call to ensure it actually routes correctly.
    const registerPromise = page.waitForResponse("**/api/v1/auth/register");
    await page.click('button:has-text("Complete Registration")');

    const response = await registerPromise;
    expect(response.status()).toBe(201); // Created

    // Expect to be routed to admin or dashboard safely
    await page.waitForURL("**/admin*", { timeout: 10000 });
  });

  test("Scenario 2: Registration - Join via Invalid Invite Code", async ({ page }) => {
    console.log("Starting Registration Flow - Invalid Invite...");

    // Step 1: Basic Info
    await page.fill('input[placeholder="John Doe"]', `Test Invitee ${runId}`);
    await page.fill('input[type="email"]', `test_invitee_${runId}@flyfleet.com`);
    await page.locator('input[type="password"]').first().fill("SecurePassword123!");
    
    const confirmInput = page.locator('input[type="password"]').nth(1);
    if (await confirmInput.isVisible()) {
      await confirmInput.fill("SecurePassword123!");
    }

    await page.click('button:has-text("Continue to Organization")');
    await expect(page.locator('button:has-text("Join via Invite")')).toBeVisible({ timeout: 5000 });

    // Switch to Invite Mode
    await page.click('button:has-text("Join via Invite")');

    // The frontend debounces for 600ms then hits `/api/v1/auth/invites/validate`
    const validationPromise = page.waitForResponse(response => 
      response.url().includes("/api/v1/auth/invites/validate") && 
      response.request().method() === "POST"
    );
    
    // Simulate typing an invalid UUID/Code
    await page.fill('input[placeholder="XXXXXXXX-XXXX-4XXX-YXXX-XXXXXXXXXXXX"]', "invalid-code-1234567890");
    
    // Wait for debounce and response
    const validationResponse = await validationPromise;
    const status = validationResponse.status();
    
    // The backend returns 200 OK with { valid: false } for security obfuscation
    if (status === 200) {
      const json = await validationResponse.json();
      expect(json.valid).toBe(false);
    } else {
      expect(status).toBe(429); // HTTP 429 Too Many Requests
    }

    // Assert the UI blocks progression and displays the error state
    await expect(page.locator('text=/Invalid or expired invite code|Too many validation/')).toBeVisible({ timeout: 5000 });
  });
});

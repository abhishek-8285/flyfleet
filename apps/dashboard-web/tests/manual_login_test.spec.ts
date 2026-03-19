import { test, expect } from '@playwright/test';

test('End-to-End Login for abhishek@test1.com', async ({ page }) => {
  console.log("Navigating to login...");
  await page.goto("http://localhost:3000/login", { timeout: 60000 });
  
  console.log("Filling email...");
  await page.fill('input[type="text"], input[name="email"], input[placeholder*="email" i]', "abhishek@test1.com");
  
  // The login form might be single-step or multi-step. 
  // Let's try clicking Continue if it exists, otherwise just fill password directly.
  const continueBtn = page.locator('button:has-text("Continue")');
  if (await continueBtn.count() > 0 && await continueBtn.isVisible()) {
    console.log("Clicking Continue...");
    await continueBtn.click();
  }
  
  console.log("Filling password...");
  await page.fill('input[type="password"], input[name="password"]', "Abhishek@123");
  
  console.log("Clicking Sign In...");
  const loginPromise = page.waitForResponse("**/api/v1/auth/login", { timeout: 15000 }).catch(e => {
     console.log("Warning: Could not capture /api/v1/auth/login response, maybe URL differs?", e);
     return null;
  });
  
  // Try different sign in button texts
  const signInBtn = page.locator('button:has-text("Sign In"), button:has-text("Login"), button:has-text("Log In")').first();
  await signInBtn.click();
  
  const response = await loginPromise;
  if (response) {
     console.log(`Login API resolved with status: ${response.status()}`);
     expect(response.status()).toBe(200);
  }

  console.log("Waiting for dashboard redirect...");
  // Wait for network idle or a specific URL
  await page.waitForURL(
    (url) => {
      const path = url.pathname.replace(/\/$/, "");
      return path === "/admin" || path === "/dashboard" || path === "/admin/dashboard" || path.includes('dashboard') || path.includes('admin');
    },
    { timeout: 15000 }
  );

  console.log(`Successfully redirected to: ${page.url()}`);
  
  await page.screenshot({ path: "abhishek_login_success.png" });
  console.log("Screenshot captured to abhishek_login_success.png !");
});

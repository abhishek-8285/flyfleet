const { chromium } = require('@playwright/test');

(async () => {
  console.log("Launching browser...");
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  page.on('console', msg => console.log('PAGE LOG:', msg.text()));
  page.on('pageerror', exception => console.log('PAGE ERROR:', exception));

  console.log("Navigating to login...");
  await page.goto('http://localhost:3000/login');
  await page.fill('input[type="email"]', 'admin@flyfleet.com');
  await page.fill('input[type="password"]', 'fleet_password');
  
  console.log("Submitting login...");
  const loginPromise = page.waitForResponse("**/api/v1/auth/login");
  await page.click('button:has-text("Sign In")', { force: true });
  const res = await loginPromise;
  console.log("Login response status:", res.status());

  console.log("Waiting 2s...");
  await page.waitForTimeout(2000);

  console.log("Navigating to control tower...");
  await page.goto('http://localhost:3000/admin/control-tower');

  console.log("Waiting for __CERTIFICATION_CONTROLLER__...");
  try {
    await page.waitForFunction(() => (window).__CERTIFICATION_CONTROLLER__ !== undefined, { timeout: 10000 });
    console.log("CONTROLLER FOUND!");
  } catch(e) {
    console.log("CONTROLLER NOT FOUND!", e.message);
  }

  console.log("Triggering startStorm(50000)...");
  await page.evaluate(() => {
    const ctrl = (window).__CERTIFICATION_CONTROLLER__;
    if (ctrl) ctrl.startStorm(50000);
  });

  console.log("Wait 5s...");
  await page.waitForTimeout(5000);

  console.log("Fetching metrics...");
  const metrics = await page.evaluate(() => {
    return (window).__CERTIFICATION_CONTROLLER__.getMetrics();
  });
  console.log("Metrics:", metrics);

  console.log("Done.");
  await browser.close();
})();

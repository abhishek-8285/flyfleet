const { chromium } = require('playwright');
(async () => {
    const browser = await chromium.launch();
    const page = await browser.newPage();
    await page.goto('http://127.0.0.1:3001/login');
    await page.fill('input[type="email"]', 'fleet_manager@delta.io');
    await page.fill('input[type="password"]', 'fleet_password');
    const responsePromise = page.waitForResponse("**/api/v1/auth/login");
    await page.click('button:has-text("Sign In")');
    const response = await responsePromise;
    console.log("Status:", response.status());
    console.log("Body:", await response.text());
    await browser.close();
})();

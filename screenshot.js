const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('http://localhost:3000');
  
  // Wait for map to load
  await page.waitForTimeout(5000);
  
  await page.screenshot({ path: 'map_screenshot.png' });
  await browser.close();
})();

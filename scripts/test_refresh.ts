import { chromium } from 'playwright';

async function main() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();
  
  await page.goto('http://localhost:3000/login');
  
  // Fill login form
  await page.waitForSelector('input[type="text"]', { state: 'visible', timeout: 5000 });
  await page.fill('input[type="text"]', 'org_admin@delta.io');
  await page.fill('input[type="password"]', 'fleet_password');
  
  // Click login button
  await page.click('button[type="submit"]');

  // Check if Org Selection appears
  try {
    await page.waitForSelector('text=Select Workspace', { timeout: 2000 });
    console.log('Org selection required');
    // Click the first workspace
    await page.click('button.group');
  } catch (e) {
    // No org selection required
  }

  // Wait for navigation
  await page.waitForURL('**/(dashboard|admin|portal)**', { timeout: 15000 }).catch(() => console.log('Did not hit exact URL, continuing...'));
  
  // Wait for network idle and rendering
  await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => null);

  // get cookies
  const cookies = await context.cookies();
  console.log('Cookies after login:', cookies.map(c => ({ name: c.name, targetPath: c.path, domain: c.domain, secure: c.secure, httpOnly: c.httpOnly, sameSite: c.sameSite })));
  
  // try to refresh using axios configuration from the app context or directly fetch
  const result = await page.evaluate(async () => {
    try {
      const resp = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'using' // actually include is standard, let's use include
      });
      return { status: resp.status, body: await resp.text() };
    } catch(e: any) {
      return { error: e.message };
    }
  });

  console.log('Refresh result (fetch without include):', result);
  
  const result2 = await page.evaluate(async () => {
    try {
      const resp = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
      });
      return { status: resp.status, body: await resp.text() };
    } catch(e: any) {
      return { error: e.message };
    }
  });
  console.log('Refresh result (fetch with include):', result2);

  await browser.close();
}

main().catch(console.error);

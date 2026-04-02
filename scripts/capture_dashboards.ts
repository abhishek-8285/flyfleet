import { chromium } from 'playwright';
import path from 'path';

const USERS = [
  { role: 'super_admin', email: 'admin@flyfleet.com' },
  { role: 'org_admin', email: 'org_admin@delta.io' },
  { role: 'fleet_manager', email: 'fleet_manager@delta.io' },
  { role: 'dispatcher', email: 'dispatcher@delta.io' },
  { role: 'fleet_operator', email: 'operator@delta.io' },
  { role: 'maintenance_manager', email: 'maintenance@delta.io' },
  { role: 'analyst', email: 'analyst@delta.io' },
  { role: 'support_agent', email: 'support_agent@delta.io' },
  { role: 'billing_admin', email: 'billing@delta.io' },
  { role: 'driver', email: 'driver@delta.io' }
];

const PASS = 'fleet_password';
const URL = 'http://localhost:3000';

async function main() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });

  for (const user of USERS) {
    console.log(`Processing ${user.role}...`);
    const page = await context.newPage();
    try {
      await page.goto(`${URL}/login`);
      
      // Fill login form
      await page.waitForSelector('input[type="text"]', { state: 'visible', timeout: 5000 });
      await page.fill('input[type="text"]', user.email);
      await page.fill('input[type="password"]', PASS);
      
      // Click login button
      await page.click('button[type="submit"]');

      // Check if Org Selection appears
      try {
        await page.waitForSelector('text=Select Workspace', { timeout: 2000 });
        console.log(`Org selection required for ${user.email}`);
        // Click the first workspace
        await page.click('button.group');
      } catch (e) {
        // No org selection required
      }

      // Wait for navigation
      await page.waitForURL('**/(dashboard|admin|portal)**', { timeout: 15000 }).catch(() => console.log('Did not hit exact URL, continuing...'));
      
      // Wait for network idle and rendering
      await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => null);
      await new Promise(r => setTimeout(r, 3000));

      const screenshotPath = path.join(__dirname, '..', `dashboard_${user.role}.png`);
      await page.screenshot({ path: screenshotPath, fullPage: true });
      console.log(`Saved screenshot to ${screenshotPath}`);
      
      await context.clearCookies();
    } catch (e) {
      console.error(`Error processing ${user.role}:`, e);
    } finally {
      await page.close();
    }
  }

  await browser.close();
}

main().catch(console.error);

const fs = require('node:fs/promises');
const path = require('node:path');
const { chromium } = require('playwright');

const BASE_URL = 'http://127.0.0.1:3000';
const OUT_DIR = '/home/abhishek/Desktop/flyfleet/artifacts/role-page-screenshots';

const roles = [
  { role: 'super_admin', email: 'admin@flyfleet.com', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'org_admin', email: 'org_admin@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'fleet_manager', email: 'fleet_manager@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'dispatcher', email: 'dispatcher@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'fleet_operator', email: 'operator@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'maintenance_manager', email: 'maintenance@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'analyst', email: 'analyst@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'support_agent', email: 'support_agent@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'billing_admin', email: 'billing@delta.io', password: 'fleet_password', dashboard: '/admin/dashboard' },
  { role: 'driver', email: 'driver@delta.io', password: 'fleet_password', dashboard: '/dashboard' },
];

const DEMO_ID = '00000000-0000-0000-0000-000000000001';

function normalizePath(p) {
  return (p || '/').replace(/\/+$/, '') || '/';
}

function slugifyRoute(route) {
  return route
    .replace(/^\//, '')
    .replace(/\//g, '__')
    .replace(/[^a-zA-Z0-9_\-]/g, '_') || 'root';
}

function classify(bodyText, finalPath) {
  const t = (bodyText || '').toLowerCase();
  if (finalPath.startsWith('/login')) return 'redirect_login';
  if (t.includes('403') || t.includes('access denied') || t.includes('permission')) return 'forbidden';
  if (t.includes('404 not found') || t.includes("page you're looking for doesn't exist")) return 'not_found';
  if (t.includes('500 internal server error') || t.includes('internal server error')) return 'server_error';
  return 'ok';
}

function materializeDynamicRoute(route) {
  return route
    .replace(/:id/g, DEMO_ID)
    .replace(/:org_id/g, DEMO_ID)
    .replace(/:userId/g, DEMO_ID)
    .replace(/:token/g, 'demo-token');
}

async function collectRouteInventory() {
  const appPath = '/home/abhishek/Desktop/flyfleet/apps/dashboard-web/src/App.tsx';
  const src = await fs.readFile(appPath, 'utf8');
  const raw = new Set(src.match(/path=\"([^\"]+)\"/g) || []);
  const values = [...raw].map((x) => x.replace(/^path=\"/, '').replace(/\"$/, ''));
  const routes = new Set();

  for (const p of values) {
    if (!p || p === '*') continue;
    if (p.startsWith('/')) {
      routes.add(materializeDynamicRoute(p));
      continue;
    }
    // Most relative routes in App.tsx are mounted under /admin.
    routes.add(materializeDynamicRoute(`/admin/${p}`));
  }

  routes.add('/admin/dashboard');
  routes.add('/dashboard');
  routes.add('/admin/control-tower');
  routes.add('/admin/vehicles');
  routes.add('/admin/drivers');
  routes.add('/admin/dispatch');

  return [...routes]
    .filter((r) => !r.includes('*'))
    .sort();
}

async function ensureDir(p) {
  await fs.mkdir(p, { recursive: true });
}

async function writeCsv(filePath, rows) {
  const header = ['role', 'email', 'requested_route', 'final_url', 'status', 'screenshot'];
  const lines = [header.join(',')];
  for (const r of rows) {
    const vals = [r.role, r.email, r.requestedRoute, r.finalUrl, r.status, r.screenshot].map((v) => {
      const s = String(v ?? '');
      if (s.includes(',') || s.includes('"') || s.includes('\n')) {
        return '"' + s.replace(/"/g, '""') + '"';
      }
      return s;
    });
    lines.push(vals.join(','));
  }
  await fs.writeFile(filePath, lines.join('\n'));
}

async function writeHtmlReport(filePath, rows) {
  const cards = rows.map((r) => {
    const rel = path.relative(path.dirname(filePath), r.screenshot).replaceAll('\\', '/');
    return `<tr>
<td>${r.role}</td>
<td><code>${r.requestedRoute}</code></td>
<td>${r.status}</td>
<td><a href="${rel}" target="_blank">open</a></td>
</tr>`;
  }).join('\n');

  const html = `<!doctype html>
<html><head><meta charset="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Role Page Access Matrix</title>
<style>
body{font-family:Segoe UI,Arial,sans-serif;background:#071425;color:#e9f3ff;margin:0;padding:20px}
h1{margin:0 0 10px}
small{color:#9bb5d1}
table{width:100%;border-collapse:collapse;margin-top:14px;background:#0d223a}
th,td{border:1px solid #244767;padding:8px;text-align:left;font-size:14px}
th{background:#123152}
code{background:#0b1b2d;padding:2px 6px;border-radius:5px}
a{color:#9fd0ff}
</style></head><body>
<h1>Role to Page Access Matrix</h1>
<small>Generated from automated crawl of sidebar links per role.</small>
<table>
<tr><th>Role</th><th>Requested Route</th><th>Status</th><th>Screenshot</th></tr>
${cards}
</table>
</body></html>`;

  await fs.writeFile(filePath, html);
}

async function collectSidebarRoutes(page, fallbackRoute) {
  const links = await page.locator('aside a').evaluateAll((els) =>
    els.map((e) => e.getAttribute('href')).filter(Boolean)
  ).catch(() => []);

  const routes = new Set([fallbackRoute]);
  for (const href of links) {
    if (!href) continue;
    if (href.startsWith('http')) continue;
    if (!href.startsWith('/')) continue;
    if (href === '#') continue;
    routes.add(href);
  }
  return [...routes];
}

async function login(page, role) {
  await page.goto(`${BASE_URL}/login`, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.fill('input[type="text"]', role.email);
  await page.click('button:has-text("Continue")');
  await page.fill('input[type="password"]', role.password);

  await Promise.all([
    page.waitForResponse((resp) => resp.url().includes('/api/v1/auth/login'), { timeout: 60000 }),
    page.click('button:has-text("Sign In")', { force: true }),
  ]);

  await page.waitForTimeout(2500);
}

async function run() {
  await ensureDir(OUT_DIR);
  const inventory = await collectRouteInventory();
  const browser = await chromium.launch({ headless: true });
  const matrix = [];
  const roleFilter = process.env.ROLES
    ? new Set(process.env.ROLES.split(',').map((s) => s.trim()).filter(Boolean))
    : null;
  const activeRoles = roleFilter ? roles.filter((r) => roleFilter.has(r.role)) : roles;
  const routeLimit = Number(process.env.ROUTE_LIMIT || '0');

  for (const role of activeRoles) {
    const roleDir = path.join(OUT_DIR, role.role);
    await ensureDir(roleDir);

    const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
    const page = await context.newPage();

    try {
      await login(page, role);
      const sidebarRoutes = await collectSidebarRoutes(page, role.dashboard);
      let routes = [...new Set([...inventory, ...sidebarRoutes, role.dashboard])];
      if (routeLimit > 0) routes = routes.slice(0, routeLimit);

      for (const route of routes) {
        const url = `${BASE_URL}${route}`;
        let finalUrl = '';
        let status = 'ok';
        const file = path.join(roleDir, `${slugifyRoute(route)}.png`);

        try {
          await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 20000 });
          await page.waitForTimeout(250);
          finalUrl = page.url();
          const finalPath = new URL(finalUrl).pathname;
          const body = await page.locator('body').innerText().catch(() => '');
          status = classify(body, finalPath);
          if (status === 'ok' && normalizePath(finalPath) !== normalizePath(route)) {
            status = `redirected:${normalizePath(finalPath)}`;
          }
          await page.screenshot({ path: file, fullPage: false });
        } catch (err) {
          status = 'navigation_error';
          finalUrl = page.url();
          await page.screenshot({ path: file, fullPage: false }).catch(() => {});
        }

        matrix.push({
          role: role.role,
          email: role.email,
          requestedRoute: route,
          finalUrl,
          status,
          screenshot: file,
        });
      }
    } catch (err) {
      const failureFile = path.join(roleDir, `login_failure.png`);
      await page.screenshot({ path: failureFile, fullPage: true }).catch(() => {});
      matrix.push({
        role: role.role,
        email: role.email,
        requestedRoute: role.dashboard,
        finalUrl: page.url(),
        status: `login_error`,
        screenshot: failureFile,
      });
    }

    await context.close();
  }

  await browser.close();

  const jsonPath = '/home/abhishek/Desktop/flyfleet/artifacts/role_page_access_matrix.json';
  const csvPath = '/home/abhishek/Desktop/flyfleet/artifacts/role_page_access_matrix.csv';
  const htmlPath = '/home/abhishek/Desktop/flyfleet/artifacts/role_page_access_matrix.html';

  await fs.writeFile(jsonPath, JSON.stringify(matrix, null, 2));
  await writeCsv(csvPath, matrix);
  await writeHtmlReport(htmlPath, matrix);

  const summary = matrix.reduce((acc, r) => {
    acc[r.role] = acc[r.role] || { total: 0, ok: 0, nonOk: 0 };
    acc[r.role].total += 1;
    if (r.status === 'ok') acc[r.role].ok += 1;
    else acc[r.role].nonOk += 1;
    return acc;
  }, {});

  console.log(JSON.stringify({
    totalEntries: matrix.length,
    totalRoutesTestedPerRole: inventory.length,
    output: { jsonPath, csvPath, htmlPath, screenshotsDir: OUT_DIR },
    summary,
  }, null, 2));
}

run().catch((err) => {
  console.error(err);
  process.exit(1);
});

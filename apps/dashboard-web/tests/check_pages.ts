import { chromium } from "playwright-core";

(async () => {
  const browser = await chromium.launch({
    executablePath: "/usr/bin/google-chrome",
    headless: true,
  });
  const context = await browser.newContext();
  const page = await context.newPage();

  console.log("Setting up tokens...");
  await page.goto("http://localhost:3000/login");
  await page.evaluate(() => {
    localStorage.setItem("access_token", "test-token");
    localStorage.setItem(
      "user",
      JSON.stringify({ id: "admin-1", role: "GLOBAL_ADMIN" }),
    );
  });

  const pages = [
    { url: "/admin/dashboard", name: "dashboard" },
    { url: "/admin/control-tower", name: "control_tower" },
    { url: "/admin/fleet", name: "fleet" },
    { url: "/admin/drivers", name: "drivers" },
    { url: "/admin/analytics", name: "analytics" },
  ];

  for (const p of pages) {
    console.log(`Checking ${p.name} at ${p.url}...`);
    try {
      await page.goto(`http://localhost:3000${p.url}`, {
        waitUntil: "domcontentloaded",
        timeout: 15000,
      });
      // Wait a moment for rendering and initial data fetches
      await page.waitForTimeout(3000);
      const path = `/home/abhishek/.gemini/antigravity/brain/25e249e9-ea62-4782-b18f-aa0b2dc63765/page_${p.name}.png`;
      await page.screenshot({ path, fullPage: true });
      console.log(`Saved screenshot for ${p.name} to ${path}`);
    } catch (e: any) {
      console.error(`Failed to load ${p.name}:`, e.message);
    }
  }

  await browser.close();
})().catch((e) => {
  console.error("Test failed:", e);
  process.exit(1);
});

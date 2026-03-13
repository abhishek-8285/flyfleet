import { chromium } from "playwright-core";

(async () => {
  console.log("Connecting to local chrome...");
  const browser = await chromium.launch({
    executablePath: "/usr/bin/google-chrome",
    headless: true,
  });
  const context = await browser.newContext();
  const page = await context.newPage();
  await page.route("**/*.{woff,woff2,ttf}", (route) => route.abort());
  await page.route("**/fonts/**", (route) => route.abort());
  await page.route("**/*.pbf", (route) => {
    if (route.request().url().includes("fonts")) route.abort();
    else route.continue();
  });

  console.log("Setting up tokens...");
  await page.goto("http://localhost:3000/login", {
    waitUntil: "domcontentloaded",
  });
  await page.evaluate(() => {
    localStorage.setItem("access_token", "test-token");
    localStorage.setItem(
      "user",
      JSON.stringify({ id: "admin", name: "Admin", role: "admin" }),
    );
  });

  console.log("Opening Control Tower...");
  page.on("console", (msg) => console.log(`[BROWSER]: ${msg.text()}`));
  page.on("pageerror", (error) =>
    console.log(`[BROWSER_ERROR]: ${error.message}`),
  );

  await page.goto("http://localhost:3000/admin/control-tower", {
    waitUntil: "domcontentloaded",
  });

  await page.waitForTimeout(5000);

  console.log("Clicking Start Sim...");
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll("button"));
    const startBtn = buttons.find(
      (b) => b.textContent && b.textContent.toUpperCase().includes("START SIM"),
    );
    if (startBtn) {
      startBtn.click();
      console.log("Start Sim button clicked.");
    } else {
      console.log("Start Sim button NOT found.");
    }
  });

  console.log("Waiting 20s for metrics...");
  await page.waitForTimeout(20000);

  const path =
    "/home/abhishek/.gemini/antigravity/brain/25e249e9-ea62-4782-b18f-aa0b2dc63765/final_hyperscale_playwright.png";
  await page.screenshot({ path, timeout: 5000 });
  console.log("Screenshot saved to: " + path);

  const metricsObj = await page.evaluate(() => {
    const getText = (label: string) => {
      const el = Array.from(document.querySelectorAll("p")).find((p) =>
        p.textContent?.toUpperCase().includes(label),
      );
      if (el && el.nextElementSibling) {
        return el.nextElementSibling.textContent?.trim() || "EMPTY";
      }
      return "N/A";
    };
    return {
      VRAM: getText("VRAM"),
      FLEET: getText("ACTIVE FLEET"),
      H3: getText("ACTIVE H3"),
      FPS: getText("FPS"),
      LAG: getText("LAG"),
    };
  });
  console.log("EXTRACTED METRICS:", JSON.stringify(metricsObj, null, 2));

  await browser.close();
})().catch((e) => {
  console.error(e);
  process.exit(1);
});

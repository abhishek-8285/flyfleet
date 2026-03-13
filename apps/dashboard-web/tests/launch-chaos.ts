import { chromium } from "playwright-core";

(async () => {
  const browser = await chromium.launch({
    executablePath: "/usr/bin/google-chrome",
    headless: true,
  });
  const context = await browser.newContext();
  const page = await context.newPage();

  page.on("console", (msg) => {
    // Only print relevant worker manager logs to verify recovery
    const text = msg.text();
    if (text.includes("[WorkerManager]") || text.includes("crashing")) {
      console.log(`[BROWSER]: ${text}`);
    }
  });

  console.log("=== SCENARIO 3: Worker Deadlock Simulation ===");
  console.log("Loading Control Tower...");

  await page.goto("http://localhost:3000/login");
  await page.evaluate(() => {
    localStorage.setItem("access_token", "test-token");
    localStorage.setItem(
      "user",
      JSON.stringify({ id: "admin-1", role: "GLOBAL_ADMIN" }),
    );
  });
  await page.goto("http://localhost:3000/admin/control-tower", {
    waitUntil: "domcontentloaded",
  });

  await page.waitForTimeout(5000);

  console.log("Clicking 'Start Sim'...");
  await page.evaluate(() => {
    const btn = Array.from(document.querySelectorAll("button")).find((b) =>
      b.textContent?.toUpperCase().includes("START SIM"),
    );
    btn?.click();
  });

  await page.waitForTimeout(5000);

  console.log("Injecting Fatal Fault: Terminating Decoder/Telemetry Worker...");
  await page.evaluate(() => {
    const wnd = window as any;
    if (wnd.__DEBUG_WORKER_MANAGERS__) {
      const mgr = wnd.__DEBUG_WORKER_MANAGERS__.telemetry;
      console.log(
        "Triggering explicit handleCrash() to simulate unrecoverable worker exception.",
      );
      // Invoke the private handleCrash method which triggers recovery logic
      mgr.handleCrash();
    }
  });

  console.log("Waiting 5 seconds to verify autonomous auto-recovery...");
  await page.waitForTimeout(5000);

  const finalMetrics = await page.evaluate(() => {
    const getText = (label: string) => {
      const el = Array.from(document.querySelectorAll("p")).find((p) =>
        p.textContent?.toUpperCase().includes(label),
      );
      return el?.nextElementSibling?.textContent?.trim() || "N/A";
    };
    return {
      FLEET: getText("ACTIVE FLEET"),
      FPS: getText("RENDER FPS"),
    };
  });

  console.log(
    "Post-Crash Recovery Metrics:",
    JSON.stringify(finalMetrics, null, 2),
  );

  const path =
    "/home/abhishek/.gemini/antigravity/brain/25e249e9-ea62-4782-b18f-aa0b2dc63765/chaos_worker_deadlock.png";
  await page.screenshot({ path, timeout: 5000 });
  console.log(`Screenshot saved to: ${path}`);

  await browser.close();
})().catch((e) => {
  console.error("Test failed:", e);
  process.exit(1);
});

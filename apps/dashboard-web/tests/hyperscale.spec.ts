import { test } from "@playwright/test";

test("Hyperscale Verification", async ({ page }) => {
  test.setTimeout(60000); // 60s timeout for the whole test

  // Go to a page on the domain to set localStorage
  await page.goto("http://localhost:3000/login");

  await page.evaluate(() => {
    localStorage.setItem("access_token", "test-token");
    localStorage.setItem(
      "user",
      JSON.stringify({ id: "admin", name: "Admin", role: "admin" }),
    );
  });

  await page.goto("http://localhost:3000/admin/control-tower");

  console.log("Arrived at Control Tower, waiting 5 seconds for WebGL init...");
  await page.waitForTimeout(5000);

  // Click Start Sim
  const startBtn = page.locator("button", { hasText: /START SIM/i });
  if ((await startBtn.count()) > 0) {
    await startBtn.first().click();
    console.log("Successfully clicked 'Start Sim' button.");
  } else {
    console.log("ERROR: 'Start Sim' button not found.");
  }

  console.log(
    "Waiting 20 seconds for the telemetry storm to saturate the workers...",
  );
  await page.waitForTimeout(20000);

  const path =
    "/home/abhishek/.gemini/antigravity/brain/25e249e9-ea62-4782-b18f-aa0b2dc63765/final_hyperscale_playwright.png";
  await page.screenshot({ path, fullPage: true });
  console.log("Screenshot successfully saved to:", path);

  // Try to extract metrics
  const text = await page.locator("body").innerText();
  console.log("--- PAGE TEXT EXTRACT ---");
  const lines = text.split("\n");
  lines.forEach((line, i) => {
    const uLine = line.toUpperCase();
    if (
      uLine.includes("VRAM") ||
      uLine.includes("FPS") ||
      uLine.includes("LAG") ||
      uLine.includes("ACTIVE FLEET")
    ) {
      console.log(`[Metric Match]: ${line}`);
      if (lines[i + 1]) console.log(`  Value Follows -> ${lines[i + 1]}`);
    }
  });
});

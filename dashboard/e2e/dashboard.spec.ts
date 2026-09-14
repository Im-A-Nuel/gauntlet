import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

test("recorded reports, inspection, filters, comparisons and mobile layout", async ({
  page,
}) => {
  test.setTimeout(90000);
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("/");
  await expect(page.getByText("Recorded demo", { exact: true })).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Verification by file" }),
  ).toBeVisible();
  const runOptions = await page
    .getByRole("combobox", { name: "Run", exact: true })
    .locator("option")
    .allTextContents();
  expect(runOptions.length).toBeGreaterThanOrEqual(2);
  await page.screenshot({
    path: "../artifacts/qa/overview-desktop.png",
    fullPage: true,
  });
  const axe = await new AxeBuilder({ page })
    .withTags(["wcag2a", "wcag2aa", "wcag21aa"])
    .analyze();
  expect(
    axe.violations.map((v) => ({
      id: v.id,
      nodes: v.nodes.map((n) => n.target),
    })),
  ).toEqual([]);
  await page.getByRole("button", { name: "Copy command", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Command copied", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("navigation")
    .getByRole("link", { name: "Kill matrix", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Every mutation. Every outcome." }),
  ).toBeVisible();
  const cells = page.getByRole("button", { name: /mutant .* in .* line/ });
  expect(await cells.count()).toBeGreaterThan(0);
  await cells.first().click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).not.toBeVisible();
  await page.getByLabel("Filter files").fill("nonexistent-path");
  await expect(page.getByText(/No files match/)).toBeVisible();
  await page.getByLabel("Filter files").fill("");
  await page
    .getByRole("navigation")
    .getByRole("link", { name: /Survivors/ })
    .click();
  await page.getByLabel("Search mutations").fill("nonexistent-path");
  await expect(
    page.getByRole("heading", { name: "No mutations match this search." }),
  ).toBeVisible();
  await page.getByLabel("Search mutations").fill("");
  await page
    .getByRole("combobox", { name: "Outcome", exact: true })
    .selectOption("killed");
  await expect(page.locator(".survivor-row").first()).toBeVisible();
  await page
    .getByRole("navigation")
    .getByRole("link", { name: "Compare runs", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Did the tests get stronger?" }),
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "What changed by file" }),
  ).toBeVisible();
  await page.screenshot({
    path: "../artifacts/qa/compare-desktop.png",
    fullPage: true,
  });
  const after = await page.getByLabel("Follow-up run").inputValue();
  await page.getByLabel("Baseline run").selectOption(after);
  await expect(page.getByText(/The same run is selected twice/)).toBeVisible();
  for (const width of [375, 768, 1024, 1440]) {
    await page.setViewportSize({ width, height: 900 });
    for (const route of ["/", "/matrix", "/survivors", "/compare"]) {
      await page.goto(route);
      await expect(page.locator("main h1")).toBeVisible();
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= window.innerWidth,
        ),
        `${route} overflow at ${width}`,
      ).toBe(true);
    }
  }
  await page.setViewportSize({ width: 375, height: 900 });
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/");
  await expect(page.getByRole("main")).toBeVisible();
  await page.screenshot({
    path: "../artifacts/qa/overview-mobile.png",
    fullPage: true,
  });
  expect(errors).toEqual([]);
});

test("API validates IDs and returns real comparison values", async ({
  request,
}) => {
  const list = await request.get("/api/runs");
  expect(list.ok()).toBe(true);
  expect(list.headers()["x-gauntlet-source"]).toBe("sample");
  const runs = await list.json();
  expect(runs.length).toBeGreaterThanOrEqual(2);
  const detail = await request.get(`/api/runs/${runs[0].runId}`);
  expect(detail.ok()).toBe(true);
  expect((await detail.json()).totals).toEqual(runs[0].totals);
  expect((await request.get("/api/runs/not-a-run")).status()).toBe(404);
  expect((await request.get("/api/compare")).status()).toBe(400);
  const result = await request.get(
    `/api/compare?a=${runs[1].runId}&b=${runs[0].runId}`,
  );
  expect(result.ok()).toBe(true);
  expect((await result.json()).perFile.length).toBeGreaterThan(0);
});

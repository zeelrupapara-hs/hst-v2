import { test, expect } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";

const LOGIN = process.env.PW_LOGIN || "1000";
const PASSWORD = process.env.PW_PASSWORD || "Bootstrap-Admin-2026!";
const OUT_DIR = path.join(import.meta.dirname, "../.playwright-audit");

const TAB_NAMES = [
  "Common",
  "Currency",
  "Quotes",
  "Trade",
  "Execution",
  "Margin",
  "Margin Rates",
  "Swaps",
  "Sessions",
];

const VIEWPORTS = [
  { label: "1280x800", width: 1280, height: 800 },
  { label: "1024x768", width: 1024, height: 768 },
];

async function loginAdmin(page) {
  await page.goto("/login");
  await page.locator("#login").fill(LOGIN);
  await page.locator("#password").fill(PASSWORD);
  await page.locator("#panel").selectOption("32");
  await page.getByRole("button", { name: "Login" }).click();
  await page.waitForURL(/\/admin/);
}

/** Collect horizontal overflow and out-of-dialog clipping inside the active config panel. */
async function auditPanel(page, tabName, viewportLabel) {
  return page.evaluate(
    ({ tab, viewport }) => {
      const issues = [];
      const dialog = document.querySelector(".sym-config-window, .settings-dialog.sym-config-window");
      const panel = document.querySelector(".config-panel.active");
      if (!dialog || !panel) {
        issues.push({
          viewport,
          tab,
          kind: "missing",
          detail: "dialog or active panel not found",
        });
        return issues;
      }

      const dialogRect = dialog.getBoundingClientRect();
      const body = dialog.querySelector(".config-body");
      const bodyRect = body?.getBoundingClientRect();

      if (body && body.scrollWidth > body.clientWidth + 1) {
        issues.push({
          viewport,
          tab,
          kind: "body-overflow-x",
          detail: `config-body scrollWidth ${body.scrollWidth} > clientWidth ${body.clientWidth}`,
        });
      }

      if (dialog.scrollWidth > dialog.clientWidth + 1) {
        issues.push({
          viewport,
          tab,
          kind: "dialog-overflow-x",
          detail: `dialog scrollWidth ${dialog.scrollWidth} > clientWidth ${dialog.clientWidth}`,
        });
      }

      if (dialogRect.right > window.innerWidth + 1) {
        issues.push({
          viewport,
          tab,
          kind: "viewport-clip-right",
          detail: `dialog right ${Math.round(dialogRect.right)} > viewport ${window.innerWidth}`,
        });
      }

      const selectors = [
        ".form-grid",
        ".sym-form-two-col",
        ".sym-swaps-row",
        ".sym-sessions-main",
        ".sym-sessions-table-wrap",
        ".prop-select",
        ".prop-select-box",
        ".sym-swaps-presets",
        ".config-tabs",
      ];

      for (const sel of selectors) {
        for (const el of panel.querySelectorAll(sel)) {
          const r = el.getBoundingClientRect();
          if (r.width === 0 && r.height === 0) continue;

          if (el.scrollWidth > el.clientWidth + 2) {
            issues.push({
              viewport,
              tab,
              kind: "overflow-x",
              detail: `${sel} scroll ${el.scrollWidth} > client ${el.clientWidth}`,
              cls: el.className,
            });
          }

          const right = r.right;
          const limit = dialogRect.right - 4;
          if (right > limit + 1) {
            issues.push({
              viewport,
              tab,
              kind: "clip-right",
              detail: `${sel} right ${Math.round(right)} > dialog ${Math.round(limit)}`,
              cls: el.className,
            });
          }

          if (bodyRect && r.left < bodyRect.left - 1) {
            issues.push({
              viewport,
              tab,
              kind: "clip-left",
              detail: `${sel} left ${Math.round(r.left)} < body ${Math.round(bodyRect.left)}`,
              cls: el.className,
            });
          }
        }
      }

      if (tab === "Swaps") {
        const days = panel.querySelector(".sym-swaps-year-row .prop-select");
        const holidays = panel.querySelector(".sym-swaps-holidays-check");
        if (days && holidays) {
          const dTop = days.getBoundingClientRect().top;
          const hTop = holidays.getBoundingClientRect().top;
          if (Math.abs(dTop - hTop) > 8) {
            issues.push({
              viewport,
              tab,
              kind: "row-mismatch",
              detail: `Days dropdown top ${Math.round(dTop)} vs holidays top ${Math.round(hTop)}`,
            });
          }
        }

        const inputs = panel.querySelectorAll(".sym-swaps-positions input");
        if (inputs.length >= 2) {
          const w0 = inputs[0].getBoundingClientRect().width;
          const w1 = inputs[1].getBoundingClientRect().width;
          if (Math.abs(w0 - w1) > 2) {
            issues.push({
              viewport,
              tab,
              kind: "width-mismatch",
              detail: `Long ${Math.round(w0)}px vs Short ${Math.round(w1)}px`,
            });
          }
        }
      }

      if (tab === "Trade") {
        const volInputs = panel.querySelectorAll(".sym-volumes-grid input");
        if (volInputs.length >= 2) {
          const widths = [...volInputs].map((el) => el.getBoundingClientRect().width);
          const maxDiff = Math.max(...widths) - Math.min(...widths);
          if (maxDiff > 4) {
            issues.push({
              viewport,
              tab,
              kind: "width-mismatch",
              detail: `Volume inputs widths ${widths.map((w) => Math.round(w)).join(", ")}`,
            });
          }
        }
      }

      return issues;
    },
    { tab: tabName, viewport: viewportLabel },
  );
}

/** Audit SessionEditorDialog opened from the Sessions tab. */
async function auditSessionEditor(page, viewportLabel) {
  return page.evaluate((viewport) => {
    const issues = [];
    const dialog = document.querySelector(".sym-session-dialog");
    if (!dialog) {
      issues.push({
        viewport,
        tab: "SessionEditor",
        kind: "missing",
        detail: "session editor dialog not found",
      });
      return issues;
    }

    const dialogRect = dialog.getBoundingClientRect();
    const body = dialog.querySelector(".sym-session-dialog-body");
    const bodyRect = body?.getBoundingClientRect();

    if (dialogRect.right > window.innerWidth + 1) {
      issues.push({
        viewport,
        tab: "SessionEditor",
        kind: "viewport-clip-right",
        detail: `dialog right ${Math.round(dialogRect.right)} > viewport ${window.innerWidth}`,
      });
    }

    if (dialogRect.bottom > window.innerHeight + 1) {
      issues.push({
        viewport,
        tab: "SessionEditor",
        kind: "viewport-clip-bottom",
        detail: `dialog bottom ${Math.round(dialogRect.bottom)} > viewport ${window.innerHeight}`,
      });
    }

    if (body && body.scrollWidth > body.clientWidth + 1) {
      issues.push({
        viewport,
        tab: "SessionEditor",
        kind: "body-overflow-x",
        detail: `session body scrollWidth ${body.scrollWidth} > clientWidth ${body.clientWidth}`,
      });
    }

    if (dialog.scrollWidth > dialog.clientWidth + 1) {
      issues.push({
        viewport,
        tab: "SessionEditor",
        kind: "dialog-overflow-x",
        detail: `dialog scrollWidth ${dialog.scrollWidth} > clientWidth ${dialog.clientWidth}`,
      });
    }

    const selectors = [
      ".sym-timeline-frame",
      ".sym-timeline-row",
      ".sym-sessions-intro-compact",
      ".sym-session-separate-inline",
    ];

    for (const sel of selectors) {
      for (const el of dialog.querySelectorAll(sel)) {
        const r = el.getBoundingClientRect();
        if (r.width === 0 && r.height === 0) continue;

        if (el.scrollWidth > el.clientWidth + 2) {
          issues.push({
            viewport,
            tab: "SessionEditor",
            kind: "overflow-x",
            detail: `${sel} scroll ${el.scrollWidth} > client ${el.clientWidth}`,
            cls: el.className,
          });
        }

        const limit = dialogRect.right - 4;
        if (r.right > limit + 1) {
          issues.push({
            viewport,
            tab: "SessionEditor",
            kind: "clip-right",
            detail: `${sel} right ${Math.round(r.right)} > dialog ${Math.round(limit)}`,
            cls: el.className,
          });
        }

        if (bodyRect && r.left < bodyRect.left - 1) {
          issues.push({
            viewport,
            tab: "SessionEditor",
            kind: "clip-left",
            detail: `${sel} left ${Math.round(r.left)} < body ${Math.round(bodyRect.left)}`,
            cls: el.className,
          });
        }
      }
    }

    return issues;
  }, viewportLabel);
}

async function openSymbolDialog(page) {
  await loginAdmin(page);
  await page.goto("/admin/symbols");
  await page.waitForSelector(".data-table tbody tr");
  await page.locator(".data-table tbody tr").first().dblclick();
  await page.waitForSelector(".sym-config-window");
}

async function auditSessionEditorFlow(page, viewportLabel, outSubdir) {
  await page.getByRole("button", { name: "Sessions", exact: true }).click();
  await page.waitForTimeout(150);

  const wednesdayRow = page.locator(".sym-sessions-main .sym-sessions-table tbody tr").filter({
    hasText: "Wednesday",
  });
  await wednesdayRow.dblclick();
  await page.waitForSelector(".sym-session-dialog");

  const shot = path.join(outSubdir, "session-editor-wednesday.png");
  await page.locator(".sym-session-dialog").screenshot({ path: shot });

  const issues = await auditSessionEditor(page, viewportLabel);

  await page.locator(".sym-session-cancel").click();
  await page.waitForSelector(".sym-session-dialog", { state: "detached" });

  return issues;
}

async function runFullAudit(page, viewport) {
  const outSubdir = path.join(OUT_DIR, viewport.label);
  fs.mkdirSync(outSubdir, { recursive: true });

  await page.setViewportSize({ width: viewport.width, height: viewport.height });
  await openSymbolDialog(page);

  const allIssues = [];

  for (const tab of TAB_NAMES) {
    await page.getByRole("button", { name: tab, exact: true }).click();
    await page.waitForTimeout(150);

    if (tab === "Swaps") {
      const enable = page.getByRole("checkbox", { name: "Enable swaps" });
      if (!(await enable.isChecked())) await enable.click();
    }

    const shot = path.join(outSubdir, `tab-${tab.toLowerCase().replace(/\s+/g, "-")}.png`);
    await page.locator(".sym-config-window").screenshot({ path: shot });

    allIssues.push(...(await auditPanel(page, tab, viewport.label)));

    if (tab === "Sessions") {
      allIssues.push(...(await auditSessionEditorFlow(page, viewport.label, outSubdir)));
    }
  }

  return allIssues;
}

test.describe("Symbol settings UI audit", () => {
  test.beforeAll(() => {
    fs.mkdirSync(OUT_DIR, { recursive: true });
  });

  for (const viewport of VIEWPORTS) {
    test(`admin symbols dialog — ${viewport.label} — all tabs and session editor`, async ({ page }) => {
      const allIssues = await runFullAudit(page, viewport);

      const issuesFile = path.join(OUT_DIR, `issues-${viewport.label}.json`);
      fs.writeFileSync(issuesFile, JSON.stringify(allIssues, null, 2));

      if (allIssues.length) {
        console.log(`\n=== UI ISSUES (${viewport.label}) ===`);
        for (const i of allIssues) console.log(JSON.stringify(i));
      }

      expect(allIssues, `UI issues written to ${issuesFile}`).toEqual([]);
    });
  }
});

// Real Chromium acceptance test for the entirely local OpenCMMC workbench.
// Synthetic assessment only. Never print private access URLs or cookies.
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import http from "node:http";
import { readFile, mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { chromium } from "playwright";

function startWorkbench(port, directory) {
    return new Promise((resolve, reject) => {
        const app = spawn(process.env.OPENCMMC_BIN || "./opencmmc", [
            "cmmc", "ui", "--data-dir", directory, "--port", String(port)
        ], { stdio: ["ignore", "pipe", "pipe"] });
        let output = "";
        let stderr = "";
        let settled = false;
        const timeout = setTimeout(() => {
            fail(new Error("OpenCMMC local UI startup timed out"));
        }, 20000);
        function fail(error) {
            if (settled) return;
            settled = true;
            clearTimeout(timeout);
            app.kill("SIGTERM");
            reject(error);
        }
        app.once("error", fail);
        app.once("exit", code => {
            if (!settled) fail(new Error("OpenCMMC stopped before startup, code " + code + ": " + stderr.slice(-300)));
        });
        app.stderr.on("data", part => { stderr = (stderr + String(part)).slice(-1000); });
        app.stdout.on("data", part => {
            if (settled) return;
            output += String(part);
            const pattern = new RegExp("http://127\\.0\\.0\\.1:" + port + "/\\?access=[0-9a-f]{64}");
            const matched = output.match(pattern);
            if (matched) {
                settled = true;
                clearTimeout(timeout);
                resolve({app, privateUrl: matched[0]});
            }
        });
    });
}

async function listenProxy(port) {
    // ChromeOS can preserve an IPv6 local browser Origin while forwarding
    // requests to an IPv4 loopback listener with a rewritten Host header.
    const proxy = http.createServer((incoming, outgoing) => {
        const upstream = http.request({
            hostname: "127.0.0.1",
            port,
            method: incoming.method,
            path: incoming.url,
            headers: { ...incoming.headers, host: "127.0.0.1:" + port }
        }, response => {
            outgoing.writeHead(response.statusCode || 502, response.headers);
            response.pipe(outgoing);
        });
        upstream.on("error", () => {
            if (!outgoing.headersSent) outgoing.writeHead(502);
            outgoing.end();
        });
        incoming.pipe(upstream);
    });
    await new Promise((resolve, reject) => {
        proxy.once("error", reject);
        proxy.listen({host: "::1", port, ipv6Only: true}, resolve);
    });
    return proxy;
}

async function createAssessment(page, level, scope) {
    await page.getByRole("heading", {name: "Create an offline assessment"}).waitFor();
    const form = page.locator('form[action="/create"]');
    await form.locator('select[name="level"]').selectOption(level);
    await form.locator('input[name="scope"]').fill(scope);
    await form.getByRole("button", {name: "Create assessment"}).click();
    await page.waitForLoadState("domcontentloaded");
    if (await page.getByRole("heading", {name: /^Requirement /}).count() === 0) {
        // CI uses synthetic data only. Do not print cookies, private URLs,
        // form values, or raw request headers.
        const safeText = (await page.locator("body").innerText())
            .replace(/access=[^\\s&]*/g, "access=[REDACTED]")
            .slice(0, 800);
        throw new Error("Local assessment form did not open workspace. Safe synthetic page: " + safeText);
    }
}

async function recordGap(page) {
    const form = page.locator('form[action="/decide"]').first();
    await form.locator('select[name="decision"]').selectOption("GAP");
    await form.locator('input[name="reviewer"]').fill("Synthetic QA reviewer");
    await form.locator('textarea[name="rationale"]').fill("Browser acceptance test, not an assertion of compliance");
    await form.getByRole("button", {name: "Save requirement review"}).click();
}

async function exportJSON(page) {
    const awaiting = page.waitForEvent("download");
    await page.getByRole("link", {name: "Export local JSON findings"}).click();
    const download = await awaiting;
    const localPath = await download.path();
    const raw = await readFile(localPath, "utf8");
    assert.ok(raw.length > 100, "readiness download is empty");
    const parsed = JSON.parse(raw);
    assert.equal(typeof parsed, "object");
    return parsed;
}

const root = await mkdtemp(join(tmpdir(), "opencmmc-chromium-"));
const processes = [];
let browser;
let proxy;
try {
    const port = 18765;
    const first = await startWorkbench(port, join(root, "level1"));
    processes.push(first.app);
    browser = await chromium.launch({headless: true});
    const context = await browser.newContext({acceptDownloads: true});
    const page = await context.newPage();
    await page.goto(first.privateUrl, {waitUntil: "domcontentloaded"});
    await page.getByRole("heading", {name: "Create an offline assessment"}).waitFor();
    assert.ok(!page.url().includes("access="), "private access code was retained in the browser URL");

    // Cross-origin forged form attempts must fail before writing any state.
    const forged = await context.request.post("http://127.0.0.1:" + port + "/create", {
        headers: {
            "Origin": "https://untrusted.invalid",
            "Sec-Fetch-Site": "cross-site",
            "Content-Type": "application/x-www-form-urlencoded"
        },
        data: "scope=forged&level=1&csrf=bogus",
        failOnStatusCode: false
    });
    assert.equal(forged.status(), 403, "forged cross-site POST was not forbidden");
    await page.reload();
    await page.getByRole("heading", {name: "Create an offline assessment"}).waitFor();

    await createAssessment(page, "1", "Synthetic Chromium Level 1");
    await recordGap(page);
    await page.getByText(/Gaps: 1/).waitFor();
    const evidenceForm = page.locator('form[action="/evidence"]').first();
    await evidenceForm.locator('input[name="file"]').setInputFiles({
        name: "synthetic-proof.txt",
        mimeType: "text/plain",
        buffer: Buffer.from("synthetic browser regression evidence only\n")
    });
    await evidenceForm.getByRole("button", {name: "Attach evidence locally"}).click();
    await page.getByText(/1 local evidence reference/).waitFor();
    await exportJSON(page);

    proxy = await listenProxy(port);
    const forwarded = await browser.newContext({acceptDownloads: true});
    const forwardedPage = await forwarded.newPage();
    await forwardedPage.goto(first.privateUrl.replace("127.0.0.1", "[::1]"), {
        waitUntil: "domcontentloaded"
    });
    await forwardedPage.getByRole("heading", {name: /^Requirement /}).waitFor();
    await forwardedPage.locator(".control").nth(1).click();
    await recordGap(forwardedPage);
    await forwardedPage.getByText(/Gaps: 2/).waitFor();
    await exportJSON(forwardedPage);

    const second = await startWorkbench(port + 1, join(root, "level2"));
    processes.push(second.app);
    const level2 = await browser.newContext({acceptDownloads: true});
    const level2Page = await level2.newPage();
    await level2Page.goto(second.privateUrl, {waitUntil: "domcontentloaded"});
    await createAssessment(level2Page, "2", "Synthetic Chromium Level 2");
    await exportJSON(level2Page);
    console.log("PASS: Chromium local sessions, cross-site rejection, Level 1 review/evidence/export, forwarded IPv6 origin and Level 2 export");
} finally {
    if (browser) await browser.close();
    if (proxy) await new Promise(resolve => proxy.close(resolve));
    for (const app of processes) app.kill("SIGTERM");
    await rm(root, {recursive: true, force: true});
}

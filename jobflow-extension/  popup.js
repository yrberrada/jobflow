// jobflow-extension/popup.js

document.addEventListener("DOMContentLoaded", () => {
    const btn = document.getElementById("send-job");
    const statusEl = document.getElementById("status");
    const stageEl = document.getElementById("stage");
    const notesEl = document.getElementById("quick-notes");

    if (!btn) {
        console.error("[JobFlow] Button #send-job not found in popup.html");
        if (statusEl) {
            statusEl.textContent = "Popup error: button not found.";
        }
        return;
    }

    btn.addEventListener("click", async () => {
        if (statusEl) statusEl.textContent = "Collecting job info…";

        try {
            const [tab] = await chrome.tabs.query({
                active: true,
                lastFocusedWindow: true,
            });

            if (!tab || !tab.id) {
                throw new Error("No active tab found.");
            }

            const stage = stageEl ? stageEl.value : "Applied";
            const quickNotes = notesEl ? notesEl.value.trim() : "";

            const [{ result }] = await chrome.scripting.executeScript({
                target: { tabId: tab.id },
                func: scrapeJobFromPage,
                args: [stage, quickNotes],
            });

            console.log("[JobFlow] Scraped job:", result);

            if (!result) {
                throw new Error(
                    "Scraper returned nothing – are you on a LinkedIn job page?"
                );
            }

            if (statusEl) statusEl.textContent = "Sending to JobFlow…";

            const resp = await fetch("http://localhost:8081/apply", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(result),
            });

            if (!resp.ok) {
                const txt = await resp.text();
                console.error("[JobFlow] JobFlow API error:", txt);
                if (statusEl) {
                    statusEl.textContent =
                        "JobFlow API error: " + txt.slice(0, 140) + (txt.length > 140 ? "…" : "");
                }
                return;
            }

            const data = await resp.json();
            console.log("[JobFlow] JobFlow response:", data);
            if (statusEl) statusEl.textContent = "Sent ✅";
        } catch (err) {
            console.error("[JobFlow] unexpected error:", err);
            if (statusEl) {
                statusEl.textContent = "Error: " + (err.message || String(err));
            }
        }
    });
});

// Runs in the LinkedIn tab context.
function scrapeJobFromPage(stage, quickNotes) {
    function pickText(selectors) {
        for (const sel of selectors) {
            const el = document.querySelector(sel);
            if (el && el.textContent) {
                return el.textContent.trim();
            }
        }
        return "";
    }

    // Title
    const title = pickText([
        "h1.jobs-unified-top-card__job-title",
        "h1.top-card-layout__title",
        "h1",
    ]);

    // Company
    const company = pickText([
        "a.jobs-unified-top-card__company-name",
        "a.topcard__org-name-link",
        "[data-test-company-name]",
        "a[href*='/company/']",
    ]);

    // Location + possible work mode mixed in
    let locationRaw = pickText([
        "[data-test-job-location]",
        ".jobs-unified-top-card__bullet",
        ".topcard__flavor--bullet",
        ".topcard__flavor",
    ]);

    // Description (fallback to body text)
    let description = pickText([
        ".jobs-description__content",
        ".jobs-box__html-content",
        "[data-test='job-description']",
        "article",
    ]);
    if (!description) {
        description = document.body.innerText.slice(0, 2000);
    }

    const url = window.location.href;

    // --- Derive work mode + clean location ---------------------------------

    let workMode = "";
    let location = locationRaw;

    if (location) {
        const modeCandidates = [
            { kw: "Remote", val: "Remote" },
            { kw: "Hybrid", val: "Hybrid" },
            { kw: "On-site", val: "On-site" },
            { kw: "On site", val: "On-site" },
            { kw: "On-site", val: "On-site" }, // weird hyphen char
        ];

        for (const { kw, val } of modeCandidates) {
            if (location.includes(kw)) {
                workMode = val;
                location = location
                    .replace(kw, "")
                    .replace(/[()·]/g, " ")
                    .replace(/\s+/g, " ")
                    .trim();
                break;
            }
        }
    }

    if (!workMode && description) {
        const lower = description.toLowerCase();
        if (lower.includes("remote")) workMode = "Remote";
        else if (lower.includes("hybrid")) workMode = "Hybrid";
        else if (lower.includes("on-site") || lower.includes("on site"))
            workMode = "On-site";
    }

    // --- Salary heuristic from description ---------------------------------

    let salary = "";
    if (description) {
        const salaryRegex =
            /(?:USD|CAD|CA\$|\$|€|£)\s?\d[\d,]*(?:\s?[–-]\s?\d[\d,]*)?.{0,40}?(?:hour|hr|day|week|month|year|annum|annually|per hour|per year)/i;
        const m = description.match(salaryRegex);
        if (m) {
            salary = m[0].replace(/\s+/g, " ").trim();
        }
    }

    const baseNotes = "Captured from LinkedIn";
    const notes = quickNotes
        ? baseNotes + "\n\n" + quickNotes
        : baseNotes;

    return {
        external_id: url,
        position: title,
        company: company,
        location: location,
        url: url,
        work_mode: workMode,
        salary: salary,
        description: description,
        notes: notes,
        stage: stage || "Applied",
        outcome: "Active",
        next_interview: "",
    };
}

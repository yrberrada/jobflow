// popup.js

document.addEventListener("DOMContentLoaded", () => {
    const btn = document.getElementById("send-job");
    const statusEl = document.getElementById("status");

    if (!btn || !statusEl) {
        console.error("[JobFlow] Button #send-job not found in popup.html");
        if (statusEl) {
            statusEl.textContent = "Popup error: button not found.";
        }
        return;
    }

    btn.addEventListener("click", async () => {
        statusEl.textContent = "Collecting job info…";

        try {
            // 1) Active tab
            const [tab] = await chrome.tabs.query({
                active: true,
                lastFocusedWindow: true,
            });

            if (!tab || !tab.id) {
                statusEl.textContent = "No active tab found.";
                return;
            }

            // 2) Run scraper in the LinkedIn tab
            const [injected] = await chrome.scripting.executeScript({
                target: { tabId: tab.id },
                func: scrapeJobFromPage,
            });

            const result = injected && injected.result;
            console.log("[JobFlow] Scraped job:", result);

            if (!result) {
                statusEl.textContent =
                    "Scraping failed (no result). Check page console.";
                return;
            }

            // If no title, warn but still send (we still get description & URL)
            if (!result.position) {
                statusEl.textContent =
                    "Warning: no title found, sending anyway…";
            } else {
                statusEl.textContent = "Sending to JobFlow…";
            }

            // 3) Send to local Go API
            const resp = await fetch("http://localhost:8081/apply", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(result),
            });

            if (!resp.ok) {
                const txt = await resp.text();
                console.error("[JobFlow] JobFlow API error:", txt);
                statusEl.textContent =
                    "JobFlow returned an error (check server logs).";
                return;
            }

            const data = await resp.json();
            console.log("[JobFlow] JobFlow response:", data);
            statusEl.textContent = "Sent ✅";
        } catch (err) {
            console.error("[JobFlow] Popup error:", err);
            statusEl.textContent =
                "Error: " + (err && err.message ? err.message : String(err));
        }
    });
});

// This function runs inside the LinkedIn tab context.
function scrapeJobFromPage() {
    function text(el) {
        return el ? el.innerText.trim() : "";
    }

    function pickText(selectors) {
        for (const sel of selectors) {
            const el = document.querySelector(sel);
            const val = text(el);
            console.log("[JobFlow][scrape] selector", sel, "=>", val);
            if (val) return val;
        }
        return "";
    }

    function getMeta(prop) {
        const el = document.querySelector(`meta[property="${prop}"]`);
        return el && el.content ? el.content.trim() : "";
    }

    function inferLocationFromDescription(desc) {
        const cityStateRegex = /\b([A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)*),\s*([A-Z]{2})\b/;
        const m = desc.match(cityStateRegex);
        if (m) {
            return `${m[1]}, ${m[2]}`;
        }
        return "";
    }

    function inferCompanyFromDescription(desc) {
        // e.g. "As a software developer at Epic, you’ll..."
        const reAt = /\bat\s+([A-Z][A-Za-z0-9&.\-]*(?:\s+[A-Z][A-Za-z0-9&.\-]*)?)/;
        const m1 = desc.match(reAt);
        if (m1) {
            return m1[1].trim();
        }

        // e.g. "Radixlink is a trusted partner..."
        const reIs = /\b([A-Z][A-Za-z0-9&.\-]*(?:\s+[A-Z][A-Za-z0-9&.\-]*)?)\s+(?:is|are)\s+(?:a|an)\s+/;
        const m2 = desc.match(reIs);
        if (m2) {
            return m2[1].trim();
        }

        return "";
    }

    function inferWorkModeFromDescription(desc) {
        const lower = desc.toLowerCase();
        if (lower.includes("remote")) return "Remote";
        if (lower.includes("hybrid")) return "Hybrid";
        if (lower.includes("on-site") || lower.includes("on site") || lower.includes("onsite"))
            return "On-site";

        // campus / relocation ⇒ strongly suggests on-site
        if (
            lower.includes("based on our campus") ||
            lower.includes("on our campus") ||
            lower.includes("relocation to") ||
            lower.includes("requires relocation") ||
            lower.includes("relocation assistance")
        ) {
            return "On-site";
        }

        return "";
    }

    function inferWorkModeFromPills() {
        const labels = ["On-site", "On site", "Remote", "Hybrid"];
        const els = Array.from(document.querySelectorAll("button, span, div"));
        for (const el of els) {
            const txt = el.innerText.trim();
            if (!txt) continue;
            if (labels.includes(txt)) {
                if (txt === "On site") return "On-site";
                return txt;
            }
        }
        return "";
    }

    const url = window.location.href;

    // --- Title ---

    let title = pickText([
        "h1.top-card-layout__title",
        "h1.jobs-unified-top-card__job-title",
        "h1[data-test-single-line-truncate]",
        "h1[data-test-job-title]",
        "h1"
    ]);

    // Fallback: og:title or document.title
    if (!title) {
        let source = "";
        let t = getMeta("og:title");
        if (t) {
            source = "og:title";
        } else if (document && document.title) {
            t = document.title;
            source = "document.title";
        }
        if (t) {
            // ignore plain "LinkedIn" / "Jobs | LinkedIn" junk
            if (!/linkedin/i.test(t) || /\-/.test(t)) {
                const main = t.replace(/\s+\|.*$/, "");
                const parts = main.split(" - ").map((p) => p.trim());
                if (parts.length >= 1) {
                    title = parts[0];
                }
                console.log("[JobFlow][scrape] title from", source, "=>", title);
            } else {
                console.log("[JobFlow][scrape] title from", source, "ignored:", t);
            }
        }
    }

    // --- Company ---

    let company = pickText([
        "a.topcard__org-name-link",
        ".topcard__flavor a",
        "a.jobs-unified-top-card__company-name",
        ".jobs-unified-top-card__company-name a",
        ".jobs-unified-top-card__company-name",
        "header a[href*='/company/'] span",
        "header a[href*='/company/']",
        "a[href*='/company/'] span",
        "a[href*='/company/']",
        "a.app-aware-link[href*='/company/'] span",
        "a.app-aware-link[href*='/company/']"
    ]);

    if (!company) {
        let titleLike = getMeta("og:title") || (document && document.title) || "";
        if (titleLike && /-/.test(titleLike)) {
            const main = titleLike.replace(/\s+\|.*$/, "");
            const parts = main.split(" - ").map((p) => p.trim());
            if (parts.length >= 2 && !/linkedin/i.test(parts[1])) {
                company = parts[1];
            }
        }
        console.log("[JobFlow][scrape] company from meta/title =>", company);
    }

    // --- Location ---

    let location = pickText([
        ".topcard__flavor--bullet",
        ".topcard__flavor.job-card-container__metadata-item",
        ".jobs-unified-top-card__primary-description",
        ".jobs-unified-top-card__subtitle-primary-grouping span",
        ".jobs-unified-top-card__primary-description span",
        "[data-test-job-location]"
    ]);

    // --- Description ---

    const descEl =
        document.querySelector(".jobs-description__content") ||
        document.querySelector(".jobs-box__html-content") ||
        document.querySelector("[data-test='job-description']") ||
        document.querySelector(".description__text") ||
        document.querySelector("section.jobs-description");

    let description = text(descEl);

    if (!description) {
        const ogDesc = getMeta("og:description");
        if (ogDesc) {
            description = ogDesc;
            console.log("[JobFlow][scrape] description from og:description");
        } else {
            description = (document.body.innerText || "").slice(0, 3000);
            console.log("[JobFlow][scrape] description from body fallback");
        }
    }

    if (description.toLowerCase().startsWith("0 notifications")) {
        description = description.replace(/^0 notifications\s*/i, "");
    }

    // --- Infer from description where needed ---

    if (!location && description) {
        const inferredLoc = inferLocationFromDescription(description);
        if (inferredLoc) {
            location = inferredLoc;
            console.log("[JobFlow][scrape] location inferred from description =>", location);
        }
    }

    if (!company && description) {
        const inferredCo = inferCompanyFromDescription(description);
        if (inferredCo) {
            company = inferredCo;
            console.log("[JobFlow][scrape] company inferred from description =>", company);
        }
    }

    // --- Work mode: pills first, then description ---

    let workMode = inferWorkModeFromPills();
    if (!workMode && description) {
        workMode = inferWorkModeFromDescription(description);
    }
    if (workMode) {
        console.log("[JobFlow][scrape] work_mode inferred =>", workMode);
    }

    const payload = {
        external_id: url,
        position: title,
        company: company,
        location: location,
        url: url,
        work_mode: workMode,
        salary: "",          // can be added later
        description: description,
        notes: "Captured from LinkedIn",
        stage: "Applied",
        outcome: "Active",
        next_interview: "",
    };

    console.log("[JobFlow] scraped data (final):", payload);
    return payload;
}


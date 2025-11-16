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
            const [{ result }] = await chrome.scripting.executeScript({
                target: { tabId: tab.id },
                func: scrapeJobFromPage,
            });

            console.log("[JobFlow] Scraped job:", result);

            if (!result || !result.position) {
                statusEl.textContent = "Could not detect a job title on this page.";
                return;
            }

            statusEl.textContent = "Sending to JobFlow…";

            // 3) Send to local Go API
            const resp = await fetch("http://localhost:8081/apply", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(result),
            });

            if (!resp.ok) {
                const txt = await resp.text();
                console.error("[JobFlow] JobFlow API error:", txt);
                statusEl.textContent = "JobFlow returned an error (check server logs).";
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

    const title =
        text(document.querySelector("h1.top-card-layout__title")) ||
        text(document.querySelector("h1.jobs-unified-top-card__job-title")) ||
        text(document.querySelector("h1"));

    const company =
        text(document.querySelector("a.topcard__org-name-link")) ||
        text(document.querySelector(".topcard__flavor a")) ||
        text(document.querySelector("a.jobs-unified-top-card__company-name")) ||
        text(
            document.querySelector("a[href*='/company/'] span") ||
            document.querySelector("a[href*='/company/']")
        ) ||
        text(
            document.querySelector(
                ".jobs-unified-top-card__company-name a, header a[href*='/company/']"
            )
        );

    const location =
        text(document.querySelector(".topcard__flavor--bullet")) ||
        text(
            document.querySelector(".topcard__flavor.job-card-container__metadata-item")
        ) ||
        text(
            document.querySelector(
                ".jobs-unified-top-card__bullet span, .jobs-unified-top-card__workplace-type"
            )
        ) ||
        text(document.querySelector("[data-test-job-location]")) ||
        text(document.querySelector(".jobs-unified-top-card__primary-description"));

    const descEl =
        document.querySelector(".jobs-description__content") ||
        document.querySelector(".jobs-box__html-content") ||
        document.querySelector("[data-test='job-description']") ||
        document.querySelector(".description__text") ||
        document.querySelector("section.jobs-description");

    const description = text(descEl) || document.body.innerText.slice(0, 1500);

    const url = window.location.href;

    const payload = {
        external_id: url,
        position: title,
        company: company,
        location: location,
        url: url,
        work_mode: "",
        salary: "",
        description: description,
        notes: "Captured from LinkedIn",
        stage: "Applied",
        outcome: "Active",
        next_interview: "",
    };

    console.log("[JobFlow] scraped data:", payload);
    return payload;
}

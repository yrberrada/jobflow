// jobflow-extension/popup.js

// Runs inside the LinkedIn tab and returns structured data.
function scrapeJobFromPage() {
    function text(el) {
        return el ? el.innerText.trim() : "";
    }

    // Title
    const title =
        text(document.querySelector("h1.top-card-layout__title")) ||
        text(document.querySelector("h1.jobs-unified-top-card__job-title")) ||
        text(document.querySelector("h1"));

    // Company
    const company =
        text(document.querySelector("a.topcard__org-name-link")) ||
        text(document.querySelector("a.jobs-unified-top-card__company-name")) ||
        text(
            document.querySelector("a[href*='/company/'] span") ||
            document.querySelector("a[href*='/company/']")
        );

    // Location
    const location =
        text(document.querySelector(".topcard__flavor--bullet")) ||
        text(
            document.querySelector(
                ".jobs-unified-top-card__bullet span, .jobs-unified-top-card__workplace-type"
            )
        ) ||
        text(document.querySelector(".topcard__flavor"));

    // Full description
    const descEl =
        document.querySelector(".jobs-description__content") ||
        document.querySelector(".jobs-box__html-content") ||
        document.querySelector("[data-test='job-description']") ||
        document.querySelector(".description__text");

    const description = text(descEl);

    const url = window.location.href;

    console.log("[JobFlow] scraped data:", {
        title,
        company,
        location,
        url,
        description,
    });

    return {
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
}

// Popup logic
document.addEventListener("DOMContentLoaded", () => {
    const btn = document.getElementById("send-job");
    const statusEl = document.getElementById("status");

    if (!btn) {
        console.error("[JobFlow] Button #send-job not found in popup.html");
        if (statusEl) {
            statusEl.textContent = "Popup error: button not found.";
            statusEl.classList.add("status-error");
        }
        return;
    }

    btn.addEventListener("click", async () => {
        statusEl.textContent = "Collecting job info…";
        statusEl.classList.remove("status-ok", "status-error");

        try {
            // 1) Get current active tab
            const [tab] = await chrome.tabs.query({
                active: true,
                currentWindow: true,
            });

            // 2) Run scraper inside LinkedIn tab
            const [injectionResult] = await chrome.scripting.executeScript({
                target: { tabId: tab.id },
                func: scrapeJobFromPage,
            });

            const job = injectionResult && injectionResult.result;
            console.log("[JobFlow] job from page:", job);

            if (!job || !job.position) {
                statusEl.textContent = "Could not read job from this page.";
                statusEl.classList.add("status-error");
                return;
            }

            statusEl.textContent = "Sending to JobFlow…";

            // 3) Send to local Go API
            const resp = await fetch("http://localhost:8081/apply", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(job),
            });

            if (!resp.ok) {
                const txt = await resp.text();
                console.error("[JobFlow] JobFlow API error:", resp.status, txt);
                statusEl.textContent =
                    "JobFlow error (" + resp.status + "). See console.";
                statusEl.classList.add("status-error");
                return;
            }

            const data = await resp.json();
            console.log("[JobFlow] JobFlow response:", data);
            statusEl.textContent = "Sent ✅ (job_id " + data.job_id + ")";
            statusEl.classList.add("status-ok");
        } catch (err) {
            console.error("[JobFlow] unexpected error:", err);
            statusEl.textContent = "Error: " + (err.message || String(err));
            statusEl.classList.add("status-error");
        }
    });
});

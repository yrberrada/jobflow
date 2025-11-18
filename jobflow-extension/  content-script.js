// jobflow-extension/content-script.js

console.log("[JobFlow] content-script loaded on", window.location.href);

// Right now the actual scraping is done from the popup via chrome.scripting.
// You *could* move logic here later if you want automatic scraping on page load.

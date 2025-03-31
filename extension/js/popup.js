document.addEventListener("DOMContentLoaded", async () => {
  const checkBtn = document.getElementById("check");
  const statusText = document.getElementById("status");
  const urlInput = document.getElementById("url");

  // Manual check handler
  checkBtn.addEventListener("click", async () => {
    let urlStr = urlInput.value.trim();

    if (!urlStr) {
      const tabs = await chrome.tabs.query({
        active: true,
        currentWindow: true,
      });
      if (tabs[0]?.url) urlStr = tabs[0].url;
      else return (statusText.textContent = "No URL found");
    }

    try {
      statusText.textContent = "Checking...";
      const { apiEndpoint } = await chrome.storage.local.get("apiEndpoint");

      const response = await fetch(apiEndpoint || DEFAULT_API_ENDPOINT, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ urls: [urlStr] }),
      });

      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();

      statusText.textContent = data.results[0].is_real
        ? "✅ Legitimate website"
        : "❌ Potential phishing";
    } catch (error) {
      statusText.textContent = `Error: ${error.message}`;
    }
  });

  // Load automatic scan results
  const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
  const currentTab = tabs[0];
  if (!currentTab?.url) return;

  const { pageData, urlCache } = await chrome.storage.local.get([
    "pageData",
    "urlCache",
  ]);
  updateResultsDisplay(pageData?.[currentTab.url], urlCache);

  // Real-time updates
  chrome.storage.local.onChanged.addListener((changes) => {
    if (changes.pageData || changes.urlCache) {
      chrome.storage.local.get(["pageData", "urlCache"], (data) => {
        updateResultsDisplay(data.pageData?.[currentTab.url], data.urlCache);
      });
    }
  });
});

function updateResultsDisplay(pageData, urlCache) {
  const linkList = document.getElementById("linkList");
  const totalLinksElem = document.getElementById("totalLinks");
  const phishingLinksElem = document.getElementById("phishingLinks");

  if (!pageData?.links?.length) {
    linkList.innerHTML =
      '<div class="no-links">No links found on this page</div>';
    totalLinksElem.textContent = "0";
    phishingLinksElem.textContent = "0";
    return;
  }

  let phishingCount = 0;
  linkList.innerHTML = "";

  pageData.links.forEach((link) => {
    const li = document.createElement("div");
    const hostname = new URL(link).hostname;
    const result = urlCache?.[hostname];

    li.className = `link-item ${result?.is_real ? "" : "phishing"}`;
    li.innerHTML = `
      <span class="link-url">${hostname}</span>
      <span class="link-status ${result?.is_real ? "legit" : "phishing"}">
        ${result ? (result.is_real ? "Legit" : "Phishing") : "Checking..."}
      </span>
    `;

    if (result && !result.is_real) phishingCount++;
    linkList.appendChild(li);
  });

  totalLinksElem.textContent = pageData.links.length;
  phishingLinksElem.textContent = phishingCount;
}

/* Light mode / Dark mode */

const toggle = document.getElementById("theme-toggle");

// Load the saved theme from localStorage
if (localStorage.getItem("theme") === "dark") {
  document.body.classList.add("dark-mode");
  toggle.checked = true;
}

// Listen for toggle changes
toggle.addEventListener("change", function () {
  if (this.checked) {
    document.body.classList.add("dark-mode");
    localStorage.setItem("theme", "dark");
  } else {
    document.body.classList.remove("dark-mode");
    localStorage.setItem("theme", "light");
  }
});

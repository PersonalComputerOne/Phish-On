document.addEventListener("DOMContentLoaded", async () => {
  const checkBtn = document.getElementById("check");
  const statusText = document.getElementById("status");
  const urlInput = document.getElementById("url");

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

  await refreshCurrentTabData();

  setupRealtimeUpdates();
});

async function getCurrentTab() {
  const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
  return tabs[0];
}

async function refreshCurrentTabData() {
  const currentTab = await getCurrentTab();
  if (!currentTab?.url) return;

  const { pageData, urlCache } = await chrome.storage.local.get([
    "pageData",
    "urlCache",
  ]);

  updateResultsDisplay(pageData?.[currentTab.url], urlCache);
}

function setupRealtimeUpdates() {
  chrome.storage.local.onChanged.addListener(async (changes) => {
    if (changes.pageData || changes.urlCache) {
      await refreshCurrentTabData();
    }
  });

  chrome.tabs.onUpdated.addListener(async (tabId, changeInfo) => {
    const currentTab = await getCurrentTab();
    if (tabId === currentTab?.id && changeInfo.status === "complete") {
      await refreshCurrentTabData();
    }
  });

  chrome.tabs.onActivated.addListener(async () => {
    await refreshCurrentTabData();
  });

  const refreshInterval = setInterval(async () => {
    await refreshCurrentTabData();
  }, 2000);

  window.addEventListener("unload", () => {
    clearInterval(refreshInterval);
  });
}

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
    try {
      const url = new URL(link);
      const hostname = url.hostname;
      const result = urlCache?.[hostname];

      const li = document.createElement("div");
      li.className = `link-item ${result?.is_real === false ? "phishing" : ""}`;

      let statusText = "Checking...";
      let statusClass = "checking";

      if (result) {
        statusText = result.is_real ? "Legit" : "Phishing";
        statusClass = result.is_real ? "legit" : "phishing";
        if (!result.is_real) phishingCount++;
      }

      const displayUrl = truncateUrl(url.toString(), 50);

      li.innerHTML = `
        <span class="link-url" title="${url}">${displayUrl}</span>
        <span class="link-status ${statusClass}">
          ${statusText}
        </span>
      `;

      linkList.appendChild(li);
    } catch (error) {
      console.error(`Invalid URL in links: ${link}`, error);
    }
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
function truncateUrl(url, maxLength) {
  if (url.length <= maxLength) return url;

  const urlObj = new URL(url);
  let shortened = urlObj.origin + "/";

  if (urlObj.pathname.length > 1) {
    const pathParts = urlObj.pathname.split("/").filter(Boolean);
    if (pathParts.length > 0) {
      shortened += pathParts[0] + "/...";
    }
  }

  if (shortened.length > maxLength) {
    return shortened.substring(0, maxLength - 3) + "...";
  }

  return shortened;
}

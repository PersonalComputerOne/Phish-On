const DEFAULT_API_ENDPOINT =
  "http://localhost:8080/api/v1/levenshtein/parallel";
const CACHE_EXPIRATION = 24 * 60 * 60 * 1000;
const PAGE_DATA_EXPIRATION = 30 * 60 * 1000;

// Initial setup
chrome.runtime.onInstalled.addListener(() => {
  chrome.tabs.create({ url: "html/landing.html" });
  chrome.storage.local.get(["apiEndpoint", "urlCache"], (data) => {
    const updates = {};
    if (!data.apiEndpoint) updates.apiEndpoint = DEFAULT_API_ENDPOINT;
    if (!data.urlCache) updates.urlCache = {};
    if (Object.keys(updates).length) chrome.storage.local.set(updates);
  });
});

// Main URL check before navigation
chrome.webNavigation.onBeforeNavigate.addListener(async (details) => {
  if (details.frameId !== 0) return;
  try {
    const url = new URL(details.url);
    if (
      ["chrome:", "chrome-extension:", "about:"].includes(url.protocol) ||
      url.hostname.match(/localhost|127\.0\.0\.\d+/)
    )
      return;

    const { urlCache } = await chrome.storage.local.get("urlCache");
    const now = Date.now();

    if (urlCache[url.hostname]?.timestamp + CACHE_EXPIRATION > now) {
      if (!urlCache[url.hostname].is_real) {
        redirectToBlockPage(details.tabId, urlCache[url.hostname]);
      }
      return;
    }

    const { apiEndpoint } = await chrome.storage.local.get("apiEndpoint");
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    try {
      const response = await fetch(apiEndpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ urls: [details.url] }),
        signal: controller.signal,
      });
      clearTimeout(timeoutId);

      if (!response.ok) return;
      const data = await response.json();
      const result = data.results[0];

      result.timestamp = now;
      urlCache[url.hostname] = result;
      await chrome.storage.local.set({ urlCache });

      if (!result.is_real) redirectToBlockPage(details.tabId, result);
    } catch (error) {
      clearTimeout(timeoutId);
      console.error("API Error:", error);
    }
  } catch (error) {
    console.error("Navigation Error:", error);
  }
});

chrome.webNavigation.onCompleted.addListener(async (details) => {
  if (details.frameId !== 0) return;

  try {
    const url = new URL(details.url);
    if (["chrome:", "chrome-extension:", "about:"].includes(url.protocol))
      return;

    const injectionResults = await chrome.scripting.executeScript({
      target: { tabId: details.tabId },
      func: () => {
        return Array.from(document.querySelectorAll("a"))
          .map((a) => a.href)
          .filter((href) => href && href.startsWith("http"));
      },
    });

    const links = injectionResults[0]?.result || [];

    // Store page data
    const pageData = {
      url: details.url,
      links,
      timestamp: Date.now(),
    };

    // Save page data first
    chrome.storage.local.get(["pageData"], (data) => {
      const pageDataMap = data.pageData || {};
      pageDataMap[details.url] = pageData;

      // Cleanup old entries
      const now = Date.now();
      Object.keys(pageDataMap).forEach((url) => {
        if (now - pageDataMap[url].timestamp > PAGE_DATA_EXPIRATION) {
          delete pageDataMap[url];
        }
      });

      chrome.storage.local.set({ pageData: pageDataMap });
    });

    const { apiEndpoint, urlCache } = await chrome.storage.local.get([
      "apiEndpoint",
      "urlCache",
    ]);
    const endpoint = apiEndpoint || DEFAULT_API_ENDPOINT;
    const now = Date.now();
    let updatedCache = { ...urlCache };

    const processLinks = async () => {
      for (const link of links) {
        try {
          const linkUrl = new URL(link);

          if (
            updatedCache[linkUrl.hostname]?.timestamp + CACHE_EXPIRATION >
            now
          ) {
            continue;
          }

          const response = await fetch(endpoint, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ urls: [link] }),
          });

          if (!response.ok) continue;

          const data = await response.json();
          if (data.results && data.results[0]) {
            const result = data.results[0];
            result.timestamp = now;

            updatedCache[linkUrl.hostname] = result;

            if (Object.keys(updatedCache).length % 5 === 0) {
              await chrome.storage.local.set({ urlCache: updatedCache });
            }
          }
        } catch (error) {
          console.error(`Error checking link: ${link}`, error);
        }
      }

      await chrome.storage.local.set({ urlCache: updatedCache });
    };

    processLinks().catch((error) => {
      console.error("Error in link processing:", error);
    });
  } catch (error) {
    console.error("Link scanning error:", error);
  }
});

function redirectToBlockPage(tabId, result) {
  const blockedUrl = chrome.runtime.getURL(
    `html/blocked.html?input=${encodeURIComponent(
      result.input_url
    )}&similarity_map=${encodeURIComponent(
      JSON.stringify(result.similarity_map)
    )}&is_phishing=${encodeURIComponent(result.is_phishing)}`
  );
  chrome.tabs.update(tabId, { url: blockedUrl });
}

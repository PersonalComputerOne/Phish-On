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

// Optimized link scanning with host-level deduplication
chrome.webNavigation.onCompleted.addListener(async (details) => {
  if (details.frameId !== 0) return;

  try {
    const url = new URL(details.url);
    if (["chrome:", "chrome-extension:", "about:"].includes(url.protocol))
      return;

    // Collect and normalize links
    const injectionResults = await chrome.scripting.executeScript({
      target: { tabId: details.tabId },
      func: () => {
        const hrefs = Array.from(document.querySelectorAll("a"))
          .map((a) => a.href)
          .filter((href) => href && href.startsWith("http"));

        return [...new Set(hrefs)]
          .map((url) => {
            try {
              const u = new URL(url);
              u.hash = "";
              u.search = "";
              return u.toString().replace(/\/$/, "");
            } catch {
              return null;
            }
          })
          .filter((url) => url !== null);
      },
    });

    const links = injectionResults[0]?.result || [];

    // Store page data
    const pageData = {
      url: details.url,
      links,
      timestamp: Date.now(),
    };

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

    // Host-level processing
    const hostProcessor = async () => {
      const { urlCache } = await chrome.storage.local.get("urlCache");
      const { apiEndpoint } = await chrome.storage.local.get("apiEndpoint");
      const now = Date.now();

      // Extract unique hosts
      const uniqueHosts = new Set();
      links.forEach((link) => {
        try {
          const url = new URL(link);
          uniqueHosts.add(url.hostname);
        } catch (error) {
          console.error(`Invalid URL: ${link}`, error);
        }
      });

      // Process each unique host
      Array.from(uniqueHosts).forEach(async (host) => {
        if (urlCache[host]?.timestamp + CACHE_EXPIRATION > now) return;

        try {
          const controller = new AbortController();
          setTimeout(() => controller.abort(), 10000);

          const response = await fetch(apiEndpoint, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ urls: [`http://${host}`] }),
            signal: controller.signal,
          });

          if (!response.ok) return;
          const data = await response.json();
          const result = data.results[0];

          result.timestamp = now;
          urlCache[host] = result;
          await chrome.storage.local.set({ urlCache });
        } catch (error) {
          console.error(`Host processing error (${host}):`, error);
        }
      });
    };

    await hostProcessor();
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
    )}`
  );
  chrome.tabs.update(tabId, { url: blockedUrl });
}

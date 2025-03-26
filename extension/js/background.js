const DEFAULT_API_ENDPOINT =
  "http://localhost:8080/api/v1/levenshtein/parallel";
const CACHE_EXPIRATION = 24 * 60 * 60 * 1000;

chrome.runtime.onInstalled.addListener(() => {
  chrome.tabs.create({ url: "html/landing.html" });
  chrome.storage.local.get(["apiEndpoint", "urlCache"], (data) => {
    const updates = {};
    if (!data.apiEndpoint) {
      updates.apiEndpoint = DEFAULT_API_ENDPOINT;
    }
    if (!data.urlCache) {
      updates.urlCache = {};
    }

    if (Object.keys(updates).length > 0) {
      chrome.storage.local.set(updates);
    }
  });
});

chrome.webNavigation.onBeforeNavigate.addListener(async (details) => {
  if (details.frameId !== 0) return;

  try {
    const url = new URL(details.url);

    if (
      url.protocol === "chrome:" ||
      url.protocol === "chrome-extension:" ||
      url.protocol === "about:" ||
      url.hostname === "localhost" ||
      url.hostname.startsWith("127.0.0.0")
    ) {
      return;
    }

    const { urlCache } = await chrome.storage.local.get("urlCache");
    const now = Date.now();

    if (
      urlCache[url.hostname] &&
      urlCache[url.hostname].timestamp + CACHE_EXPIRATION > now
    ) {
      const cachedResult = urlCache[url.hostname];
      if (!cachedResult.is_real) {
        redirectToBlockPage(details.tabId, cachedResult);
      }
      return;
    }

    const { apiEndpoint } = await chrome.storage.local.get("apiEndpoint");

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); // 5 second timeout

    try {
      const response = await fetch(apiEndpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          urls: [details.url],
        }),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        console.error("API request failed:", response.status);
        return;
      }

      const data = await response.json();

      const result = data.results[0];

      console.log("Result:", JSON.stringify(result));

      result.timestamp = now;

      urlCache[url.hostname] = result;
      await chrome.storage.local.set({ urlCache });

      if (!result.is_real) {
        redirectToBlockPage(details.tabId, result);
      }
    } catch (fetchError) {
      clearTimeout(timeoutId);
      console.error("Fetch error:", fetchError);
    }
  } catch (error) {
    console.error("Error during URL validation:", error);
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

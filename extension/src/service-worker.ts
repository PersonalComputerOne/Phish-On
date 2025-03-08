chrome.runtime.onInstalled.addListener(() => {
  chrome.tabs.create({ url: "landing/index.html" });
});

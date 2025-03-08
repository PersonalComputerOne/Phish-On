document.addEventListener("DOMContentLoaded", () => {
  const checkBtn = document.getElementById("check");
  const statusText = document.getElementById("status");
  const urlInput = document.getElementById("url");

  checkBtn.addEventListener("click", async () => {
    let urlStr = urlInput.value.trim();

    if (!urlStr) {
      const tabs = await new Promise((resolve) =>
        chrome.tabs.query({ active: true, currentWindow: true }, resolve)
      );
      if (tabs[0] && tabs[0].url) {
        urlInput.value = tabs[0].url;
        urlStr = tabs[0].url;
      } else {
        statusText.innerText = "No URL provided.";
        return;
      }
    }

    if (!/^https?:\/\//i.test(urlStr)) {
      urlStr = "https://" + urlStr;
    }

    try {
      statusText.innerText = "Checking website legitimacy...";

      const storageData = await new Promise((resolve) =>
        chrome.storage.local.get("apiEndpoint", resolve)
      );
      const apiEndpoint =
        storageData.apiEndpoint || "http://localhost:8080/api/v1/levenshtein";

      const response = await fetch(apiEndpoint, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ urls: [urlStr] }),
      });

      if (!response.ok) {
        statusText.innerText = `API request failed: ${response.status}`;
        return;
      }

      const data = await response.json();
      const result = data.results[0];

      if (result.is_real) {
        statusText.innerText = "Website is legit.";
      } else {
        statusText.innerText = "Website is not legit.";
      }
    } catch (error) {
      statusText.innerText = `Error checking URL: ${error.message}`;
    }
  });
});

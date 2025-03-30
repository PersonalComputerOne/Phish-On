const urlParams = new URLSearchParams(window.location.search);
const attemptedUrl = urlParams.get("input");
const suggestedUrl = urlParams.get("similarity_map");

document.getElementById("attempted-url").textContent =
  attemptedUrl || "Unknown URL";

const suggestionContainer = document.querySelector(".suggestion-container");

suggestionContainer.innerHTML = "";

if (suggestedUrl) {
  try {
    const suggestions = JSON.parse(suggestedUrl);

    if (Object.keys(suggestions).length === 0) {
      suggestionContainer.innerHTML = `<p style="color: #999; font-style: italic;">No suggestions available</p>`;
    } else {
      Object.keys(suggestions).forEach((key) => {
        const fullUrl = key.startsWith("http") ? key : `https://${key}`;

        const suggestionDiv = document.createElement("div");
        suggestionDiv.classList.add("suggestion");

        const linkElement = document.createElement("a");
        linkElement.href = fullUrl;
        linkElement.textContent = fullUrl;
        linkElement.target = "_blank";

        suggestionDiv.appendChild(linkElement);

        suggestionContainer.appendChild(suggestionDiv);
      });
    }
  } catch (error) {
    console.error("Error parsing JSON:", error);
    suggestionContainer.innerHTML = `<p style="color: red;">Invalid suggestion data</p>`;
  }
} else {
  suggestionContainer.innerHTML = `<p style="color: #999; font-style: italic;">No suggestions available</p>`;
}

const leaveButton = document.getElementById("leave-btn");
leaveButton.addEventListener("click", () => {
  window.close();
  chrome.tabs.create({ url: "chrome://newtab" });
});

const visitButton = document.getElementById("visit-btn");
visitButton.addEventListener("click", () => {
  window.open("_self");
});

function showPhishingWarning(isPhishing) {
  const warningText = document.getElementById("phishing-warning");
  const defaultText = document.getElementById("default-warning");

  if (isPhishing) {
    warningText.style.display = "block";
    defaultText.style.display = "none";
  } else {
    warningText.style.display = "none";
  }
}

// const isPhishingDetected = true;
// showPhishingWarning(isPhishingDetected);
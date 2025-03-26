const urlParams = new URLSearchParams(window.location.search);
const attemptedUrl = urlParams.get("input");
const suggestedUrl = urlParams.get("closest");

document.getElementById("attempted-url").textContent =
  attemptedUrl || "Unknown URL";

const suggestedElement = document.getElementById("suggested-url");
if (suggestedUrl) {
  suggestedElement.textContent = `https://${suggestedUrl}`;
  const fullUrl = suggestedUrl.startsWith("http")
    ? suggestedUrl
    : `https://${suggestedUrl}`;
  suggestedElement.href = fullUrl;
} else {
  suggestedElement.textContent = "No suggestion available";
  suggestedElement.removeAttribute("href");
  suggestedElement.style.color = "#999";
  suggestedElement.style.fontStyle = "italic";
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
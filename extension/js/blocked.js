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

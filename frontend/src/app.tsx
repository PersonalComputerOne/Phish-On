import { useEffect, useState } from "react";

function App() {
  const [urls, setUrls] = useState<string[]>([]);

  useEffect(() => {
    const getUrls = async () => {
      const [tab] = await chrome.tabs.query({
        active: true,
        currentWindow: true,
      });

      const data = await chrome.scripting.executeScript({
        target: { tabId: tab.id! },
        func: () => {
          return Array.from(document.querySelectorAll("a")).map((i) => i.href);
        },
      });

      setUrls(data[0].result!);
    };

    getUrls();
  }, []);

  return (
    <>
      {urls.length === 0 ? (
        <div>No urls</div>
      ) : (
        urls.map((url, i) => <div key={i}>{url}</div>)
      )}
    </>
  );
}

export default App;

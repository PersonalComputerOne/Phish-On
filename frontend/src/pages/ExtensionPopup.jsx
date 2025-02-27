import React, { useState } from "react";
import "../styles/Popup.css";

const ExtensionPopup = () => {
    const [url, setUrl] = useState("");

    const handleCheck = () => {
        alert(`Checking URL: ${url}`);
    };

    return (
        <div className="popup-container">
            <h2>Phish On!</h2>
            <p className="subtitle">Block Phishing Website</p>

            <div className="details">
                <p><strong>Detected URL:</strong> _______________</p>
                <p><strong>Closest Possible URL:</strong> _______________</p>
                <p><strong>Server Type:</strong> _______________</p>
                <p><strong>IP Address:</strong> _______________</p>
                <p><strong>Levenshtein Distance:</strong> _______________</p>
            </div>

            <div className="url-checker">
                <label>URL Checker:</label>
                <textarea
                    placeholder="Paste here"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                />
                <button className="check-btn" onClick={handleCheck}>Check</button>
            </div>
        </div>
    );
};

export default ExtensionPopup;

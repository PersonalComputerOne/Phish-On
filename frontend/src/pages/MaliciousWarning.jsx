import React from "react";
import { useNavigate } from "react-router-dom";
import logo from "../assets/images/logo.png";
import "../styles/Warnings.css";

const MaliciousWarning = () => {
    const navigate = useNavigate(); // React Router for navigation

    // Function to return to the previous page
    const handleLeave = () => {
        if (window.history.length > 1) {
            window.history.back(); //Previous page
        } else {
            navigate("/"); // If no previous history, go to home
        }
    };

    return (
        <>
            {/* Header Section */}
            <div className="malicious-header">
                <h1 className="title">Phish On!</h1>
                <button className="leave-btn" onClick={handleLeave}>Leave</button>
            </div>
            <div className="container">

                {/* Main Warning Content */}
                <div className="malicious-content">
                    <img src={logo} alt="Phish-On!-Icon" className="malicious-icon" />
                    <h1>Caution!</h1>
                    <h1>Malicious Website!</h1>

                    {/* Website Details */}
                    <div className="malicious-details">
                        <p><strong>Detected URL:</strong> _________________</p>
                        <p><strong>Closest Possible URL:</strong> _______________</p>
                        <p><strong>Server Type:</strong> _______________</p>
                        <p><strong>IP Address:</strong> _______________</p>
                        <p><strong>Levenshtein Distance:</strong> _______________</p>
                    </div>

                    {/* Warning Message */}
                    <p className="malicious-text">
                        This website is malicious and may be designed to steal your personal information.
                        Proceed with caution or avoid entering any sensitive data.
                    </p>

                    {/* Buttons for User Action */}
                    <div className="malicious-buttons">
                        <button className="report-btn">Report Mistake</button>
                        <button className="visit-btn">Visit Anyway</button>
                    </div>
                </div>
                <p className="footer-malicious">© 2025 Phish On! All rights reserved.</p>
            </div>
        </>
    );
};

export default MaliciousWarning;

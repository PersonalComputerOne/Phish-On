import React from "react";
import { Link } from "react-router-dom";
import "../styles/HeaderFooter.css";
import logo from "../assets/images/logo.png";

const Header = ({ scrollToAbout }) => {
    return (
        <header className="header">
            <div className="logo-container">
                <img src={logo} alt="Phish On! Logo" className="logo" />
                <h1 className="logo-text">Phish On!</h1>
            </div>
            <nav className="nav-links">
                <button className="btn-about" onClick={scrollToAbout}>About</button>
                <Link to="/" className="btn-get-started">Get Started</Link>
            </nav>
        </header>
    );
};

export default Header;

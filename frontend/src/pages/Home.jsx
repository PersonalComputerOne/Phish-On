import React, { useRef } from "react";
import { Link } from "react-router-dom";
import Header from "../components/Header";
import Carousel from "../components/Carousel";
import Footer from "../components/Footer";
import "../styles/Home.css";

const Home = () => {
    const aboutRef = useRef(null); // Reference to About section

    const scrollToAbout = () => {
        aboutRef.current.scrollIntoView({ behavior: "smooth" });
    };

    return (
        <div className="home-container">
            <Header scrollToAbout={scrollToAbout} /> {/* Pass function to Header */}
            <>
                <main>
                    <h1>Phish On! - Empowering Safe Browsing.</h1>
                    <p className="subtitle-home">Real-time phishing URL protection.</p>

                    <Carousel />

                    <Link to="/phishing-warning">
                        <button className="try-now">Try Now</button>
                    </Link>

                    <p ref={aboutRef} className="mini-text">
                        <em>This tool detects phishing URLs in real-time, keeping users safe from online threats.</em>
                    </p>

                    {/* About Section */}
                    <section className="about">
                        <h1>About Phish On!</h1>
                        <div className="text-container">
                            <p>
                                As phishing scams grow increasingly sophisticated, Phish On! aims to shield users by identifying and warning against URLs that closely resemble legitimate websites.
                                Initiated as a solution for real-time phishing detection, Phish On! leverages the Levenshtein algorithm to measure URL similarity against a comprehensive database of verified sites, immediately alerting users to potentially malicious links.
                            </p>
                            <p>
                                The core functionalities of Phish On! include real-time URL scanning, similarity analysis, and instant alerts to flag potentially dangerous links.
                                The extension processes each URL within the Document Object Model (DOM) of a webpage, analyzing links as they appear.
                            </p>
                            <p>
                                By employing parallel computing techniques, Phish On! optimizes its detection capabilities, allowing for rapid analysis across multiple links.
                                As a result, users can quickly recognize suspicious sites, minimizing the risk of unauthorized access to sensitive information.
                            </p>
                        </div>
                    </section>
                </main>
            </>
            <Footer />
        </div>
    );
};

export default Home;

import React, { useState } from "react";
import "../styles/Carousel.css";
import maliciousWebpage from "../assets/images/malicious-webpage.png";
import phishingWebpage from "../assets/images/phishing-webpage.png";
import extensionImg from "../assets/images/extension.png";

const images = [
    { id: 1, src: maliciousWebpage, alt: "Malicious Webpage" },
    { id: 2, src: phishingWebpage, alt: "Phishing Webpage" },
    { id: 3, src: extensionImg, alt: "Browser Extension" }
];

const Carousel = () => {
    const [currentIndex, setCurrentIndex] = useState(0);
    const [selectedImage, setSelectedImage] = useState(null);

    const nextSlide = () => {
        setCurrentIndex((prev) => (prev + 1) % images.length);
    };

    const prevSlide = () => {
        setCurrentIndex((prev) => (prev - 1 + images.length) % images.length);
    };

    return (
        <div className="carousel">
            {/* Enlarged Image */}
            {selectedImage && (
                <div className="modal" onClick={() => setSelectedImage(null)}>
                    <img src={selectedImage} alt="Enlarged" className="modal-image" />
                </div>
            )}

            {/* Main Carousel */}
            <div className="carousel-content">
                <button className="prev" onClick={prevSlide}>&#10094;</button>

                <div className="image-container">
                    <img
                        src={images[currentIndex].src}
                        alt={images[currentIndex].alt}
                        className="carousel-image"
                        onClick={() => setSelectedImage(images[currentIndex].src)}
                    />
                    {images.length > 1 && (
                        <img
                            src={images[(currentIndex + 1) % images.length].src}
                            alt={images[(currentIndex + 1) % images.length].alt}
                            className="carousel-image"
                            onClick={() => setSelectedImage(images[(currentIndex + 1) % images.length].src)}
                        />
                    )}
                </div>

                <button className="next" onClick={nextSlide}>&#10095;</button>
            </div>

            {/* Thumbnails */}
            <div className="thumbnails">
                {images.map((image, index) => (
                    <img
                        key={image.id}
                        src={image.src}
                        alt={image.alt}
                        className={index === currentIndex ? "active thumbnail-image" : "thumbnail-image"}
                        onClick={() => setCurrentIndex(index)}
                    />
                ))}
            </div>
        </div>
    );
};

export default Carousel;

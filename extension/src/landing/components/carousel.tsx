import React, { useState } from "react";
import "../styles/Carousel.css";
import maliciousWebpage from "../assets/malicious-webpage.png";
import phishingWebpage from "../assets/phishing-webpage.png";
import extensionImg from "../assets/extension.png";

const IMAGES = [
  { id: 1, src: maliciousWebpage, alt: "Malicious Webpage" },
  { id: 2, src: phishingWebpage, alt: "Phishing Webpage" },
  { id: 3, src: extensionImg, alt: "Browser Extension" },
];

export default function Carousel() {
  const [currentIndex, setCurrentIndex] = useState<number>(0);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);

  const nextSlide = () => {
    setCurrentIndex((prev) => (prev + 1) % IMAGES.length);
  };

  const prevSlide = () => {
    setCurrentIndex((prev) => (prev - 1 + IMAGES.length) % IMAGES.length);
  };

  return (
    <div className="carousel">
      {/* Enlarged Image */}
      {selectedImage && (
        <div
          className="modal"
          onClick={() => setSelectedImage(null)}
        >
          <img
            alt="Enlarged"
            className="modal-image"
            src={selectedImage}
          />
        </div>
      )}

      {/* Main Carousel */}
      <div className="carousel-content">
        <button
          className="prev"
          onClick={prevSlide}
        >
          &#10094;
        </button>

        <div className="image-container">
          <img
            alt={IMAGES[currentIndex].alt}
            className="carousel-image"
            src={IMAGES[currentIndex].src}
            onClick={() => setSelectedImage(IMAGES[currentIndex].src)}
          />
          {IMAGES.length > 1 && (
            <img
              alt={IMAGES[(currentIndex + 1) % IMAGES.length].alt}
              className="carousel-image"
              src={IMAGES[(currentIndex + 1) % IMAGES.length].src}
              onClick={() =>
                setSelectedImage(IMAGES[(currentIndex + 1) % IMAGES.length].src)
              }
            />
          )}
        </div>

        <button
          className="next"
          onClick={nextSlide}
        >
          &#10095;
        </button>
      </div>

      {/* Thumbnails */}
      <div className="thumbnails">
        {IMAGES.map((image, index) => (
          <img
            key={image.id}
            alt={image.alt}
            className={
              index === currentIndex
                ? "active thumbnail-image"
                : "thumbnail-image"
            }
            src={image.src}
            onClick={() => setCurrentIndex(index)}
          />
        ))}
      </div>
    </div>
  );
}

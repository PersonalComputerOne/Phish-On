const IMAGES = [
  { id: 1, src: "/assets/malicious.png", alt: "Malicious Webpage" },
  { id: 2, src: "/assets/phishing.png", alt: "Phishing Webpage" },
  { id: 3, src: "/assets/popup-screen.png", alt: "Browser Extension" },
];

document.addEventListener("DOMContentLoaded", () => {
  const aboutBtn = document.getElementById("about-btn");
  const aboutTarget = document.getElementById("about-target");
  aboutBtn.addEventListener("click", () => {
    aboutTarget.scrollIntoView({ behavior: "smooth" });
  });

  let currentIndex = 0;
  const currentImageEl = document.getElementById("current-image");
  const currImageEl = document.getElementById("curr-image");
  const prevBtn = document.getElementById("prev-btn");
  const nextBtn = document.getElementById("next-btn");
  const thumbnailsContainer = document.getElementById("thumbnails");
  const modal = document.getElementById("modal");
  const modalImage = document.getElementById("modal-image");

  function renderThumbnails() {
    thumbnailsContainer.innerHTML = "";
    IMAGES.forEach((img, index) => {
      const thumb = document.createElement("img");
      thumb.src = img.src;
      thumb.alt = img.alt;
      thumb.className = "thumbnail-image";
      if (index === currentIndex) {
        thumb.classList.add("active");
      }
      thumb.addEventListener("click", () => {
        currentIndex = index;
        updateCarousel();
      });
      thumbnailsContainer.appendChild(thumb);
    });
  }

  function updateCarousel() {
    currentImageEl.src = IMAGES[currentIndex].src;
    currentImageEl.alt = IMAGES[currentIndex].alt;

    let nextIndex = (currentIndex + 1) % IMAGES.length;
    currImageEl.src = IMAGES[nextIndex].src;
    currImageEl.alt = IMAGES[nextIndex].alt;

    document.querySelectorAll("#thumbnails img").forEach((thumb, idx) => {
      if (idx === currentIndex) {
        thumb.classList.add("active");
      } else {
        thumb.classList.remove("active");
      }
    });
  }

  nextBtn.addEventListener("click", () => {
    currentIndex = (currentIndex + 1) % IMAGES.length;
    updateCarousel();
  });

  prevBtn.addEventListener("click", () => {
    currentIndex = (currentIndex - 1 + IMAGES.length) % IMAGES.length;
    updateCarousel();
  });

  currentImageEl.addEventListener("click", () => {
    modal.style.display = "flex";
    modalImage.src = currentImageEl.src;
    modalImage.alt = currentImageEl.alt;
  });

  currImageEl.addEventListener("click", () => {
    modal.style.display = "flex";
    modalImage.src = currImageEl.src;
    modalImage.alt = currImageEl.alt;
  });

  modal.addEventListener("click", () => {
    modal.style.display = "none";
  });

  renderThumbnails();
  updateCarousel();
});

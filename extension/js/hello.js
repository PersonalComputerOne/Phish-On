const greetings = [
  "👋 Hello",
  "👋 Hola",
  "👋 Bonjour",
  "👋 Hallo",
  "👋 Ciao",
  "👋 こんにちは",
  "👋 안녕하세요",
  "👋 你好",
  "👋 Привет",
  "👋 مرحبا",
];

let index = 0;
const helloText = document.getElementById("hello-text");

setInterval(() => {
  helloText.style.opacity = 1;
  setTimeout(() => {
    index = (index + 1) % greetings.length;
    helloText.textContent = greetings[index];
    helloText.style.opacity = 1;
  }, 100);
}, 100);

window.onload = function () {
  const loadingScreen = document.getElementById("loading-screen");
  const content = document.getElementById("content");

  // Fade out loading screen
  loadingScreen.style.opacity = "1";
  setTimeout(() => {
    loadingScreen.style.display = "none";
    content.style.display = "block";
    setTimeout(() => {
      content.style.opacity = "1";
      content.style.transform = "translateY(0)";
    }, 300);
  }, 1500);
};

const images = document.querySelectorAll(".carousel-image"); // Select carousel images
const body = document.body;

// Function to disable scrolling
const disableScroll = () => {
  body.classList.add("modal-open");
};

// Function to enable scrolling
const enableScroll = () => {
  body.classList.remove("modal-open");
};

// Add event listener to all images in the carousel
images.forEach((image) => {
  image.addEventListener("click", (event) => {
    event.stopPropagation(); // Prevent click from immediately closing
    disableScroll(); // Disable scrolling
  });
});

// Add event listener to the whole document to enable scrolling when clicking anywhere
document.addEventListener("click", enableScroll);

setTimeout(() => {
  const cursorDot = document.querySelector(".dot"); // Ensure it's properly selected
  const Line = document.querySelector(".dot-img"); // Ensure it's properly selected

  window.addEventListener("mousemove", function (e) {
    const posX = e.clientX;
    const posY = e.clientY;

    cursorDot.style.left = `${posX}px`;
    cursorDot.style.top = `${posY}px`;

    Line.animate(
      {
        left: `${posX}px`,
        top: `${posY}px`,
      },
      { duration: 500, fill: "forwards" }
    );
  });
}, 3000); // Delay by 3 seconds

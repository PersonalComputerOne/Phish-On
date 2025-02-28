import { createRoot } from "react-dom/client";
import { StrictMode } from "react";
import App from "./app";
import "./index.css";

const elem = document.getElementById("root")!;
const app = (
  <StrictMode>
    <App />
  </StrictMode>
);

createRoot(elem).render(app);

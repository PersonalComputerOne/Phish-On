import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import MaliciousWarning from "./pages/MaliciousWarning";
import PhishingWarning from "./pages/PhishingWarning";

const App = () => (
  <Routes>
    <Route element={<Home />} path="/" />
    <Route element={<MaliciousWarning />} path="/malicious-warning" />
    <Route element={<PhishingWarning />} path="/phishing-warning" />
  </Routes>
);

export default App;

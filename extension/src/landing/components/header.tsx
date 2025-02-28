import "../styles/header-footer.css";
import logo from "../assets/logo.png";

interface HeaderProps {
  scrollToAbout: () => void;
}

export default function Header({ scrollToAbout }: HeaderProps) {
  return (
    <header className="header">
      <div className="logo-container">
        <img
          alt="Phish On! Logo"
          className="logo"
          src={logo}
        />
        <h1 className="logo-text">Phish On!</h1>
      </div>
      <nav className="nav-links">
        <button
          className="btn-about"
          onClick={scrollToAbout}
        >
          About
        </button>
        <div className="btn-get-started">Get Started</div>
      </nav>
    </header>
  );
}

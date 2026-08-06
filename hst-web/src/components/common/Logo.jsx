import { useTheme } from "../../context/ThemeContext";

const Logo = ({ className = "" }) => {
  const { theme } = useTheme();

  return (
    <img
      src={`/assets/logo/${theme === "light" ? "logo.png" : "logo-light.png"}`}
      alt="Logo"
      className={`h-8 ${className}`}
    />
  );
};

export default Logo;

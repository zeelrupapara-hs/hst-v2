const Icon = ({ Icon, size = 20, isActive = false, className = "" }) => {
  return (
    <Icon
      size={size}
      className={`text-theme-text ${
        isActive ? "!text-primary" : ""
      } ${className}`}
    />
  );
};

export default Icon;

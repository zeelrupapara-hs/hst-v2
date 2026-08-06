import { Spin } from "antd";

const LoadableButton = ({
  lable,
  type = "button",
  isLoading = false,
  loadingLable = "Loading...",
  className = "",
  onClick,
  disabled = false,
  children,
}) => {
  return (
    <button
      type={type}
      className={`btn-primary ${className} ${
        isLoading || disabled ? "opacity-70 cursor-not-allowed" : ""
      }`}
      onClick={onClick}
      disabled={disabled || isLoading}
    >
      {isLoading ? (
        <div className="ant-white-spin flex items-center justify-center gap-3">
          <Spin />
          <span>{loadingLable}</span>
        </div>
      ) : (
        <div className="flex items-center justify-center gap-2">
          {children ? children : <span>{lable}</span>}
        </div>
      )}
    </button>
  );
};

export default LoadableButton;

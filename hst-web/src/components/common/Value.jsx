import { InputNumber } from "antd";
import { LuPlus, LuMinus } from "react-icons/lu";
import Icon from "./Icon";

const Value = ({
  min = 0.01,
  max = 100,
  digits = 2,
  value,
  setValue,
  size = "medium",
  className = "",
  disabled = false,
}) => {
  const sizeClass = { small: "!p-1", medium: "!p-2", large: "!p-3" };
  const allowedKeys = ["Backspace", "ArrowLeft", "ArrowRight", "Home", "End"];

  const power = Math.pow(10, digits);
  const step = 1 / power;

  const updateAmount = (type, inputValue) => {
    let newValue;

    if (type === "increment") {
      newValue = Math.min(max, value + step);
    } else if (type === "decrement") {
      newValue = Math.max(min, value - step);
    } else if (type === "manual") {
      newValue = Number(inputValue);
    }

    setValue(Number(newValue.toFixed(digits)) || null, type);
  };

  return (
    <div
      className={`flex items-center justify-between bg-theme-bg border border-theme-border rounded-sm ${className}`}
    >
      <button
        className={`btn-icon-hover ${sizeClass[size]}`}
        onClick={() => updateAmount("decrement")}
        disabled={disabled || value <= min}
      >
        <Icon Icon={LuMinus} size={16} />
      </button>

      <InputNumber
        variant="borderless"
        size={size}
        // min={min}
        // max={max}
        step={step}
        value={value}
        controls={false}
        disabled={disabled}
        onChange={(val) => updateAmount("manual", val)}
        onKeyDown={(e) => {
          const isNumberOrDot = /[0-9.]/.test(e.key);
          const alreadyHasDot = value?.toString().includes(".");

          if (!isNumberOrDot && !allowedKeys.includes(e.key)) {
            e.preventDefault();
          } else if (e.key === "." && alreadyHasDot) {
            e.preventDefault();
          }
        }}
        className="flex-1 ant-input-number-center"
      />

      <button
        className={`btn-icon-hover ${sizeClass[size]}`}
        onClick={() => updateAmount("increment")}
        disabled={disabled || value >= max}
      >
        <Icon Icon={LuPlus} size={16} />
      </button>
    </div>
  );
};

export default Value;

import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const Buy = ({ symbolId, disabled, createOrder, variant }) => {
  const { digits, last_ask, askColor } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  const textColor = askColor ? `text-${askColor}` : "text-gray";
  const bgColor = askColor ? `bg-${askColor}` : "bg-gray";
  const borderColor = askColor ? `border-${askColor}` : "border-gray";

  return (
    <button
      className={`w-[120px] flex items-center justify-end text-xs rounded-r-sm overflow-hidden ${
        variant === "filled"
          ? `${bgColor} text-white`
          : `border ${borderColor} ${textColor}`
      }`}
      onClick={() => createOrder(0)}
      disabled={disabled}
    >
      <div className="flex-1 p-1 text-end bg-theme-bg">{last_ask?.toFixed(digits || 2)}</div>
      <div className={`p-1 ${bgColor} text-white`}>BUY</div>
    </button>
  );
};

export default Buy;

import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const Sell = ({ symbolId, disabled, createOrder, variant }) => {
  const { digits, last_bid, bidColor } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  const textColor = bidColor ? `text-${bidColor}` : "text-gray";
  const bgColor = bidColor ? `bg-${bidColor}` : "bg-gray";
  const borderColor = bidColor ? `border-${bidColor}` : "border-gray";

  return (
    <button
      className={`w-[120px] flex items-center justify-start text-xs rounded-l-sm overflow-hidden ${
        variant === "filled"
          ? `${bgColor} text-white`
          : `border ${borderColor} ${textColor}`
      }`}
      onClick={() => createOrder(1)}
      disabled={disabled}
    >
      <div className={`p-1 ${bgColor} text-white`}>SELL</div>
      <div className="flex-1 p-1 text-start bg-theme-bg">{last_bid?.toFixed(digits || 2)}</div>
    </button>
  );
};

export default Sell;

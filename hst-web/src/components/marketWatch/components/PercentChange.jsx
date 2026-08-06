import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const PercentChange = ({ symbolId }) => {
  const percentChange = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]?.percentChange
  );

  return (
    <span className={`${percentChange < 0 ? "text-red" : "text-green"}`}>
      {percentChange || "0.00"}%
    </span>
  );
};

export default PercentChange;

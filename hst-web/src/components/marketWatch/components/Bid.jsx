import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const Bid = ({ symbolId }) => {
  const { digits, last_bid, bidColor } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  return (
    <span className={`text-${bidColor}`}>{last_bid?.toFixed(digits || 2)}</span>
  );
};

export default Bid;

import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const Ask = ({ symbolId }) => {
  const { digits, last_ask, askColor } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  return (
    <span className={`text-${askColor}`}>{last_ask?.toFixed(digits || 2)}</span>
  );
};

export default Ask;

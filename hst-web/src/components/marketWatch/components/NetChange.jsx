import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const NetChange = ({ symbolId }) => {
  const netChange = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]?.netChange
  );

  return (
    <span className={`${netChange < 0 ? "text-red" : "text-green"}`}>
      {netChange || "0.00"}
    </span>
  );
};

export default NetChange;

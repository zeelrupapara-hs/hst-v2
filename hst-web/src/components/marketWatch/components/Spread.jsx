import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const Spread = ({ symbolId }) => {
  const spread = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]?.newSpread
  );

  if (spread == null) return "—";
  return spread;
};

export default Spread;
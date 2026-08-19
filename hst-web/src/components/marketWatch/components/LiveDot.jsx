import useSymbolLive from "../../../hooks/useSymbolLive";

// The little status lamp before a symbol's name: green while prices flow, grey when the
// instrument is quiet — a dead feed and a closed session read the same at a glance.
const LiveDot = ({ symbolId }) => {
  const isLive = useSymbolLive(symbolId);

  return (
    <span
      className={`inline-block h-2 w-2 shrink-0 rounded-full ${isLive ? "bg-green" : "bg-gray"}`}
      aria-label={isLive ? "live" : "no quotes"}
    />
  );
};

export default LiveDot;

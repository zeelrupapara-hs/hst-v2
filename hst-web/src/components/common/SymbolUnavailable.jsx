const SymbolUnavailable = ({ symbol, message = "Symbol unavailable — no live quotes" }) => (
  <div className="flex h-full w-full flex-col items-center justify-center gap-2 bg-theme-bg p-6 text-center">
    <p className="text-sm font-medium text-theme-text">{symbol || "Symbol"}</p>
    <p className="text-xs text-theme-text/70">{message}</p>
  </div>
);

export default SymbolUnavailable;

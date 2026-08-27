import useSymbolStore from "../store/useSymbolStore";
import useLiveSymbolStore from "../store/useLiveSymbolStore";
import { isSymbolQuoted } from "../utils/symbol";

const useSymbolQuoted = (symbolId) => {
  const symbol = useSymbolStore((state) => state.symbols?.[symbolId]);
  const live = useLiveSymbolStore((state) => state.liveSymbols?.[symbolId]);
  return isSymbolQuoted(symbol, live);
};

export default useSymbolQuoted;

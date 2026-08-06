import useSymbolStore from "../store/useSymbolStore";
import useLiveSymbolStore from "../store/useLiveSymbolStore";
import { isSymbolLive } from "../utils/symbol";

const useSymbolLive = (symbolId) => {
  const symbol = useSymbolStore((state) => state.symbols?.[symbolId]);
  const live = useLiveSymbolStore((state) => state.liveSymbols?.[symbolId]);

  return isSymbolLive(symbol, live);
};

export default useSymbolLive;

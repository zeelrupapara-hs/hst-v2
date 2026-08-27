import useSymbolStore from "../../store/useSymbolStore";
import { SOCKET_EVENTS } from "../events";

let timer = null;

// An admin edit of a symbol (volumes, trade mode, flags, contract size…) reaches the terminal at once;
// a burst of edits refetches once.
const handleSymbolMessage = (message) => {
  if (message.type !== SOCKET_EVENTS.SYMBOL_UPDATED) return;
  clearTimeout(timer);
  timer = setTimeout(() => useSymbolStore.getState().refreshSymbols(), 500);
};

export default handleSymbolMessage;

import useSymbolStore from "../../store/useSymbolStore";
import useHistoryStore from "../../store/useHistoryStore";
import usePositionStore from "../../store/usePositionStore";
import { SOCKET_EVENTS } from "../events";
import { POSITION_SIDE } from "../../utils/constants";
import { successToast } from "../../components/common/CustomToast";

const handlePositionMessage = (message) => {
  const symbols = useSymbolStore.getState().symbols;
  const positionStore = usePositionStore.getState();
  const historyStore = useHistoryStore.getState();

  const { type, payload: data } = message;
  const {
    id,
    symbol_id,
    side,
    volume,
    open_price,
    close_price,
    stop_loss,
    take_profit,
  } = data;

  const getMessage = () => {
    const symbol = symbols[symbol_id]?.symbol || "";
    const sideLabel = POSITION_SIDE[side];
    const price =
      type === SOCKET_EVENTS.POSITION_CLOSE ? close_price : open_price;

    // Trade info
    let message = `${symbol} ${sideLabel} ${volume} @${price}`;

    // Add SL/TP only if not position close
    if (type !== SOCKET_EVENTS.POSITION_CLOSE) {
      message += `\nSL: ${stop_loss || "-"} | TP: ${take_profit || "-"}`;
    }

    return message;
  };

  const handlePositionCreate = () => {
    positionStore.addPosition(data);
    successToast(getMessage(), "Position Created");
  };

  const handlePositionUpdate = () => {
    positionStore.updatePosition(id, data);
    successToast(getMessage(), "Position Updated");
  };

  const handlePositionClose = () => {
    positionStore.closePosition(id);
    historyStore.addPosition(data);
    successToast(getMessage(), "Position Closed");
  };

  switch (type) {
    case SOCKET_EVENTS.POSITION_CREATE:
      handlePositionCreate();
      break;

    case SOCKET_EVENTS.POSITION_UPDATE:
      handlePositionUpdate();
      break;

    case SOCKET_EVENTS.POSITION_CLOSE:
      handlePositionClose();
      break;

    default:
      break;
  }
};

export default handlePositionMessage;

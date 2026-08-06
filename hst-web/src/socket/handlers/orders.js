import useSymbolStore from "../../store/useSymbolStore";
import useHistoryStore from "../../store/useHistoryStore";
import usePositionStore from "../../store/usePositionStore";
import { SOCKET_EVENTS } from "../events";
import { POSITION_SIDE } from "../../utils/constants";
import { successToast } from "../../components/common/CustomToast";

const handleOrderMessage = (message) => {
  const symbols = useSymbolStore.getState().symbols;
  const historyStore = useHistoryStore.getState();
  const positionStore = usePositionStore.getState();

  const { type, payload: data } = message;
  const {
    id,
    status,
    symbol_id,
    side,
    volume,
    order_limit_price,
    stop_loss,
    take_profit,
  } = data;

  const getMessage = () => {
    const symbol = symbols[symbol_id]?.symbol || "";
    const sideLabel = POSITION_SIDE[side];
    const price = order_limit_price;

    // Trade info
    let message = `${symbol} ${sideLabel} ${volume} @${price}`;

    // Add SL/TP only if not position close
    if (type !== SOCKET_EVENTS.ORDER_CANCEL) {
      message += `\nSL: ${stop_loss || "-"} | TP: ${take_profit || "-"}`;
    }

    return message;
  };

  const handleOrderCreate = () => {
    historyStore.addOrder(data);

    if ([1, 2].includes(status)) {
      positionStore.addOrder(data);
      successToast(getMessage(), "Order Created");
    }
  };

  const handleOrderUpdate = () => {
    historyStore.updateOrder(id, data);

    if ([1, 2].includes(status)) {
      positionStore.updateOrder(id, data);
      successToast(getMessage(), "Order Updated");
    }

    if ([3, 4, 5, 6].includes(status)) positionStore.removeOrder(id);
  };

  const handleOrderCancel = () => {
    positionStore.removeOrder(id);
    successToast(getMessage(), "Order Cancelled");
  };

  switch (type) {
    case SOCKET_EVENTS.ORDER_CREATE:
      handleOrderCreate();
      break;

    case SOCKET_EVENTS.ORDER_UPDATE:
      handleOrderUpdate();
      break;

    case SOCKET_EVENTS.ORDER_CANCEL:
      handleOrderCancel();
      break;

    default:
      break;
  }
};

export default handleOrderMessage;

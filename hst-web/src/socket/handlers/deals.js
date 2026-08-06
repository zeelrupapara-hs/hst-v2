import useHistoryStore from "../../store/useHistoryStore";
import { SOCKET_EVENTS } from "../events";

const handleDealMessage = (message) => {
  const historyStore = useHistoryStore.getState();

  const { type, payload: data } = message;

  switch (type) {
    case SOCKET_EVENTS.DEAL_CREATE:
      historyStore.addDeal(data);
      break;

    default:
      break;
  }
};

export default handleDealMessage;

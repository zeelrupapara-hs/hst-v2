import usePositionStore from "../../store/usePositionStore";
import { SOCKET_EVENTS } from "../events";

// The socket is already subscribed to the group this account trades under, so an admin editing it
// reaches the terminal on its own. Only the currency digits are read here: everything else the
// group decides is worked out by the engine, which republishes the account summary itself.
const handleGroupMessage = (message) => {
  const { type, payload: data } = message;

  switch (type) {
    case SOCKET_EVENTS.GROUP_UPDATED:
      if (data?.currency_digits != null)
        usePositionStore.getState().setCurrencyDigits(data.currency_digits);
      break;

    default:
      break;
  }
};

export default handleGroupMessage;

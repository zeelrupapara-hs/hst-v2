import { errorToast } from "../components/common/CustomToast";
import { clearLocalStorage } from "../utils/utils";
import { fromServerPayload } from "./adapt";
import handleOrderMessage from "./handlers/orders";
import handlePositionMessage from "./handlers/positions";
import handleDealMessage from "./handlers/deals";
import handleJournalMessage from "./handlers/journals";
import handleAlertMessage from "./handlers/alert";
import handleMailMessage from "./handlers/mail";
import handleReportMessage from "./handlers/report";
import handleGroupMessage from "./handlers/group";
import handleSymbolMessage from "./handlers/symbol";

export const handleSocketMessage = (raw) => {
  if (!raw?.type) return;

  // the server's shape becomes this terminal's shape once, here, rather than in every handler
  const message = { ...raw, payload: fromServerPayload(raw.type, raw.payload) };

  const type = message.type.split("_")[0];

  // An account holds one session at a time, so signing in elsewhere ends this one. Without this
  // the terminal stays on screen looking signed in while every feed has already gone silent.
  if (message.type === "session.revoked") {
    errorToast("You have been signed out because this account signed in elsewhere.");
    clearLocalStorage();
    window.location.href = "/login";
    return;
  }

  if (message.type === "bad_request") {
    errorToast(message?.payload?.message);
    return;
  }

  if (message.type === "order_rejected") {
    const { retcode, message: serverMessage } = message.payload ?? {};

    if (retcode === 10010 || serverMessage === "unknown symbol") {
      errorToast("trade attempted from unspecified symbol", "Order Rejected");
    } else {
      errorToast(serverMessage || "Order was rejected", "Order Rejected");
    }
    return;
  }

  switch (type) {
    case "order":
      handleOrderMessage(message);
      break;
    case "position":
      handlePositionMessage(message);
      break;
    case "deal":
      handleDealMessage(message);
      break;
    case "journal":
      handleJournalMessage(message);
      break;
    case "alert":
      handleAlertMessage(message);
      break;
    case "email":
      handleMailMessage(message);
      break;
    case "report":
      handleReportMessage(message);
      break;
    // an admin editing the group this account trades under: the digits money is shown to are the
    // group's to decide, so the terminal takes the new one without waiting for a reload
    case "symbol":
      handleSymbolMessage(message);
      break;
    case "group":
      handleGroupMessage(message);
      break;
    default:
      break;
  }
};

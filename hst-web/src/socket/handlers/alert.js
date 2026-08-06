import useAlertStore from "../../store/useAlertStore";
import { successToast } from "../../components/common/CustomToast";
import { SOCKET_EVENTS } from "../events";

const handleAlertMessage = (message) => {
  const alertStore = useAlertStore.getState();

  const { type, payload: data } = message;

  const getAlertData = () => {
    const [type, condition, triggerPrice, symbolId] = data?.formula?.split(",");

    const alertData = {
      ...data,
      symbol_id: symbolId,
      condition: type + " " + condition,
      trigger_price: triggerPrice,
    };

    return alertData;
  };

  const handleAlertCreate = () => {
    const alertData = getAlertData(data);
    alertStore.addAlert(alertData);
    successToast("Alert created");
  };

  const handleAlertUpdate = () => {
    const alertData = getAlertData(data);
    alertStore.updateAlert(data?.id, alertData);
    successToast("Alert updated");
  };

  const handleAlertDelete = () => {
    alertStore.deleteAlert(data?.id);
    successToast("Alert deleted");
  };

  const handleAlertTrigger = () => {
    const alertData = getAlertData(data);
    alertStore.updateAlert(data?.id, alertData);
    successToast("Alert triggered");
  };

  switch (type) {
    case SOCKET_EVENTS.ALERT_CREATE:
      handleAlertCreate();
      break;

    case SOCKET_EVENTS.ALERT_UPDATE:
      handleAlertUpdate();
      break;

    case SOCKET_EVENTS.ALERT_DELETE:
      handleAlertDelete();
      break;

    case SOCKET_EVENTS.ALERT_TRIGGER:
      handleAlertTrigger();
      break;

    default:
      break;
  }
};

export default handleAlertMessage;

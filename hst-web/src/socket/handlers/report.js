import useReportStore from "../../store/useReportStore";
import { SOCKET_EVENTS } from "../events";

const handleReportMessage = (message) => {
  const reportStore = useReportStore.getState();

  const { type, payload: data } = message;

  switch (type) {
    case SOCKET_EVENTS.REPORT_CREATE:
      reportStore.addReport(data);
      break;

    case SOCKET_EVENTS.REPORT_STATUS:
      reportStore.updateReport(data?.id, data);
      break;

    case SOCKET_EVENTS.REPORT_DELETE:
      reportStore.deleteReport(data?.id);
      break;

    default:
      break;
  }
};

export default handleReportMessage;

import useJournalStore from "../../store/useJournalStore";
import { SOCKET_EVENTS } from "../events";

const handleJournalMessage = (message) => {
  const journalStore = useJournalStore.getState();

  const { type, payload: data } = message;

  switch (type) {
    case SOCKET_EVENTS.JOURNAL_CREATE:
      journalStore.addJournal(data);
      break;

    default:
      break;
  }
};

export default handleJournalMessage;

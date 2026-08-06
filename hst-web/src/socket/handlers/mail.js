import useMailStore from "../../store/useMailStore";
import { successToast } from "../../components/common/CustomToast";
import { SOCKET_EVENTS } from "../events";

const handleMailMessage = (message) => {
  const mailStore = useMailStore.getState();

  const { type, payload: data } = message;

  const handleMailInbox = () => {
    mailStore.addInboxMail(data);
    successToast("Mail received");
  };

  const handleMailOutbox = () => {
    mailStore.addOutboxMail(data);
    successToast("Mail sent");
  };

  const handleMailDraft = () => {
    mailStore.addDraftMail(data);
    successToast("Mail saved as draft");
  };

  const handleMailTrash = () => {
    mailStore.trashMail(data?.id);
    successToast("Mail moved to trash");
  };

  const handleMailDelete = () => {
    mailStore.deleteMail(data?.id);
    successToast("Mail deleted");
  };

  switch (type) {
    case SOCKET_EVENTS.MAIL_INBOX:
      handleMailInbox();
      break;

    case SOCKET_EVENTS.MAIL_OUTBOX:
      handleMailOutbox();
      break;

    case SOCKET_EVENTS.MAIL_DRAFT:
      handleMailDraft();
      break;

    case SOCKET_EVENTS.MAIL_TRASH:
      handleMailTrash();
      break;

    case SOCKET_EVENTS.MAIL_DELETE:
      handleMailDelete();
      break;

    default:
      break;
  }
};

export default handleMailMessage;

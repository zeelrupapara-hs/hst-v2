import { useEffect, useState } from "react";
import { LuPaperclip, LuReply, LuX } from "react-icons/lu";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";
import MailModal from "./MailModal";
import { downloadMailAttachment, getMail } from "../../../../api/request/mail";
import useMailStore from "../../../../store/useMailStore";

const formatSize = (n) =>
  n >= 1024 * 1024 ? `${(n / 1024 / 1024).toFixed(1)} MB` : `${Math.max(1, Math.round(n / 1024))} KB`;

const PreviewModal = ({ isOpen, setIsOpen, data }) => {
  const [mail, setMail] = useState(null);
  const [isReplyOpen, setIsReplyOpen] = useState(false);
  const { markRead } = useMailStore();

  // the single fetch carries the attachments and marks an inbox mail read
  useEffect(() => {
    if (!isOpen || !data?.id) return;
    getMail(data.id)
      .then((r) => {
        const m = r?.data?.data;
        if (!m) return;
        setMail(m);
        if (m.read_at) markRead(data.id, m.read_at);
      })
      .catch(() => setMail(null));
  }, [isOpen, data?.id, markRead]);

  const shown = mail || data;
  const canReply = shown?.folder === 1 && shown?.sender_login;

  return (
    <>
      <ModalComponent isOpen={isOpen} width={800}>
        <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
          <span>Preview</span>

          <div className="flex items-center gap-3">
            {canReply && (
              <button
                title="Reply"
                onClick={() => {
                  setIsOpen(false);
                  setIsReplyOpen(true);
                }}
              >
                <Icon Icon={LuReply} size={18} className="!text-white" />
              </button>
            )}
            <button onClick={() => setIsOpen(false)}>
              <Icon Icon={LuX} size={18} className="!text-white" />
            </button>
          </div>
        </div>

        <div className="p-4">
          <div className="flex flex-col gap-5">
            <div className="p-2 border border-theme-border rounded">
              <p className="font-semibold mb-1">Subject: </p>
              <p>{shown?.subject || ""}</p>
            </div>

            {(shown?.attachments || []).length > 0 && (
              <div className="p-2 border border-theme-border rounded">
                <p className="font-semibold mb-1">Attachments:</p>
                <div className="flex flex-wrap gap-2">
                  {shown.attachments.map((a) => (
                    <button
                      key={a.attachment_id}
                      className="flex items-center gap-1 px-2 py-1 border border-theme-border rounded"
                      onClick={() => downloadMailAttachment(a.attachment_id, a.name)}
                    >
                      <Icon Icon={LuPaperclip} size={14} />
                      {a.name} ({formatSize(a.size)})
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="p-2 border border-theme-border rounded">
              <p className="font-semibold mb-1">Message:</p>
              <div
                className="prose dark:prose-invert"
                dangerouslySetInnerHTML={{ __html: shown?.body || "" }}
              />
            </div>
          </div>
        </div>
      </ModalComponent>

      <MailModal
        isOpen={isReplyOpen}
        setIsOpen={setIsReplyOpen}
        replyTo={isReplyOpen ? shown : null}
      />
    </>
  );
};

export default PreviewModal;

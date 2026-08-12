import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { sendMail } from "@/api/endpoints/mails.js";
import { GroupTreeSelect } from "@/components/ui/GroupTreeSelect.jsx";

/** Compose window: mail the given logins, or every account matching a group mask. */
export function MailDialog({ logins = [], onClose }) {
  const [mask, setMask] = useState("");
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [error, setError] = useState("");
  const [sent, setSent] = useState(null);
  const { offset, onTitlePointerDown } = useDialogDrag("mail");
  const close = useDialogStack(onClose);

  async function handleSend() {
    if (!subject.trim() || !body.trim()) {
      setError("Subject and body are required");
      return;
    }
    const payload = mask.trim()
      ? { group_mask: mask.trim(), subject, body }
      : { logins, subject, body };
    if (!payload.group_mask && !logins.length) {
      setError("No recipients");
      return;
    }
    const res = await sendMail(payload);
    if (!res.ok) {
      setError(res.message || "send failed");
      return;
    }
    setError("");
    setSent(res.data?.sent ?? 0);
    setTimeout(close, 900);
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          width={460}
          height={360}
          className="mail-dialog"
          title="Send Mail"
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              {sent !== null && <span className="mail-sent">sent to {sent} accounts</span>}
              <button type="button" className="config-ok" onClick={handleSend} disabled={sent !== null}>
                Send
              </button>
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="form-grid">
              <label>To</label>
              <input type="text" value={logins.join(", ")} readOnly disabled={!logins.length} />
              <label>Group</label>
              <GroupTreeSelect maskable value={mask} onChange={setMask} onPick={setMask} />
              <label>Subject</label>
              <input
                type="text"
                maxLength={128}
                value={subject}
                autoFocus
                onChange={(e) => setSubject(e.target.value)}
              />
              <label>Body</label>
              <textarea
                className="mail-body"
                rows={8}
                maxLength={4000}
                value={body}
                onChange={(e) => setBody(e.target.value)}
              />
            </div>
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

import { useEffect, useRef, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { RichTextEdit } from "@/components/ui/RichTextEdit.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";
import {
  sendMail,
  previewMail,
  fetchMailTemplates,
  saveMailTemplate,
  deleteMailTemplate,
  uploadMailAttachments,
} from "@/api/endpoints/mails.js";
import { fetchMailServers } from "@/api/endpoints/mailServers.js";
import { useSession } from "@/hooks/useSession.js";

const BODY_LIMIT = 65536;

// MT5 attachment limits: 5 files, 8MB each, 16MB in total, whitelisted types
const ATTACH_MAX_FILES = 5;
const ATTACH_MAX_FILE = 8 * 1024 * 1024;
const ATTACH_MAX_TOTAL = 16 * 1024 * 1024;
const ATTACH_EXTS = new Set([
  "png", "jpg", "jpeg", "bmp", "gif", "zip", "7z", "doc", "xls",
  "docx", "xlsx", "odt", "rtf", "csv", "txt", "log",
]);

export function checkAttachments(files) {
  if (files.length > ATTACH_MAX_FILES) return `up to ${ATTACH_MAX_FILES} files can be attached`;
  let total = 0;
  for (const f of files) {
    if (!ATTACH_EXTS.has(f.name.split(".").pop().toLowerCase())) {
      return `"${f.name}" cannot be attached`;
    }
    if (f.size > ATTACH_MAX_FILE) return `"${f.name}" is over 8MB`;
    total += f.size;
  }
  return total > ATTACH_MAX_TOTAL ? "attachments exceed 16MB in total" : "";
}

const MACROS = [
  { value: "", label: "Macros" },
  { value: "#LOGIN#", label: "Login" },
  { value: "#USERNAME#", label: "Name" },
  { value: "#USER_CURRENCY#", label: "Currency" },
  { value: "#USER_BALANCE#", label: "Balance" },
  { value: "#USER_CREDIT#", label: "Credit" },
  { value: "#USER_EQUITY#", label: "Equity" },
  { value: "#USER_LEVERAGE#", label: "Leverage" },
  { value: "#USER_MARGIN#", label: "Margin" },
  { value: "#USER_MARGIN_FREE#", label: "Free margin" },
  { value: "#USER_MARGIN_LEVEL#", label: "Margin level" },
];

const htmlToText = (html) => html.replace(/<[^>]*>/g, "").trim();

/**
 * Compose window: internal mail and/or email to the accounts a To expression names.
 * The expression takes logins, ranges (1000-2000), group:, country: and city: terms,
 * any of them negated with !, comma-joined and ';'-separated into OR groups.
 */
export function MailDialog({ logins = [], replyTo = null, onClose }) {
  const session = useSession();
  // a reply locks the recipient and continues the thread of the original
  const replyLogin = replyTo ? (replyTo.folder === 2 ? replyTo.recipient_login : replyTo.sender_login) : 0;
  const [to, setTo] = useState(replyTo ? String(replyLogin) : logins.join(", "));
  const [subject, setSubject] = useState(
    replyTo ? (/^re:/i.test(replyTo.subject) ? replyTo.subject : `Re: ${replyTo.subject}`) : ""
  );
  const [body, setBody] = useState(
    replyTo ? `<p><br></p><hr><blockquote>${replyTo.body}</blockquote>` : ""
  );
  const [files, setFiles] = useState([]);
  const [internal, setInternal] = useState(true);
  const [email, setEmail] = useState(false);
  const [serverId, setServerId] = useState(0);
  const [servers, setServers] = useState([]);
  const [templates, setTemplates] = useState([]);
  const [templateId, setTemplateId] = useState(0);
  const [saveName, setSaveName] = useState(null);
  const [count, setCount] = useState(null);
  const [withEmail, setWithEmail] = useState(0);
  const [toError, setToError] = useState("");
  const [error, setError] = useState("");
  const [sent, setSent] = useState(null);
  const subjectRef = useRef(null);
  const focusRef = useRef("body");
  const { offset, onTitlePointerDown } = useDialogDrag("mail");
  const { confirm, confirmElement } = useConfirm();
  const close = useDialogStack(onClose);

  useEffect(() => {
    fetchMailServers().then((res) => {
      if (res.ok) setServers((res.data || []).filter((s) => s.enabled));
    });
    loadTemplates();
  }, []);

  // the count badge follows the expression, a beat behind the typing
  useEffect(() => {
    if (!to.trim()) {
      setCount(null);
      setToError("");
      return;
    }
    const t = setTimeout(async () => {
      const res = await previewMail({ to: to.trim() });
      if (!res.ok) {
        setToError(res.message || "invalid recipients");
        setCount(null);
        return;
      }
      setToError("");
      setCount(res.data?.count ?? 0);
      setWithEmail(res.data?.with_email ?? 0);
    }, 400);
    return () => clearTimeout(t);
  }, [to]);

  async function loadTemplates() {
    const res = await fetchMailTemplates();
    if (res.ok) setTemplates(res.data || []);
  }

  async function pickTemplate(id) {
    setTemplateId(id);
    const tpl = templates.find((t) => t.template_id === id);
    if (!tpl) return;
    const dirty = subject.trim() || htmlToText(body);
    if (dirty && !(await confirm({ title: "Mail", message: "Replace the current subject and body with the template?" }))) {
      return;
    }
    setSubject(tpl.subject);
    setBody(tpl.body);
  }

  async function handleSaveTemplate() {
    const name = (saveName ?? "").trim();
    if (!name) return;
    const res = await saveMailTemplate({ name, subject, body });
    if (!res.ok) {
      setError(res.message || "template save failed");
      return;
    }
    setError("");
    setSaveName(null);
    await loadTemplates();
    setTemplateId(res.data?.template_id ?? 0);
  }

  async function handleDeleteTemplate() {
    const tpl = templates.find((t) => t.template_id === templateId);
    if (!tpl) return;
    if (!(await confirm({ title: "Mail", message: `Delete template "${tpl.name}"?` }))) return;
    const res = await deleteMailTemplate(tpl.template_id);
    if (!res.ok) {
      setError(res.message || "template delete failed");
      return;
    }
    setTemplateId(0);
    await loadTemplates();
  }

  function insertMacro(token, exec) {
    if (!token) return;
    if (focusRef.current === "subject" && subjectRef.current) {
      const el = subjectRef.current;
      const at = el.selectionStart ?? subject.length;
      setSubject(subject.slice(0, at) + token + subject.slice(el.selectionEnd ?? at));
      return;
    }
    exec("insertText", token);
  }

  async function handleSend() {
    const hasBody = htmlToText(body) || body.includes("<img");
    if (!subject.trim() || !hasBody) {
      setError("Subject and body are required");
      return;
    }
    if (!to.trim()) {
      setError("No recipients");
      return;
    }
    if (body.length > BODY_LIMIT) {
      setError("Body is too large");
      return;
    }

    let attachmentIds = [];
    if (files.length) {
      const bad = checkAttachments(files);
      if (bad) {
        setError(bad);
        return;
      }
      const up = await uploadMailAttachments(files);
      if (!up.ok) {
        setError(up.message || "attachment upload failed");
        return;
      }
      attachmentIds = (up.data || []).map((a) => a.attachment_id);
    }

    const res = await sendMail({
      to: to.trim(),
      subject,
      body,
      internal,
      email,
      mail_server_id: Number(serverId) || 0,
      reply_to: replyTo?.tracking_id || "",
      attachment_ids: attachmentIds,
    });
    if (!res.ok) {
      setError(res.message || "send failed");
      return;
    }
    setError("");
    setSent(res.data ?? { sent: 0 });
    setTimeout(close, 1200);
  }

  // MT5 parity: internal mail needs a mailbox name on the sender's manager account
  const noMailbox = internal && !session.mailbox;
  const sendDisabled =
    sent !== null || (!internal && !email) || Boolean(toError) || count === 0 || noMailbox;

  const sentMessage = (r) => {
    const parts = [];
    if (internal) parts.push(`sent to ${r.sent} accounts`);
    if (email) parts.push(`${r.emails_queued} emails queued`);
    if (email && r.skipped_no_email > 0) parts.push(`${r.skipped_no_email} without email`);
    return parts.join(", ");
  };

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
          width={640}
          height={540}
          className="mail-dialog"
          title="Mail"
        >
          <div className="config-panel active">
            <div className="mail-head">
              <label>To:</label>
              <span className="mail-to">
                <input
                  type="text"
                  value={to}
                  readOnly={Boolean(replyTo)}
                  placeholder="logins, 1000-2000, group:demo\*, country:…, city:…, !…"
                  title="Comma-joined terms narrow; ';' starts another OR group; '!' excludes"
                  onChange={(e) => setTo(e.target.value)}
                />
                {count !== null && !toError && (
                  <span className="mail-count" title={email ? `${withEmail} with an email address` : undefined}>
                    {count}
                  </span>
                )}
              </span>
              {toError && (
                <>
                  <span />
                  <span className="mail-to-error">{toError}</span>
                </>
              )}
              <label>Subject:</label>
              <input
                ref={subjectRef}
                type="text"
                maxLength={128}
                value={subject}
                autoFocus
                onFocus={() => { focusRef.current = "subject"; }}
                onChange={(e) => setSubject(e.target.value)}
              />
              <label>Template:</label>
              <span className="mail-cmd-row">
                <span className="mail-template-row">
                  <PropSelect
                    value={templateId}
                    options={[
                      { value: 0, label: "(none)" },
                      ...templates.map((t) => ({ value: t.template_id, label: t.name })),
                    ]}
                    onChange={pickTemplate}
                  />
                  {saveName === null ? (
                    <>
                      <button type="button" className="mail-flat-btn" onClick={() => setSaveName(subject)}>
                        Save…
                      </button>
                      <button
                        type="button"
                        className="mail-flat-btn"
                        disabled={!templateId}
                        onClick={handleDeleteTemplate}
                      >
                        Delete
                      </button>
                    </>
                  ) : (
                    <>
                      <input
                        type="text"
                        className="mail-template-name"
                        value={saveName}
                        autoFocus
                        maxLength={128}
                        placeholder="template name"
                        onChange={(e) => setSaveName(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleSaveTemplate()}
                      />
                      <button type="button" className="mail-flat-btn" onClick={handleSaveTemplate}>OK</button>
                      <button type="button" className="mail-flat-btn" onClick={() => setSaveName(null)}>
                        Cancel
                      </button>
                    </>
                  )}
                </span>
                <label className="mail-inline-label">Mail server:</label>
                <PropSelect
                  disabled={!email}
                  value={serverId}
                  options={[
                    { value: 0, label: "" },
                    ...servers.map((s) => ({ value: s.mail_server_id, label: s.name })),
                  ]}
                  onChange={setServerId}
                />
                <button type="button" className="mail-send-btn" onClick={handleSend} disabled={sendDisabled}>
                  Send
                </button>
              </span>
              <span />
              <span className="mail-channels">
                <label>
                  <input type="checkbox" checked={internal} onChange={(e) => setInternal(e.target.checked)} />
                  Internal mail
                </label>
                <label>
                  <input type="checkbox" checked={email} onChange={(e) => setEmail(e.target.checked)} />
                  Email
                </label>
              </span>
              <label>Attach:</label>
              <span className="mail-attach-row">
                <input
                  type="file"
                  multiple
                  accept={[...ATTACH_EXTS].map((e) => `.${e}`).join(",")}
                  onChange={(e) => {
                    const picked = [...e.target.files];
                    const bad = checkAttachments(picked);
                    setError(bad);
                    if (!bad) setFiles(picked);
                  }}
                />
                {files.length > 0 && <span className="module-note">{files.length} file(s)</span>}
              </span>
              {(error || noMailbox || sent !== null) && (
                <>
                  <span />
                  <span>
                    {error && <span className="login-error">{error}</span>}
                    {!error && noMailbox && (
                      <span className="login-error">Your account has no mailbox name; internal mail is off</span>
                    )}
                    {sent !== null && <span className="mail-sent">{sentMessage(sent)}</span>}
                  </span>
                </>
              )}
            </div>
            <div className="mail-editor" onFocusCapture={() => { focusRef.current = "body"; }}>
              <RichTextEdit
                value={body}
                onChange={setBody}
                extraTools={(exec) => (
                  <PropSelect
                    className="rich-text-macros"
                    value=""
                    options={MACROS}
                    onChange={(token) => insertMacro(token, exec)}
                  />
                )}
              />
            </div>
          </div>
          {confirmElement}
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

import { useEffect, useRef, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { RichTextEdit } from "@/components/ui/RichTextEdit.jsx";
import { GroupTreeSelect } from "@/components/ui/GroupTreeSelect.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";
import {
  sendMail,
  previewMail,
  fetchMailTemplates,
  saveMailTemplate,
  deleteMailTemplate,
} from "@/api/endpoints/mails.js";
import { fetchMailServers } from "@/api/endpoints/mailServers.js";

const BODY_LIMIT = 65536;

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
export function MailDialog({ logins = [], onClose }) {
  const [to, setTo] = useState(logins.join(", "));
  const [groupPick, setGroupPick] = useState("");
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
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

  function appendGroupTerm(mask) {
    if (!mask) return;
    const term = `group:${mask}`;
    setTo((prev) => (prev.trim() ? `${prev.trim()}, ${term}` : term));
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
    const res = await sendMail({
      to: to.trim(),
      subject,
      body,
      internal,
      email,
      mail_server_id: Number(serverId) || 0,
    });
    if (!res.ok) {
      setError(res.message || "send failed");
      return;
    }
    setError("");
    setSent(res.data ?? { sent: 0 });
    setTimeout(close, 1200);
  }

  const sendDisabled =
    sent !== null || (!internal && !email) || Boolean(toError) || count === 0;

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
          title="Send Mail"
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              {sent !== null && <span className="mail-sent">{sentMessage(sent)}</span>}
              <button type="button" className="config-ok" onClick={handleSend} disabled={sendDisabled}>
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
              <span className="mail-to">
                <input
                  type="text"
                  value={to}
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
              <label>Group</label>
              <GroupTreeSelect maskable value={groupPick} onChange={setGroupPick} onPick={appendGroupTerm} />
              <label>Subject</label>
              <input
                ref={subjectRef}
                type="text"
                maxLength={128}
                value={subject}
                autoFocus
                onFocus={() => { focusRef.current = "subject"; }}
                onChange={(e) => setSubject(e.target.value)}
              />
              <label>Template</label>
              <span className="mail-template-row">
                <PropSelect
                  fill
                  value={templateId}
                  options={[
                    { value: 0, label: "(none)" },
                    ...templates.map((t) => ({ value: t.template_id, label: t.name })),
                  ]}
                  onChange={pickTemplate}
                />
                {saveName === null ? (
                  <>
                    <button type="button" onClick={() => setSaveName(subject)}>Save…</button>
                    <button type="button" disabled={!templateId} onClick={handleDeleteTemplate}>
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
                    <button type="button" onClick={handleSaveTemplate}>OK</button>
                    <button type="button" onClick={() => setSaveName(null)}>Cancel</button>
                  </>
                )}
              </span>
              <label>Send as</label>
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
              <label>Mail server</label>
              <PropSelect
                fill
                disabled={!email}
                value={serverId}
                options={[
                  { value: 0, label: "Default" },
                  ...servers.map((s) => ({ value: s.mail_server_id, label: s.name })),
                ]}
                onChange={setServerId}
              />
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

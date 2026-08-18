import { useCallback, useEffect, useState } from "react";
import { deleteMail, downloadMailAttachment, fetchMail, fetchMails } from "@/api/endpoints/mails.js";
import { onEvent } from "@/api/socket.js";
import { useSession } from "@/hooks/useSession.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { MailDialog } from "@/modules/mail/MailDialog.jsx";

// MT5 shows "14:04 | Today" for today's mail, the full stamp otherwise
function mailTime(ns) {
  if (!ns) return "";
  const d = new Date(ns / 1e6);
  const p = (n) => String(n).padStart(2, "0");
  const hm = `${p(d.getHours())}:${p(d.getMinutes())}`;
  const now = new Date();
  if (d.toDateString() === now.toDateString()) return `${hm} | Today`;
  return `${d.getFullYear()}.${p(d.getMonth() + 1)}.${p(d.getDate())} ${hm}`;
}

const kb = (n) =>
  n >= 1024 * 1024 ? `${(n / 1024 / 1024).toFixed(1)} MB` : `${Math.max(1, Math.round(n / 1024))} KB`;

// the reference's little state marks: yellow closed, white open, outgoing, reply
function MailMark({ mail }) {
  const outgoing = mail.folder === 2;
  const unread = mail.folder === 1 && !mail.read_at;
  const reply = /^re:/i.test(mail.subject || "");
  const cls = outgoing ? "mail-mark-out" : unread ? "mail-mark-unread" : "mail-mark-read";
  return (
    <svg className={`mail-mark ${cls}`} width="14" height="11" viewBox="0 0 14 11" aria-hidden>
      <rect x="0.5" y="0.5" width="13" height="10" rx="1" />
      <path d="M0.5 1.5 L7 6.5 L13.5 1.5" fill="none" />
      {outgoing && <path d="M9 8.5 h4 M11 6.5 l2 2 -2 2" className="mail-mark-arrow" fill="none" />}
      {reply && !outgoing && <path d="M5 8.5 h-4 M3 6.5 l-2 2 2 2" className="mail-mark-arrow" fill="none" />}
    </svg>
  );
}

function partyName(mail, side) {
  if (side === "from") return mail.sender_name || String(mail.sender_login);
  if (!mail.recipient_login) return "group";
  return mail.recipient_name || String(mail.recipient_login);
}

/** The reference's mail window: header block, command bar, then the message itself. */
function ViewMailDialog({ trackingId, onClose, onReply, onDeleted }) {
  const [mail, setMail] = useState(null);
  const { confirm, confirmElement } = useConfirm();

  useEffect(() => {
    fetchMail(trackingId).then((res) => res.ok && setMail(res.data));
  }, [trackingId]);

  async function handleDelete() {
    if (!(await confirm({ title: "Mailbox", message: "Delete this message?" }))) return;
    const res = await deleteMail(trackingId);
    if (res.ok) {
      onDeleted();
      onClose();
    }
  }

  return (
    <DialogOverlay>
      <SettingsDialog
        onClose={onClose}
        width={620}
        height={440}
        title="Mail"
        className="mail-view-dialog"
        footer={
          <div className="config-actions">
            <button type="button" className="config-ok" onClick={onClose}>Close</button>
          </div>
        }
      >
        {mail && (
          <div className="config-panel active mail-view">
            <div className="mail-view-head">
              <label>From:</label>
              <span>{mail.sender_login ? `${mail.sender_login}, ` : ""}{mail.sender_name || ""}</span>
              <label>Date:</label>
              <span>{mailTime(mail.created_at)}</span>
              <label>Subject:</label>
              <span>{mail.subject || "(no subject)"}</span>
              <label>Attachment:</label>
              <span className="mail-view-attachments">
                {(mail.attachments || []).map((a) => (
                  <button
                    key={a.attachment_id}
                    type="button"
                    className="mail-attachment"
                    onClick={() => downloadMailAttachment(a.attachment_id, a.name)}
                  >
                    {a.name} ({kb(a.size)})
                  </button>
                ))}
              </span>
            </div>
            <div className="mail-view-commands">
              <button type="button" onClick={() => onReply(mail)}>↩ Reply</button>
              <button type="button" onClick={handleDelete}>✕ Delete</button>
            </div>
            {/* body is sanitized server-side before it is ever stored */}
            <div className="mail-body" dangerouslySetInnerHTML={{ __html: mail.body }} />
          </div>
        )}
        {confirmElement}
      </SettingsDialog>
    </DialogOverlay>
  );
}

/**
 * The manager's Mailbox as the platform draws it: one list of received and sent mail in the
 * main area — Subject, From, To, Time — with reply chains folded under their newest message.
 */
export function MailboxModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [expanded, setExpanded] = useState(() => new Set());
  const [viewing, setViewing] = useState(null);
  const [compose, setCompose] = useState(null);
  const [menu, setMenu] = useState(null);
  const { confirm, confirmElement } = useConfirm();

  const load = useCallback(() => {
    Promise.all([fetchMails({ type: "inbox" }), fetchMails({ type: "outbox" })]).then(
      ([inbox, outbox]) => {
        if (!inbox.ok && !outbox.ok) return;
        const all = [...(inbox.data ?? []), ...(outbox.data ?? [])];
        all.sort((a, b) => b.created_at - a.created_at);
        setRows(all);
      }
    );
  }, []);

  useEffect(() => {
    load();
    return onEvent("email_inbox", () => {
      load();
      session.refreshNav();
    });
  }, [load, session.refreshNav]);

  // fold each conversation under its newest message, the chain opening in place
  const threads = new Map();
  for (const r of rows ?? []) {
    threads.get(r.thread_id)?.push(r) ?? threads.set(r.thread_id, [r]);
  }

  function toggleThread(id) {
    const next = new Set(expanded);
    next.has(id) ? next.delete(id) : next.add(id);
    setExpanded(next);
  }

  function openMail(r) {
    setViewing(r.tracking_id);
    if (r.folder === 1 && !r.read_at) {
      // the fetch marks it read; reflect it here and in the navigator count
      setRows((prev) => prev.map((m) => (m.mail_id === r.mail_id ? { ...m, read_at: 1 } : m)));
      session.refreshNav();
    }
  }

  function replyTo(mail) {
    setViewing(null);
    setCompose({ replyTo: mail });
  }

  const menuRow = menu?.row;

  return (
    <div className="module-root">
      <div
        className="table-wrap"
        onContextMenu={(e) => {
          // the whole panel is the mailbox, so the empty space below the rows answers too
          if (e.defaultPrevented) return;
          e.preventDefault();
          setMenu({ x: e.clientX, y: e.clientY, row: null });
        }}
      >
        <table className="data-table data-table-grid mail-table">
          <thead>
            <tr>
              <th>Subject</th>
              <th>From</th>
              <th>To</th>
              <th>Time</th>
            </tr>
          </thead>
          <tbody>
            {[...threads.entries()].map(([threadId, chain]) => {
              const open = expanded.has(threadId);
              const shown = open ? chain : [chain[0]];
              return shown.map((r, i) => {
                const unread = r.folder === 1 && !r.read_at;
                return (
                  <tr
                    key={r.mail_id}
                    className={unread ? "mail-row-unread" : undefined}
                    onDoubleClick={() => openMail(r)}
                    onContextMenu={(e) => {
                      e.preventDefault();
                      setMenu({ x: e.clientX, y: e.clientY, row: r, threadId, chain });
                    }}
                  >
                    <td>
                      <span className="mail-subject-cell" style={i > 0 ? { paddingLeft: 18 } : undefined}>
                        {i === 0 && chain.length > 1 ? (
                          <button
                            type="button"
                            className="mail-chain-toggle"
                            title={open ? "Collapse the conversation" : "View the entire conversation"}
                            onClick={(e) => {
                              e.stopPropagation();
                              toggleThread(threadId);
                            }}
                          >
                            {open ? "▾" : "▸"}
                          </button>
                        ) : (
                          <span className="mail-chain-spacer" />
                        )}
                        <MailMark mail={r} />
                        <span className="mail-subject-text" onClick={() => openMail(r)}>
                          {r.subject || "(no subject)"}
                        </span>
                      </span>
                    </td>
                    <td>{partyName(r, "from")}</td>
                    <td>{partyName(r, "to")}</td>
                    <td className="mail-time">{mailTime(r.created_at)}</td>
                  </tr>
                );
              });
            })}
            {rows !== null && rows.length === 0 && (
              <tr><td colSpan={4} className="df-empty">No mail</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            {
              label: "New Mail…",
              icon: "mailbox",
              disabled: !session.mailbox,
              onClick: () => setCompose({}),
            },
            { label: "View", disabled: !menuRow, onClick: () => menuRow && openMail(menuRow) },
            {
              label: "Delete",
              disabled: !menuRow,
              onClick: async () => {
                if (!menuRow) return;
                if (!(await confirm({ title: "Mailbox", message: "Delete this message?" }))) return;
                const res = await deleteMail(menuRow.tracking_id);
                if (res.ok) load();
              },
            },
            {
              label: "Expand",
              disabled: !menu.chain || menu.chain.length < 2,
              onClick: () => menu.threadId && toggleThread(menu.threadId),
            },
            "sep",
            { label: "Refresh", onClick: load },
          ]}
        />
      )}

      {viewing && (
        <ViewMailDialog
          trackingId={viewing}
          onClose={() => setViewing(null)}
          onReply={replyTo}
          onDeleted={load}
        />
      )}
      {compose && (
        <MailDialog
          replyTo={compose.replyTo}
          onClose={() => {
            setCompose(null);
            load();
          }}
        />
      )}
      {confirmElement}
    </div>
  );
}

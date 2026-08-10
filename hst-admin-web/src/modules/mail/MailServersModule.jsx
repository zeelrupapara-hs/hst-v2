import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createMailServer,
  deleteMailServer,
  fetchMailServers,
  updateMailServer,
} from "@/api/endpoints/mailServers.js";

const PATCH_FIELDS = ["enabled", "name", "sender_email", "sender_name", "smtp_server", "smtp_login", "is_default"];

function MailServerDialog({ server, onClose, onSaved }) {
  const isNew = !server;
  const [draft, setDraft] = useState(
    server
      ? { ...server, smtp_password: "" }
      : { enabled: true, name: "", sender_email: "", sender_name: "", smtp_server: "", smtp_login: "", smtp_password: "", is_default: false },
  );
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(server?.mail_server_id ?? "new");
  const close = useDialogStack(onClose);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    if (!draft.name.trim() || !draft.sender_email.trim() || !draft.smtp_server.trim()) {
      setError("Name, sender email and server are required");
      return;
    }
    let res;
    if (isNew) {
      res = await createMailServer(draft);
    } else {
      const patch = {};
      for (const key of PATCH_FIELDS) {
        if (draft[key] !== server[key]) patch[key] = draft[key];
      }
      // an empty password keeps the stored one
      if (draft.smtp_password) patch.smtp_password = draft.smtp_password;
      res = Object.keys(patch).length
        ? await updateMailServer(server.mail_server_id, patch)
        : { ok: true };
    }
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    close();
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
          width={531}
          height={420}
          title={isNew ? "Mail Server" : `Mail Server: ${server.name}`}
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="mailbox" size={48} />
              </span>
              <p>
                The SMTP account the platform sends reports, statements and notifications from.
                The default server carries every group that names no other.
              </p>
            </div>
            <div className="form-grid grp-check-stack">
              <label className="sym-check">
                <input
                  type="checkbox"
                  checked={!!draft.enabled}
                  onChange={(e) => set("enabled", e.target.checked)}
                />
                <span>Enable</span>
              </label>
              <label className="sym-check">
                <input
                  type="checkbox"
                  checked={!!draft.is_default}
                  onChange={(e) => set("is_default", e.target.checked)}
                />
                <span>Use as the default server</span>
              </label>
            </div>
            <div className="form-grid">
              <label>Name</label>
              <input type="text" className="wide" value={draft.name} onChange={(e) => set("name", e.target.value)} />
              <label>Sender email</label>
              <input type="text" className="wide" value={draft.sender_email} onChange={(e) => set("sender_email", e.target.value)} />
              <label>Sender name</label>
              <input type="text" className="wide" value={draft.sender_name} onChange={(e) => set("sender_name", e.target.value)} />
              <label>SMTP server</label>
              <span className="grp-suffixed">
                <input type="text" value={draft.smtp_server} onChange={(e) => set("smtp_server", e.target.value)} />
                <span className="grp-suffix">host:port</span>
              </span>
              <label>SMTP login</label>
              <input type="text" value={draft.smtp_login} onChange={(e) => set("smtp_login", e.target.value)} />
              <label>SMTP password</label>
              <span className="grp-suffixed">
                <input
                  type="password"
                  value={draft.smtp_password}
                  onChange={(e) => set("smtp_password", e.target.value)}
                />
                {!isNew && <span className="grp-suffix">empty keeps the stored one</span>}
              </span>
            </div>
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

const avg = (r) => (r.total_sent > 0 ? `${Math.round(r.time_sum_ms / r.total_sent)} ms` : "—");

/** The SMTP accounts the platform sends from; one of them is the default. */
export function MailServersModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_mails !== false;

  const load = () => fetchMailServers().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

  const row = selected != null ? rows?.[selected] : null;

  async function onDelete(r) {
    if (!window.confirm(`Delete mail server '${r.name}'?`)) return;
    const res = await deleteMailServer(r.mail_server_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  async function setDefault(r) {
    const res = await updateMailServer(r.mail_server_id, { is_default: true });
    if (!res.ok) window.alert(res.message || "failed");
    load();
  }

  function saved() {
    load();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Name</th>
              <th>Sender</th>
              <th>Server</th>
              <th>Default</th>
              <th className="num">Sent</th>
              <th className="num">Errors</th>
              <th className="num">Avg time</th>
              <th className="num">Queue</th>
              <th>State</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((r, i) => (
              <tr
                key={r.mail_server_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ server: r })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="mailbox" />
                    {r.name}
                  </span>
                </td>
                <td>{r.sender_name ? `${r.sender_name} <${r.sender_email}>` : r.sender_email}</td>
                <td>{r.smtp_server}</td>
                <td>{r.is_default ? "Default" : ""}</td>
                <td className="num">{r.total_sent}</td>
                <td className="num">{r.total_errors}</td>
                <td className="num">{avg(r)}</td>
                <td className="num">{r.queue}</td>
                <td className={r.enabled ? "" : "nav-feed-disabled"}>{r.enabled ? "Enabled" : "Disabled"}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr>
                <td colSpan={9} className="df-empty">No mail servers configured</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {menu && canEdit && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            ...listMenuHead({
              onAdd: () => setDialog({ server: null }),
              onEdit: () => setDialog({ server: row }),
              onDelete: () => onDelete(row),
              hasSelection: !!row,
            }),
            "sep",
            { label: "Use as Default", disabled: !row || row.is_default, onClick: () => setDefault(row) },
            ...listMenuTail({}),
          ]}
        />
      )}
      {dialog && (
        <MailServerDialog server={dialog.server} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}

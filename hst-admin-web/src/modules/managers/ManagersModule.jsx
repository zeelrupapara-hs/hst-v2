import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createManager,
  deleteManager,
  fetchManagerRights,
  fetchManagers,
  updateManager,
} from "@/api/endpoints/managers.js";
import { humanize } from "@/lib/labels.js";

function ManagerDialog({ manager, onClose, onSaved }) {
  const isNew = !manager;
  const [tab, setTab] = useState("Common");
  const [draft, setDraft] = useState({
    login: manager?.login ?? "",
    name: manager?.name ?? "",
    mailbox: manager?.mailbox ?? "",
    groups: (manager?.groups ?? ["*"]).join("; "),
  });
  const [rights, setRights] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(manager?.login ?? "new");

  useEffect(() => {
    if (isNew) {
      setRights({});
      return;
    }
    fetchManagerRights(manager.login).then((res) => res.ok && setRights(res.data.rights || {}));
  }, [manager, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    if (!draft.name.trim() || (isNew && !Number(draft.login))) {
      setError("Login and name are required");
      return;
    }
    const body = {
      name: draft.name.trim(),
      mailbox: draft.mailbox,
      groups: draft.groups.split(";").map((s) => s.trim()).filter(Boolean),
      rights: Object.entries(rights || {}).filter(([, on]) => on).map(([key]) => key),
    };
    const res = isNew
      ? await createManager({ ...body, login: Number(draft.login) })
      : await updateManager(manager.login, body);
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    onClose();
  }

  const rightKeys = Object.keys(rights || {}).sort();

  return (
    <div className="dialog-overlay" onClick={onClose} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          draggable
          height={440}
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "Manager: New" : `Manager: ${manager.login} — ${manager.name}`}
          tabs={
            <div className="config-tabs">
              {["Common", "Permissions"].map((t) => (
                <button key={t} type="button" className={tab === t ? "active" : ""} onClick={() => setTab(t)}>
                  {t}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={onClose}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="managers-tree" size={48} />
              </span>
              <p>
                A manager is a staff login: the group masks decide which accounts it sees and the
                permissions decide what it may touch.
              </p>
            </div>
            {tab === "Common" ? (
              <div className="form-grid">
                <label>Login</label>
                <input
                  type="text"
                  readOnly={!isNew}
                  value={draft.login}
                  onChange={(e) => set("login", e.target.value)}
                />
                <label>Name</label>
                <input type="text" value={draft.name} onChange={(e) => set("name", e.target.value)} />
                <label>Mailbox</label>
                <input type="text" value={draft.mailbox} onChange={(e) => set("mailbox", e.target.value)} />
                <label>Groups</label>
                <input
                  type="text"
                  className="wide"
                  placeholder={String.raw`demo\*; real\*  (masks, ; separated)`}
                  value={draft.groups}
                  onChange={(e) => set("groups", e.target.value)}
                />
              </div>
            ) : (
              <div className="form-grid grp-check-stack mgr-rights-grid">
                {rightKeys.map((key) => (
                  <label key={key} className="sym-check">
                    <input
                      type="checkbox"
                      checked={!!rights[key]}
                      onChange={() => setRights({ ...rights, [key]: !rights[key] })}
                    />{" "}
                    {humanize(key.replace(/^right_/, ""))}
                  </label>
                ))}
                {isNew && !rightKeys.length && (
                  <p className="module-note">Rights are granted after the manager is created.</p>
                )}
              </div>
            )}
          </div>
        </SettingsDialog>
      </div>
    </div>
  );
}

/** Staff logins with their group masks. */
export function ManagersModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const [view, setView] = useState({ grid: true, autoArrange: true });
  const canEdit = session.can?.right_cfg_managers !== false;

  const load = () => fetchManagers().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

  async function onDelete(row) {
    if (!window.confirm(`Delete manager ${row.login} '${row.name}'?`)) return;
    const res = await deleteManager(row.login);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  function saved() {
    load();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className={`data-table${view.grid ? " data-table-grid" : ""}${view.autoArrange ? " data-table-auto" : ""}`}>
          <thead>
            <tr>
              <th>Login</th>
              <th>Name</th>
              <th>Mailbox</th>
              <th>Groups</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.login}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ manager: row })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="manager" />
                    {row.login}
                  </span>
                </td>
                <td>{row.name}</td>
                <td>{row.mailbox || "—"}</td>
                <td>{(row.groups || []).join("; ")}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr>
                <td colSpan={4} className="df-empty">No managers</td>
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
            { label: "Servers", disabled: true },
            "sep",
            { label: "Add", icon: "add", shortcut: "Ctrl+N", onClick: () => setDialog({ manager: null }) },
            { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: selected == null, onClick: () => setDialog({ manager: rows[selected] }) },
            { label: "Delete", icon: "delete", shortcut: "Ctrl+D", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
            "sep",
            { label: "Move Up", disabled: true },
            { label: "Move Down", disabled: true },
            { label: "Sort by Login", disabled: true },
            "sep",
            { label: "Export to File", disabled: true },
            { label: "Import from File", disabled: true },
            "sep",
            { label: "Find", shortcut: "Ctrl+F", disabled: true },
            "sep",
            { label: "Auto Arrange", checked: view.autoArrange, onClick: () => setView((v) => ({ ...v, autoArrange: !v.autoArrange })) },
            { label: "Grid", checked: view.grid, onClick: () => setView((v) => ({ ...v, grid: !v.grid })) },
          ]}
        />
      )}
      {dialog && (
        <ManagerDialog manager={dialog.manager} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}

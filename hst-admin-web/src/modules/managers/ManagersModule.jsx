import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { GroupTreeSelect } from "@/components/ui/GroupTreeSelect.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createManager,
  deleteManager,
  fetchManager,
  fetchManagerRights,
  fetchManagers,
  updateManager,
} from "@/api/endpoints/managers.js";
import { fetchUser } from "@/api/endpoints/users.js";
import {
  ManagerLimit_options,
  ManagerRightDeps,
  ManagerRightsTree,
  ManagerRoleTemplates,
  treeRightKeys,
} from "@/constants/managerRights.js";

const ROLES_KEY = "hst_manager_roles";

const loadRoles = () => {
  try {
    return JSON.parse(localStorage.getItem(ROLES_KEY)) || {};
  } catch {
    return {};
  }
};

// A branch of the permissions tree: a section checkbox drives its children, dependents follow.
function RightsBranch({ node, depth, rights, toggle, setMany }) {
  const [open, setOpen] = useState(depth === 0);

  if (!node.children) {
    return (
      <div className="mgr-tree-row" style={{ paddingLeft: depth * 18 + 16 }}>
        <label className="sym-check">
          <input type="checkbox" checked={!!rights[node.key]} onChange={() => toggle(node.key)} />{" "}
          {node.label}
        </label>
      </div>
    );
  }

  const keys = treeRightKeys(node.children);
  const on = keys.filter((k) => rights[k]).length;

  return (
    <>
      <div className="mgr-tree-row" style={{ paddingLeft: depth * 18 }}>
        <button type="button" className="mgr-tree-arrow" onClick={() => setOpen(!open)}>
          {open ? "▾" : "▸"}
        </button>
        <label className="sym-check">
          <input
            type="checkbox"
            checked={on === keys.length}
            ref={(el) => el && (el.indeterminate = on > 0 && on < keys.length)}
            onChange={() => setMany(keys, on !== keys.length)}
          />{" "}
          {node.label}
        </label>
      </div>
      {open &&
        node.children.map((child) => (
          <RightsBranch
            key={child.key}
            node={child}
            depth={depth + 1}
            rights={rights}
            toggle={toggle}
            setMany={setMany}
          />
        ))}
    </>
  );
}

function ManagerDialog({ manager, onClose, onSaved }) {
  const isNew = !manager;
  const [tab, setTab] = useState("Common");
  const [draft, setDraft] = useState({
    login: manager?.login ?? "",
    name: manager?.name ?? "",
    mailbox: manager?.mailbox ?? "",
    limitLogs: 0,
    limitReports: 0,
  });
  const [groups, setGroups] = useState(manager?.groups ?? ["*"]);
  const [groupSel, setGroupSel] = useState(null);
  const [groupEdit, setGroupEdit] = useState(null);
  const [rights, setRights] = useState({});
  const [roles, setRoles] = useState(loadRoles);
  const [role, setRole] = useState("");
  const [roleNaming, setRoleNaming] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(manager?.login ?? "new");
  const close = useDialogStack(onClose);

  useEffect(() => {
    if (isNew) return;
    fetchManagerRights(manager.login).then((res) => res.ok && setRights(res.data.rights || {}));
    fetchManager(manager.login).then((res) => {
      if (!res.ok) return;
      setDraft((prev) => ({
        ...prev,
        mailbox: res.data.mailbox ?? prev.mailbox,
        limitLogs: res.data.request_limit_logs ?? 0,
        limitReports: res.data.request_limit_reports ?? 0,
      }));
    });
  }, [manager, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));
  const toggle = (key) =>
    setRights((prev) => {
      const next = { ...prev, [key]: !prev[key] };
      // stripping a base right strips whatever leans on it
      if (prev[key]) {
        for (const [dep, base] of Object.entries(ManagerRightDeps)) {
          if (base === key && next[dep]) next[dep] = false;
        }
      }
      return next;
    });
  const setMany = (keys, on) =>
    setRights((prev) => {
      const next = { ...prev };
      keys.forEach((k) => (next[k] = on));
      return next;
    });

  function commitGroupRow(value) {
    const v = value.trim();
    setGroups((prev) => {
      const next = [...prev];
      if (groupEdit.index === prev.length) {
        if (v) next.push(v);
      } else if (v) {
        next[groupEdit.index] = v;
      }
      return next;
    });
    setGroupEdit(null);
  }

  async function handleOk() {
    if (isNew && !Number(draft.login)) {
      setError("A manager is an existing account: give its login");
      return;
    }
    let name = draft.name;
    if (isNew) {
      const user = await fetchUser(Number(draft.login));
      if (!user.ok) {
        setError("No account with that login");
        return;
      }
      name = user.data.name || String(draft.login);
    }
    const body = {
      name,
      mailbox: draft.mailbox,
      request_limit_logs: Number(draft.limitLogs),
      request_limit_reports: Number(draft.limitReports),
      groups,
      rights: Object.entries(rights).filter(([, on]) => on).map(([key]) => key),
    };
    const res = isNew
      ? await createManager({ ...body, login: Number(draft.login) })
      : await updateManager(manager.login, body);
    if (!res.ok) {
      setError(res.message || "save failed");
      return;
    }
    onSaved();
    close();
  }

  function saveRole(name) {
    const v = name.trim();
    if (!v || ManagerRoleTemplates[v]) return setRoleNaming(null);
    const next = {
      ...roles,
      [v]: Object.entries(rights).filter(([, on]) => on).map(([key]) => key),
    };
    localStorage.setItem(ROLES_KEY, JSON.stringify(next));
    setRoles(next);
    setRole(v);
    setRoleNaming(null);
  }

  function applyRole(name) {
    setRole(name);
    const template = ManagerRoleTemplates[name];
    const keys = template === "*" ? treeRightKeys() : template || roles[name];
    if (!keys) return;
    const next = {};
    keys.forEach((k) => (next[k] = true));
    setRights(next);
  }

  function deleteRole() {
    if (!role || ManagerRoleTemplates[role]) return;
    const next = { ...roles };
    delete next[role];
    localStorage.setItem(ROLES_KEY, JSON.stringify(next));
    setRoles(next);
    setRole("");
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          width={620}
          height={560}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "Manager: New" : `Manager: ${manager.name || manager.login}`}
          tabs={
            <div className="config-tabs">
              {["Common", "Permissions", "Reports", "IP Access List"].map((t) => (
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
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            {tab === "Common" && (
              <>
                <div className="sym-sessions-intro">
                  <span className="sym-tab-intro-icon" aria-hidden="true">
                    <Icon id="managers-tree" size={48} />
                  </span>
                  <p>
                    Managers control platform settings, accounts and trading operations. To add a
                    manager, create a trading account in the 'managers' or 'administrators' group
                    and specify its login in this tab. Then configure access to client groups.
                  </p>
                </div>
                <div className="mgr-common-grid">
                  <label>Login:</label>
                  <input
                    type="text"
                    readOnly={!isNew}
                    value={draft.login}
                    onChange={(e) => set("login", e.target.value)}
                  />
                  <label>Mailbox name:</label>
                  <input
                    type="text"
                    value={draft.mailbox}
                    onChange={(e) => set("mailbox", e.target.value)}
                  />
                </div>
                <div className="mgr-groups-section">
                  <div className="mgr-groups-left">
                    <span className="routing-conds-caption">Groups:</span>
                    <div className="routing-conds-btns">
                      <button
                        type="button"
                        onClick={() => {
                          setGroupEdit({ index: groups.length, value: "" });
                          setGroupSel(null);
                        }}
                      >
                        Add
                      </button>
                      <button
                        type="button"
                        disabled={groupSel == null}
                        onClick={() => setGroupEdit({ index: groupSel, value: groups[groupSel] })}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        disabled={groupSel == null}
                        onClick={() => {
                          setGroups(groups.filter((_, i) => i !== groupSel));
                          setGroupSel(null);
                        }}
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                  <div className="routing-conds-box mgr-groups-box">
                    <table>
                      <tbody>
                        {groups.map((mask, i) =>
                          groupEdit?.index === i ? (
                            <tr key={i}>
                              <td onClick={(e) => e.stopPropagation()}>
                                <GroupTreeSelect maskable value={groupEdit.value} onChange={(v) => setGroupEdit({ index: i, value: v })} />
                                <button type="button" className="mgr-row-ok" onClick={() => commitGroupRow(groupEdit.value)}>OK</button>
                              </td>
                            </tr>
                          ) : (
                            <tr
                              key={i}
                              className={groupSel === i ? "selected" : ""}
                              onClick={() => setGroupSel(i)}
                              onDoubleClick={() => setGroupEdit({ index: i, value: mask })}
                            >
                              <td>
                                <span className="sym-symbol-cell">
                                  <Icon id="groups" size={14} />
                                  {mask}
                                </span>
                              </td>
                            </tr>
                          ),
                        )}
                        {groupEdit?.index === groups.length && (
                          <tr>
                            <td onClick={(e) => e.stopPropagation()}>
                              <GroupTreeSelect maskable value={groupEdit.value} onChange={(v) => setGroupEdit({ index: groups.length, value: v })} />
                              <button type="button" className="mgr-row-ok" onClick={() => commitGroupRow(groupEdit.value)}>OK</button>
                            </td>
                          </tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            )}
            {tab === "Permissions" && (
              <>
                <div className="sym-sessions-intro">
                  <span className="sym-tab-intro-icon" aria-hidden="true">
                    <Icon id="managers-tree" size={48} />
                  </span>
                  <p>Set up permissions of the manager by selecting entries or a previously saved set.</p>
                </div>
                <div className="mgr-common-grid">
                  <label>Available logs:</label>
                  <PropSelect
                    value={Number(draft.limitLogs)}
                    options={ManagerLimit_options}
                    onChange={(v) => set("limitLogs", v)}
                  />
                  <label>Available reports:</label>
                  <PropSelect
                    value={Number(draft.limitReports)}
                    options={ManagerLimit_options}
                    onChange={(v) => set("limitReports", v)}
                  />
                </div>
                <div className="mgr-role-row">
                  <label>Role:</label>
                  {roleNaming == null ? (
                    <PropSelect
                      value={role}
                      options={["", ...Object.keys(ManagerRoleTemplates), ...Object.keys(roles).filter((n) => !ManagerRoleTemplates[n])]}
                      onChange={applyRole}
                    />
                  ) : (
                    <input
                      autoFocus
                      type="text"
                      placeholder="role name"
                      value={roleNaming}
                      onChange={(e) => setRoleNaming(e.target.value)}
                      onKeyDown={(e) => e.key === "Enter" && saveRole(roleNaming)}
                    />
                  )}
                  <button type="button" onClick={() => (roleNaming == null ? setRoleNaming("") : saveRole(roleNaming))}>
                    Save As
                  </button>
                  <button type="button" disabled={!role || !!ManagerRoleTemplates[role]} onClick={deleteRole}>
                    Delete
                  </button>
                </div>
                <div className="mgr-perm-row">
                  <label>Permissions:</label>
                  <div className="mgr-tree-box">
                    {ManagerRightsTree.map((node) => (
                      <RightsBranch key={node.key} node={node} depth={0} rights={rights} toggle={toggle} setMany={setMany} />
                    ))}
                  </div>
                </div>
              </>
            )}
            {(tab === "Reports" || tab === "IP Access List") && (
              <p className="module-note">This tab is configured later.</p>
            )}
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
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

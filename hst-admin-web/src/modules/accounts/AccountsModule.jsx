import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { useRegisterToolbarActions } from "@/hooks/useToolbarActions.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { deleteUser, fetchUsers } from "@/api/endpoints/users.js";
import { formatNs } from "@/lib/time.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { AccountDialog } from "./AccountDialog.jsx";
import { BalanceDialog } from "./BalanceDialog.jsx";

const money = (v) => (v ?? 0).toFixed(2);

/** Trading accounts list; double-click opens the account dialog. */
export function AccountsModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const [balance, setBalance] = useState(null);
  const [view, setView] = useState({ grid: true, autoArrange: true });
  const canEdit = session.can?.right_acc_manager !== false;

  const load = () => fetchUsers().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

  useRegisterToolbarActions({
    onAdd: canEdit ? () => setDialog({ login: "new" }) : undefined,
    onEdit: canEdit && selected != null ? () => setDialog({ login: rows[selected]?.login }) : undefined,
    onDelete: canEdit && selected != null ? () => onDelete(rows[selected]) : undefined,
    canAdd: canEdit,
    canEdit: canEdit && selected != null,
    canDelete: canEdit && selected != null,
  });

  async function onDelete(row) {
    if (!window.confirm(`Delete account ${row.login} '${row.name}'?`)) return;
    const res = await deleteUser(row.login);
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
              <th>Group</th>
              <th>Leverage</th>
              <th>Balance</th>
              <th>Credit</th>
              <th>Email</th>
              <th>Last access</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.login}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ login: row.login })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id={row.is_manager ? "manager" : "account"} />
                    {row.login}
                  </span>
                </td>
                <td>{row.name}</td>
                <td>{row.group}</td>
                <td>1 : {row.leverage}</td>
                <td>{money(row.balance)}</td>
                <td>{money(row.credit)}</td>
                <td>{row.email}</td>
                <td>{row.last_access ? formatNs(row.last_access) : "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {menu && canEdit && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            { label: "New", icon: "add", shortcut: "Ctrl+N", onClick: () => setDialog({ login: "new" }) },
            { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: selected == null, onClick: () => setDialog({ login: rows[selected]?.login }) },
            { label: "Edit group", disabled: true },
            { label: "Edit manager", disabled: true },
            { label: "Delete", icon: "delete", shortcut: "Ctrl+D", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
            "sep",
            { label: "Request", disabled: true },
            { label: "Move to Archive", disabled: true },
            {
              label: "Balance",
              items: [
                { label: "Check Balance", disabled: true },
                { label: "Fix Balance", disabled: selected == null, onClick: () => setBalance(rows[selected]) },
              ],
            },
            {
              label: "Copy As",
              items: [
                { label: "Lines", disabled: true },
                { label: "List of Logins", disabled: true },
              ],
            },
            "sep",
            { label: "Export", disabled: true },
            { label: "Import from File", disabled: true },
            { label: "Import from Server", disabled: true },
            "sep",
            { label: "E-Mail", disabled: true },
            { label: "Journal", disabled: true },
            { label: "Find", shortcut: "Ctrl+F", disabled: true },
            "sep",
            { label: "Enabled only", disabled: true },
            { label: "Auto Arrange", checked: view.autoArrange, onClick: () => setView((v) => ({ ...v, autoArrange: !v.autoArrange })) },
            { label: "Grid", checked: view.grid, onClick: () => setView((v) => ({ ...v, grid: !v.grid })) },
          ]}
        />
      )}
      {dialog && (
        <AccountDialog login={dialog.login} onClose={() => setDialog(null)} onSaved={saved} />
      )}
      {balance && (
        <BalanceDialog user={balance} onClose={() => setBalance(null)} onSaved={saved} />
      )}
    </div>
  );
}

import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
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
  const canEdit = session.can?.right_acc_manager !== false;

  const load = () => fetchUsers().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

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
        <table className="data-table data-table-grid data-table-auto">
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
            { label: "Add", onClick: () => setDialog({ login: "new" }) },
            { label: "Edit", disabled: selected == null, onClick: () => setDialog({ login: rows[selected]?.login }) },
            { label: "Balance…", disabled: selected == null, onClick: () => setBalance(rows[selected]) },
            "sep",
            { label: "Delete", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
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

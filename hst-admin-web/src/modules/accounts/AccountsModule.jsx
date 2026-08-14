import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { useRegisterToolbarActions } from "@/hooks/useToolbarActions.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { deleteUser, fetchUsers } from "@/api/endpoints/users.js";
import { checkBalances, fixBalance } from "@/api/endpoints/balance.js";
import { formatNs } from "@/lib/time.js";
import { money } from "@/lib/format.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { useLiveAccounts } from "@/hooks/useLiveAccounts.js";
import { AccountDialog } from "./AccountDialog.jsx";
import { BalanceDialog } from "./BalanceDialog.jsx";
import { BulkBalanceDialog } from "./BulkBalanceDialog.jsx";
import { MailDialog } from "@/modules/mail/MailDialog.jsx";
import { useConfirm } from "@/hooks/useConfirm.jsx";

// Live money cells: a dash until the engine's first summary line for the account arrives.
function LiveMoneyCells({ account }) {
  if (!account) return <><td>—</td><td>—</td><td>—</td><td>—</td><td>—</td></>;

  const profitClass = account.profit < 0 ? "acc-loss" : "acc-profit";

  return (
    <>
      <td>{money(account.equity)}</td>
      <td>{money(account.margin)}</td>
      <td>{money(account.free)}</td>
      <td>{account.margin > 0 ? account.level.toFixed(2) : "—"}</td>
      <td className={profitClass}>{money(account.profit)}</td>
    </>
  );
}

/** Trading accounts list; double-click opens the account dialog. */
export function AccountsModule() {
  const session = useSession();
  // the manager panel watches the money live; the admin panel keeps its config view
  const isManagerPanel = session.terminal !== "administrator";
  const live = useLiveAccounts();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [multi, setMulti] = useState(() => new Set());
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const { confirm, confirmElement } = useConfirm();
  const [balance, setBalance] = useState(null);
  const [bulk, setBulk] = useState(null);
  const [mail, setMail] = useState(null);
  const [checks, setChecks] = useState(null);
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
    if (!(await confirm({ title: "Accounts", message: `Delete account ${row.login} '${row.name}'?` }))) return;
    const res = await deleteUser(row.login);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  // logins the bulk actions target: the multi set, else the single selected row
  const targetLogins = multi.size
    ? [...multi]
    : selected != null && rows?.[selected]
      ? [rows[selected].login]
      : [];

  function onRowClick(e, i, row) {
    if (e.ctrlKey || e.metaKey) {
      setSelected(i);
      setMulti((prev) => {
        const next = new Set(prev);
        next.has(row.login) ? next.delete(row.login) : next.add(row.login);
        return next;
      });
    } else if (e.shiftKey && selected != null) {
      const [a, b] = [Math.min(selected, i), Math.max(selected, i)];
      setMulti(new Set(rows.slice(a, b + 1).map((r) => r.login)));
    } else {
      setSelected(i);
      setMulti(new Set());
    }
  }

  function saved() {
    load();
    session.refreshNav?.();
  }

  // the audit: recompute every account from its deals, flag what does not add up
  async function runCheck() {
    const res = await checkBalances();
    if (!res.ok) {
      window.alert(res.message || "check failed");
      return;
    }
    const byLogin = new Map(res.data.map((r) => [r.login, r]));
    setChecks(byLogin);
    const invalid = res.data.filter((r) => !r.ok).length;
    window.alert(invalid ? `${invalid} of ${res.data.length} accounts have an invalid balance` : `All ${res.data.length} accounts add up`);
  }

  async function runFix(row) {
    const res = await fixBalance(row.login);
    if (!res.ok) {
      window.alert(res.message || "fix failed");
      return;
    }
    setChecks((prev) => {
      const next = new Map(prev ?? []);
      next.set(row.login, res.data);
      return next;
    });
    load();
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
              {isManagerPanel && (
                <>
                  <th>Equity</th>
                  <th>Margin</th>
                  <th>Free</th>
                  <th>Level %</th>
                  <th>Profit</th>
                </>
              )}
              <th>Email</th>
              <th>Last access</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.login}
                className={`${selected === i || multi.has(row.login) ? "selected" : ""}${checks?.get(row.login)?.ok === false ? " acc-invalid" : ""}`.trim()}
                onClick={(e) => onRowClick(e, i, row)}
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
                <td>
                  {checks?.get(row.login)?.ok === false
                    ? `${money(checks.get(row.login).balance)} / ${money(checks.get(row.login).valid_balance)}`
                    : money(live.get(row.login)?.balance ?? row.balance)}
                </td>
                <td>{money(live.get(row.login)?.credit ?? row.credit)}</td>
                {isManagerPanel && <LiveMoneyCells account={live.get(row.login)} />}
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
                { label: "Check Balance", onClick: runCheck },
                { label: "Fix Balance", disabled: selected == null || checks?.get(rows[selected]?.login)?.ok !== false, onClick: () => runFix(rows[selected]) },
                "sep",
                { label: "Balance Operation…", disabled: selected == null, onClick: () => setBalance(rows[selected]) },
                { label: "Bulk Operation…", disabled: !targetLogins.length, onClick: () => setBulk(targetLogins) },
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
            { label: "E-Mail", disabled: !targetLogins.length, onClick: () => setMail(targetLogins) },
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
      {bulk && (
        <BulkBalanceDialog logins={bulk} onClose={() => setBulk(null)} onSaved={saved} />
      )}
      {mail && <MailDialog logins={mail} onClose={() => setMail(null)} />}
      {confirmElement}
    </div>
  );
}

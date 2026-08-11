import { useEffect, useState } from "react";
import { fetchUsers } from "@/api/endpoints/users.js";
import { useGroups } from "@/hooks/useGroups.js";
import { useLiveAccounts } from "@/hooks/useLiveAccounts.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { money } from "@/lib/format.js";
import { AccountDialog } from "@/modules/accounts/AccountDialog.jsx";

/** Accounts whose live margin level fell to their group's call line; red when at stop out. */
export function MarginCallsModule() {
  const [users, setUsers] = useState([]);
  const [dialog, setDialog] = useState(null);
  const { groups } = useGroups();
  const live = useLiveAccounts();

  useEffect(() => {
    fetchUsers().then((res) => res.ok && setUsers(res.data || []));
  }, []);

  const levelsOf = (groupName) => {
    const group = groups.find((g) => g.group === groupName);
    return { call: group?.margin_call ?? 0, stopOut: group?.margin_stop_out ?? 0 };
  };

  // an account is listed while its margin is in use and its level sits at or under the call line
  const rows = users
    .map((user) => ({ user, money: live.get(user.login), ...levelsOf(user.group) }))
    .filter((r) => r.money && r.money.margin > 0 && r.call > 0 && r.money.level <= r.call)
    .sort((a, b) => a.money.level - b.money.level);

  return (
    <div className="module-root">
      <div className="table-wrap">
        <table className="data-table data-table-grid">
          <thead>
            <tr>
              <th>Login</th>
              <th>Name</th>
              <th>Group</th>
              <th>Equity</th>
              <th>Margin</th>
              <th>Level %</th>
            </tr>
          </thead>
          <tbody>
            {rows.map(({ user, money: m, stopOut }) => (
              <tr
                key={user.login}
                className={m.level <= stopOut ? "mc-stopout" : ""}
                onDoubleClick={() => setDialog(user.login)}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="account" />
                    {user.login}
                  </span>
                </td>
                <td>{user.name}</td>
                <td>{user.group}</td>
                <td>{money(m.equity)}</td>
                <td>{money(m.margin)}</td>
                <td>{m.level.toFixed(2)}</td>
              </tr>
            ))}
            {!rows.length && (
              <tr><td colSpan={6} className="df-empty">No accounts under margin call</td></tr>
            )}
          </tbody>
        </table>
      </div>
      {dialog && <AccountDialog login={dialog} onClose={() => setDialog(null)} onSaved={() => {}} />}
    </div>
  );
}

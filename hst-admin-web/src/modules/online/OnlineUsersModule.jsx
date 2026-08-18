import { useEffect, useState } from "react";
import { disconnectSession, fetchOnlineUsers } from "@/api/endpoints/online.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useLiveAccounts } from "@/hooks/useLiveAccounts.js";
import { formatNs } from "@/lib/time.js";
import { money } from "@/lib/format.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";

const REFRESH_EVERY_MS = 10000;

// The terminal behind each connection, named the way the platform names them.
const CLIENT_NAMES = {
  0: "Desktop", 3: "Web API", 4: "iPhone", 5: "Android", 11: "Web Trader",
  32: "Administrator", 33: "Manager", 34: "Manager API", 36: "Admin API", 37: "Manager API Web",
};

/** Every live connection in scope; one login on three devices is three rows. */
export function OnlineUsersModule() {
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [menu, setMenu] = useState(null);
  const { confirm, confirmElement } = useConfirm();
  const live = useLiveAccounts();

  const load = () => fetchOnlineUsers().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
    const timer = setInterval(load, REFRESH_EVERY_MS);
    return () => clearInterval(timer);
  }, []);

  async function disconnect(row) {
    if (!(await confirm({ title: "Online Users", message: `Disconnect this ${CLIENT_NAMES[row.connection_type] ?? "session"} of ${row.login}?` }))) return;
    const res = await disconnectSession(row.session_id);
    if (!res.ok) window.alert(res.message || "disconnect failed");
    load();
  }

  const row = selected != null ? rows?.[selected] : null;

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Login</th>
              <th>Name</th>
              <th>Group</th>
              <th>Client</th>
              <th>OS</th>
              <th>IP</th>
              <th>Connected</th>
              <th>Equity</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((r, i) => (
              <tr
                key={r.connection_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id={r.connection_type >= 32 ? "manager" : "client"} />
                    {r.login}
                  </span>
                </td>
                <td>{r.name}</td>
                <td>{r.group}</td>
                <td>{CLIENT_NAMES[r.connection_type] ?? r.connection_type}</td>
                <td>{r.os}</td>
                <td>{r.ip}</td>
                <td>{formatNs(r.connected_at)}</td>
                <td>{live.get(r.login) ? money(live.get(r.login).equity) : "—"}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr><td colSpan={8} className="df-empty">Nobody is connected</td></tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="filter-bar">
        <span className="grp-suffix">{rows?.length ?? 0} connections</span>
      </div>
      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            { label: "Disconnect", icon: "delete", disabled: !row, onClick: () => disconnect(row) },
            "sep",
            { label: "Refresh", icon: "refresh", onClick: load },
          ]}
        />
      )}
      {confirmElement}
    </div>
  );
}

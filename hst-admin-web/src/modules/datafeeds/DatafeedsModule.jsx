import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useDatafeeds } from "@/hooks/useDatafeeds.js";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { activateDatafeed, deleteDatafeed } from "@/api/endpoints/datafeeds.js";
import { formatNs } from "@/lib/time.js";
import { DatafeedDialog } from "./DatafeedDialog.jsx";

const sourceLabel = (row) => ((row.mode ?? 1) & 2 ? "N" : "Q");
const stateLabel = (row) => {
  if (row.news_count && !row.ticks_count) return String(row.news_count);
  if (row.books_count) return `${row.ticks_count} / ${row.books_count}`;
  return String(row.ticks_count ?? 0);
};
const statusLabel = (row) =>
  row.enable !== 1 ? "Disabled" : row.sys_connection === 1 ? "Online" : "Offline";

/** Data feeds list: priority order, live state, dialog on double-click. */
export function DatafeedsModule() {
  const { datafeeds, loading, reload } = useDatafeeds();
  const session = useSession();
  const [params] = useSearchParams();
  const highlight = Number(params.get("feed")) || null;
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_datafeeds !== false;

  async function onDelete(row) {
    if (!window.confirm(`Delete data feed '${row.name}'?`)) return;
    const res = await deleteDatafeed(row.datafeed_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  async function onRestart(row) {
    const res = await activateDatafeed(row.datafeed_id);
    if (!res.ok) window.alert(res.message || "restart failed");
    reload();
  }

  function saved() {
    reload();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table df-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Name</th>
              <th>Source</th>
              <th>Server</th>
              <th>Symbols</th>
              <th>Last active</th>
              <th>State</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {datafeeds.map((row, i) => (
              <tr
                key={row.datafeed_id}
                className={selected === i || row.datafeed_id === highlight ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: row.datafeed_id })}
              >
                <td>{row.name}</td>
                <td>{sourceLabel(row)}</td>
                <td>{row.feed_server || "—"}</td>
                <td>*</td>
                <td>{row.sys_last_time ? formatNs(row.sys_last_time) : "—"}</td>
                <td>{stateLabel(row)}</td>
                <td className={row.sys_connection === 1 && row.enable === 1 ? "" : "nav-feed-disabled"}>
                  {statusLabel(row)}
                </td>
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
            { label: "Add", onClick: () => setDialog({ id: "new" }) },
            { label: "Edit", disabled: selected == null, onClick: () => setDialog({ id: datafeeds[selected]?.datafeed_id }) },
            { label: "Restart", disabled: selected == null, onClick: () => onRestart(datafeeds[selected]) },
            "sep",
            { label: "Delete", disabled: selected == null, onClick: () => onDelete(datafeeds[selected]) },
          ]}
        />
      )}
      {dialog && (
        <DatafeedDialog feedId={dialog.id} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}

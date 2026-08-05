import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useDatafeeds } from "@/hooks/useDatafeeds.js";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import {
  activateDatafeed,
  deleteDatafeed,
  fetchDatafeedModules,
  updateDatafeed,
} from "@/api/endpoints/datafeeds.js";
import { formatNs } from "@/lib/time.js";
import { DatafeedDialog } from "./DatafeedDialog.jsx";

const sourceLabel = (row) => {
  const mode = row.mode ?? 1;
  if (mode & 1 && mode & 2) return "N + Q";
  return mode & 2 ? "N" : "Q";
};
const stateLabel = (row) => {
  if (row.news_count && !row.ticks_count) return String(row.news_count);
  if (row.books_count) return `${row.ticks_count} / ${row.books_count}`;
  return String(row.ticks_count ?? 0);
};
const statusLabel = (row) =>
  row.enable !== 1 ? "Disabled" : row.sys_connection === 1 ? "Online" : "Offline";
// the reference writes dates with dot separators
const lastActive = (ns) => (ns ? formatNs(ns).replace(/-/g, ".") : "");

/** Data feeds list: priority order, live state, dialog on double-click. */
export function DatafeedsModule() {
  const { datafeeds, loading, reload } = useDatafeeds();
  const session = useSession();
  const [params] = useSearchParams();
  const highlight = Number(params.get("feed")) || null;
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const [pane, setPane] = useState("selected");
  const [modules, setModules] = useState([]);
  const canEdit = session.can?.right_cfg_datafeeds !== false;
  const row = selected == null ? null : datafeeds[selected];

  useEffect(() => {
    if (pane !== "available" || modules.length) return;
    Promise.all([fetchDatafeedModules(1), fetchDatafeedModules(2)]).then(([q, n]) => {
      const merged = new Map();
      for (const opt of [...(q.data || []), ...(n.data || [])]) merged.set(opt.module, opt);
      setModules([...merged.values()]);
    });
  }, [pane, modules.length]);

  async function onDelete(target) {
    if (!window.confirm(`Delete data feed '${target.name}'?`)) return;
    const res = await deleteDatafeed(target.datafeed_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  async function onRestart(target) {
    const res = await activateDatafeed(target.datafeed_id);
    if (!res.ok) window.alert(res.message || "restart failed");
    reload();
  }

  async function onToggleEnable(target) {
    const res = await updateDatafeed(target.datafeed_id, { enable: target.enable === 1 ? 0 : 1 });
    if (!res.ok) window.alert(res.message || "save failed");
    saved();
  }

  function saved() {
    reload();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="df-section-head">
        <div>
          <h2 className="df-section-title">Data Feeds</h2>
          <p className="df-section-sub">Data feeds deliver quotes and news from various providers</p>
        </div>
        <div className="df-view-toggle">
          <button
            type="button"
            className={pane === "available" ? "active" : ""}
            title="Available"
            onClick={() => setPane("available")}
          >
            ▦
          </button>
          <button
            type="button"
            className={pane === "selected" ? "active" : ""}
            title="Selected"
            onClick={() => setPane("selected")}
          >
            ☰
          </button>
        </div>
      </div>
      <div
        className="table-wrap"
        onContextMenu={(e) => {
          e.preventDefault();
          setMenu({ x: e.clientX, y: e.clientY });
        }}
      >
        {pane === "selected" ? (
          <table className="data-table df-table data-table-grid data-table-auto">
            <thead>
              <tr>
                <th>Name</th>
                <th>Source</th>
                <th>Server</th>
                <th className="num">Symbols</th>
                <th className="num">Last active</th>
                <th className="num">Status</th>
                <th className="num">State</th>
              </tr>
            </thead>
            <tbody>
              {datafeeds.map((feed, i) => (
                <tr
                  key={feed.datafeed_id}
                  className={selected === i || feed.datafeed_id === highlight ? "selected" : ""}
                  onClick={() => setSelected(i)}
                  onContextMenu={() => setSelected(i)}
                  onDoubleClick={() => canEdit && setDialog({ id: feed.datafeed_id })}
                >
                  <td>
                    <span className="df-row-glyph">
                      <Icon id="datafeeds" size={14} />
                    </span>
                    {feed.name}
                  </td>
                  <td>{sourceLabel(feed)}</td>
                  <td>{feed.feed_server}</td>
                  <td className="num">*</td>
                  <td className="num">{lastActive(feed.sys_last_time)}</td>
                  <td
                    className={
                      feed.sys_connection === 1 && feed.enable === 1
                        ? "num"
                        : "num nav-feed-disabled"
                    }
                  >
                    {statusLabel(feed)}
                  </td>
                  <td className="num">{stateLabel(feed)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <table className="data-table df-table data-table-grid data-table-auto">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody>
              {modules.map((m) => (
                <tr key={m.module}>
                  <td>
                    <span className="df-row-glyph">
                      <Icon id="datafeeds" size={14} />
                    </span>
                    {m.label || m.module}
                  </td>
                  <td>{m.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
      <div className="df-pane-tabs">
        <button
          type="button"
          className={pane === "selected" ? "active" : ""}
          onClick={() => setPane("selected")}
        >
          Selected
        </button>
        <button
          type="button"
          className={pane === "available" ? "active" : ""}
          onClick={() => setPane("available")}
        >
          Available
        </button>
      </div>
      {menu && canEdit && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            ...listMenuHead({
              onAdd: () => setDialog({ id: "new" }),
              onEdit: () => setDialog({ id: row?.datafeed_id }),
              onDelete: () => onDelete(row),
              hasSelection: !!row,
            }),
            "sep",
            {
              label: row?.enable === 1 ? "Disable" : "Enable",
              disabled: !row,
              onClick: () => onToggleEnable(row),
            },
            { label: "Restart", disabled: !row, onClick: () => onRestart(row) },
            ...listMenuTail({}),
          ]}
        />
      )}
      {loading && !datafeeds.length && <div className="df-empty">Loading…</div>}
      {dialog && (
        <DatafeedDialog feedId={dialog.id} onClose={() => setDialog(null)} onSaved={saved} />
      )}
    </div>
  );
}

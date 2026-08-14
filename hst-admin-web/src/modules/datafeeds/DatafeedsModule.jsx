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
  reorderDatafeeds,
  resolveDatafeedSymbols,
  updateDatafeed,
} from "@/api/endpoints/datafeeds.js";
import { formatNs } from "@/lib/time.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";
import { DatafeedDialog } from "./DatafeedDialog.jsx";
import { DatafeedStatusPage } from "./DatafeedStatusPage.jsx";

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

const symbolCountLabel = (counts, feed) => {
  const id = feed.datafeed_id;
  if (counts[id] === undefined) return "…";
  return String(counts[id]);
};

/** Data feeds list: priority order, live state, dialog on double-click. */
export function DatafeedsModule() {
  const { datafeeds, loading, reload } = useDatafeeds();
  const session = useSession();
  const [params] = useSearchParams();
  const highlight = Number(params.get("feed")) || null;
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const { confirm, confirmElement } = useConfirm();
  const [pane, setPane] = useState("selected");
  const [modules, setModules] = useState([]);
  const [symbolCounts, setSymbolCounts] = useState({});
  const canEdit = session.can?.right_cfg_datafeeds !== false;
  // feed_index is the priority order the server restarts feeds in
  const feeds = [...datafeeds].sort((a, b) => (a.feed_index ?? 0) - (b.feed_index ?? 0));
  const row = selected == null ? null : feeds[selected];

  useEffect(() => {
    if (pane !== "available" || modules.length) return;
    Promise.all([fetchDatafeedModules(1), fetchDatafeedModules(2)]).then(([q, n]) => {
      const merged = new Map();
      for (const opt of [...(q.data || []), ...(n.data || [])]) merged.set(opt.module, opt);
      setModules([...merged.values()]);
    });
  }, [pane, modules.length]);

  useEffect(() => {
    if (!datafeeds.length) return;
    let cancelled = false;
    Promise.all(
      datafeeds.map(async (feed) => {
        const res = await resolveDatafeedSymbols(feed.datafeed_id);
        if (!res.ok) return [feed.datafeed_id, 0];
        return [feed.datafeed_id, res.data?.count ?? 0];
      })
    ).then((rows) => {
      if (cancelled) return;
      setSymbolCounts(Object.fromEntries(rows));
    });
    return () => {
      cancelled = true;
    };
  }, [datafeeds]);

  async function onDelete(target) {
    if (!(await confirm({ title: "Data Feeds", message: `Delete data feed '${target.name}'?` }))) return;
    const res = await deleteDatafeed(target.datafeed_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  async function onRestart(target) {
    const res = await activateDatafeed(target.datafeed_id);
    if (!res.ok) window.alert(res.message || "restart failed");
    reload();
  }

  async function onMove(delta) {
    const to = selected + delta;
    if (selected == null || to < 0 || to >= feeds.length) return;
    const ids = feeds.map((f) => f.datafeed_id);
    [ids[selected], ids[to]] = [ids[to], ids[selected]];
    const res = await reorderDatafeeds(ids);
    if (!res.ok) {
      window.alert(res.message || "reorder failed");
      return;
    }
    setSelected(to);
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

  const statusFeed = highlight == null ? null : feeds.find((f) => f.datafeed_id === highlight);
  // picking the feed in the navigator opens its status page, as the reference does
  if (statusFeed) {
    return (
      <div className="module-root">
        <DatafeedStatusPage feed={statusFeed} canEdit={canEdit} onSaved={saved} />
      </div>
    );
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
              {feeds.map((feed, i) => (
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
                  <td className="num">{symbolCountLabel(symbolCounts, feed)}</td>
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
              label: "Move Up",
              disabled: !row || selected === 0,
              onClick: () => onMove(-1),
            },
            {
              label: "Move Down",
              disabled: !row || selected >= feeds.length - 1,
              onClick: () => onMove(1),
            },
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
      {confirmElement}
    </div>
  );
}

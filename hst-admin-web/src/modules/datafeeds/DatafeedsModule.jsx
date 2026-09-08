import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useDatafeeds } from "@/hooks/useDatafeeds.js";
import { useSession } from "@/hooks/useSession.js";
import { useToolbox } from "@/hooks/useToolbox.jsx";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import {
  activateDatafeed,
  createDatafeed,
  deleteDatafeed,
  fetchDatafeed,
  fetchDatafeedModules,
  reorderDatafeeds,
  resolveDatafeedSymbols,
  updateDatafeed,
} from "@/api/endpoints/datafeeds.js";
import { datafeedCreateBody, normalizeDatafeedDraft } from "./datafeedPayload.js";
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
  const [params, setParams] = useSearchParams();
  const highlight = Number(params.get("feed")) || null;
  const editFromNav = params.get("edit") === "1";
  const addFromNav = params.get("add") === "1";
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const { confirm, confirmElement } = useConfirm();
  const [pane, setPane] = useState("selected");
  const [pickedModule, setPickedModule] = useState(null);
  const importRef = useRef(null);
  const toolbox = useToolbox();
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

  // the reference sorts on the server, so the order survives every terminal
  async function onSort() {
    const ids = [...feeds].sort((a, b) => a.name.localeCompare(b.name)).map((f) => f.datafeed_id);
    const res = await reorderDatafeeds(ids);
    if (!res.ok) window.alert(res.message || "sort failed");
    reload();
  }

  function onFind() {
    const text = window.prompt("Find data feed", "");
    if (!text) return;
    const i = feeds.findIndex((f) => f.name.toLowerCase().includes(text.trim().toLowerCase()));
    if (i < 0) window.alert(`No data feed matches '${text}'`);
    else setSelected(i);
  }

  async function onExport(target) {
    const res = await fetchDatafeed(target.datafeed_id);
    if (!res.ok) return window.alert(res.message || "export failed");
    const a = document.createElement("a");
    a.href = URL.createObjectURL(new Blob([JSON.stringify(res.data, null, 2)], { type: "application/json" }));
    a.download = `${target.name}.json`;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  // ponytail: imports the feed record only; symbols, translations and parameters are set in the dialog
  async function onImport(file) {
    if (!file) return;
    let body;
    try {
      body = JSON.parse(await file.text());
    } catch {
      return window.alert("Not a data feed file");
    }
    const res = await createDatafeed(datafeedCreateBody(normalizeDatafeedDraft({ ...body, name: body.name || file.name.replace(/\.json$/i, "") })));
    if (!res.ok) return window.alert(res.message || "import failed");
    saved();
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
        {editFromNav && canEdit && (
          <DatafeedDialog feedId={statusFeed.datafeed_id} onClose={() => setParams({ feed: String(statusFeed.datafeed_id) })} onSaved={saved} />
        )}
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
                <tr
                  key={m.module}
                  className={pickedModule === m.module ? "selected" : ""}
                  onClick={() => setPickedModule(m.module)}
                  onContextMenu={() => setPickedModule(m.module)}
                  onDoubleClick={() => canEdit && setDialog({ id: "new", module: m.module })}
                >
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
      {menu && canEdit && pane === "available" && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            { label: "Add", icon: "add", shortcut: "Ctrl+N", onClick: () => setDialog({ id: "new", module: pickedModule || "" }) },
          ]}
        />
      )}
      {menu && canEdit && pane === "selected" && (
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
            ...listMenuTail({
              on: {
                moveUp: row && selected > 0 ? () => onMove(-1) : undefined,
                moveDown: row && selected < feeds.length - 1 ? () => onMove(1) : undefined,
                sort: feeds.length > 1 ? onSort : undefined,
                exportFile: row ? () => onExport(row) : undefined,
                importFile: () => importRef.current?.click(),
                journal: row ? () => toolbox?.openJournal?.(row.name) : undefined,
                find: feeds.length ? onFind : undefined,
              },
              extras: [
                "sep",
                {
                  label: row?.enable === 1 ? "Disable" : "Enable",
                  disabled: !row,
                  onClick: () => onToggleEnable(row),
                },
                { label: "Restart", disabled: !row, onClick: () => onRestart(row) },
              ],
            }),
          ]}
        />
      )}
      <input
        ref={importRef}
        type="file"
        accept="application/json,.json"
        hidden
        onChange={(e) => {
          onImport(e.target.files?.[0]);
          e.target.value = "";
        }}
      />
      {loading && !datafeeds.length && <div className="df-empty">Loading…</div>}
      {dialog && (
        <DatafeedDialog feedId={dialog.id} module={dialog.module} onClose={() => setDialog(null)} onSaved={saved} />
      )}
      {addFromNav && canEdit && (
        <DatafeedDialog feedId="new" onClose={() => setParams({})} onSaved={saved} />
      )}
      {confirmElement}
    </div>
  );
}

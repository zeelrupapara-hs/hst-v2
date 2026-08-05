import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useOutletContext, useParams } from "react-router-dom";
import { ApiBar, SearchField } from "../ui/index.js";
import { ContextMenu, useContextMenu } from "../ui/ContextMenu.jsx";
import { ModuleToolbar } from "../data/index.js";
import { DatafeedTable } from "./DatafeedTable.jsx";
import { DatafeedSettingsModule } from "./DatafeedSettingsModule.jsx";
import { useListData, useRowSelection } from "../../hooks/index.js";
import { buildDatafeedTableMenu } from "../../lib/moduleContextMenus.js";
import {
  createDatafeed,
  deleteDatafeeds,
  moveDatafeed,
  saveDatafeed,
  setDatafeedEnable,
} from "../../lib/data.js";

const DEFAULT_VIEW = {
  autoArrange: true,
  grid: true,
  hiddenCols: {},
};

export function DatafeedsModule({ config, panel = "admin", moduleId = "datafeeds" }) {
  const { recordId } = useParams();
  const navigate = useNavigate();
  const { refreshNav } = useOutletContext() || {};
  const [search, setSearch] = useState("");
  const [view, setView] = useState(DEFAULT_VIEW);
  const [columnSort, setColumnSort] = useState(false);
  const [toast, setToast] = useState(null);
  const fileInputRef = useRef(null);
  const { menu, show, close } = useContextMenu();

  const { rows, apiState, reload } = useListData(config, search);
  const { selected, toggleSelect, resetSelection, selectedRows, selectRow } =
    useRowSelection(rows);

  useEffect(() => {
    if (recordId === "new") return;
    if (recordId && rows.length) {
      const idx = rows.findIndex(
        (r) => String(r[config.idKey]) === String(recordId)
      );
      if (idx >= 0) {
        selectRow(idx);
        return;
      }
    }
    resetSelection(rows.length);
  }, [rows, rows.length, recordId, config.idKey, selectRow, resetSelection]);

  const showToast = useCallback((msg) => {
    setToast(msg);
    setTimeout(() => setToast(null), 2800);
  }, []);

  const refreshListAndNav = useCallback(async () => {
    await reload();
    refreshNav?.();
  }, [reload, refreshNav]);

  function listPath() {
    return `/${panel}/${moduleId}`;
  }

  function openDetail(row) {
    if (!row) return;
    navigate(`/${panel}/${moduleId}/${encodeURIComponent(row[config.idKey])}`);
  }

  function closeSettings() {
    navigate(listPath());
  }

  function handleAdd() {
    navigate(`/${panel}/${moduleId}/new`);
  }

  async function handleDelete() {
    if (!selectedRows.length) return;
    const names = selectedRows.map((r) => r.name).join(", ");
    if (!window.confirm(`Delete data feed(s): ${names}?`)) return;
    await deleteDatafeeds(selectedRows.map((r) => r.datafeed_id));
    await refreshListAndNav();
    if (recordId) closeSettings();
    showToast(`Deleted ${selectedRows.length} data feed(s)`);
  }

  async function handleMove(direction) {
    const row = selectedRows[0];
    if (!row) return;
    const ok = await moveDatafeed(row.datafeed_id, direction);
    if (ok.ok) {
      await refreshListAndNav();
      showToast(`Moved ${row.name} ${direction}`);
    }
  }

  async function handleEnable(enable) {
    for (const row of selectedRows) {
      await setDatafeedEnable(row.datafeed_id, enable);
    }
    await refreshListAndNav();
    showToast(`${enable ? "Enabled" : "Disabled"} ${selectedRows.length} feed(s)`);
  }

  function handleExportFile() {
    if (!selectedRows.length) return;
    const blob = new Blob([JSON.stringify(selectedRows, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "datafeeds-export.json";
    a.click();
    URL.revokeObjectURL(url);
    showToast(`Exported ${selectedRows.length} data feed(s)`);
  }

  function handleImportFile() {
    fileInputRef.current?.click();
  }

  async function handleImportChange(e) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    try {
      const list = JSON.parse(await file.text());
      const items = Array.isArray(list) ? list : [list];
      for (const item of items) {
        if (item.datafeed_id) {
          await saveDatafeed(item.datafeed_id, item);
        } else if (item.name) {
          await createDatafeed(item);
        }
      }
      await refreshListAndNav();
      showToast(`Imported ${items.length} data feed(s)`);
    } catch {
      showToast("Import failed — invalid JSON");
    }
  }

  const menuHandlers = {
    add: handleAdd,
    edit: () => openDetail(selectedRows[selectedRows.length - 1]),
    delete: handleDelete,
    moveUp: () => handleMove("up"),
    moveDown: () => handleMove("down"),
    sort: () => showToast("Sort Alphabetically — server-side sort not implemented"),
    enable: () => handleEnable(true),
    disable: () => handleEnable(false),
    exportFile: handleExportFile,
    importFile: handleImportFile,
    journal: () => showToast("Journal — open Toolbox › Journal for this feed"),
    find: () => document.querySelector("[data-df-search]")?.focus(),
    toggleAutoArrange: () =>
      setView((v) => ({ ...v, autoArrange: !v.autoArrange })),
    toggleGrid: () => setView((v) => ({ ...v, grid: !v.grid })),
    toggleColumn: (ctx) => {
      const colId = ctx.meta?.columnId;
      if (!colId) return;
      setView((v) => ({
        ...v,
        hiddenCols: { ...v.hiddenCols, [colId]: !v.hiddenCols[colId] },
      }));
    },
  };

  function openContextMenu(e) {
    const items = buildDatafeedTableMenu({
      selectedRows,
      hasSelection: selectedRows.length > 0,
      singleSelection: selectedRows.length === 1,
      columnSort,
      view,
      handlers: menuHandlers,
      toast: showToast,
    });
    show(e, items);
  }

  useEffect(() => {
    function onKeyDown(e) {
      if (recordId) return;
      if (!(e.ctrlKey || e.metaKey)) return;
      const key = e.key.toLowerCase();
      if (key === "n") {
        e.preventDefault();
        handleAdd();
      } else if (key === "u") {
        e.preventDefault();
        if (selectedRows.length) openDetail(selectedRows[selectedRows.length - 1]);
      } else if (key === "d") {
        e.preventDefault();
        if (selectedRows.length) handleDelete();
      } else if (key === "f") {
        e.preventDefault();
        menuHandlers.find();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  });

  return (
    <div className="module-root module-split-root" onClick={close}>
      <ApiBar state={apiState} />
      <ModuleToolbar
        onReload={reload}
        onOpen={() => openDetail(selectedRows[selectedRows.length - 1])}
        editLabel={config.editLabel || "✎ Data Feed Settings…"}
        canOpen={selectedRows.length > 0}
      />
      <SearchField
        value={search}
        onChange={setSearch}
        onSubmit={reload}
        placeholder={config.searchPlaceholder}
        inputProps={{ "data-df-search": true }}
      />
      <div className="module-split-main df-table-pane">
        <DatafeedTable
          rows={rows}
          selected={selected}
          onToggleSelect={toggleSelect}
          onOpenRow={openDetail}
          onContextMenu={openContextMenu}
          view={view}
        />
      </div>
      {recordId && (
        <DatafeedSettingsModule
          datafeedId={recordId}
          onClose={() => {
            closeSettings();
            refreshListAndNav();
          }}
        />
      )}
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menu.items} onClose={close} />
      )}
      <input
        ref={fileInputRef}
        type="file"
        accept="application/json,.json"
        hidden
        onChange={handleImportChange}
      />
      {toast && <div className="ctx-toast show">{toast}</div>}
    </div>
  );
}

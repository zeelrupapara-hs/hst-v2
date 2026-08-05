import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { ApiBar, SearchField } from "../ui/index.js";
import { ContextMenu, useContextMenu } from "../ui/ContextMenu.jsx";
import { ModuleToolbar } from "../data/index.js";
import { SymbolTable } from "../symbols/SymbolTable.jsx";
import { SymbolSettingsModule } from "../symbols/SymbolSettingsModule.jsx";
import { useSymbolsList, useRowSelection } from "../../hooks/index.js";
import { symbolNavLink } from "../../lib/symbolNavTree.js";
import { buildSymbolTableMenu } from "../../lib/symbolContextMenu.js";
import {
  createSymbol,
  deleteSymbols,
  fetchSymbol,
  moveSymbol,
  saveSymbol,
} from "../../lib/data.js";

const DEFAULT_VIEW = {
  autoArrange: true,
  grid: true,
  hiddenCols: {},
};

export function SplitModule({ config, panel = "admin", moduleId = "symbols" }) {
  const { recordId } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const folderPath = searchParams.get("folder") || "";
  const [search, setSearch] = useState("");
  const [view, setView] = useState(DEFAULT_VIEW);
  const [columnSort, setColumnSort] = useState(false);
  const [toast, setToast] = useState(null);
  const fileInputRef = useRef(null);
  const { menu, show, close } = useContextMenu();

  const { rows, apiState, reload } = useSymbolsList(
    config,
    search,
    folderPath
  );
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
    const t = setTimeout(() => setToast(null), 2800);
    return () => clearTimeout(t);
  }, []);

  function openDetail(row) {
    if (!row) return;
    const q = folderPath ? `?folder=${encodeURIComponent(folderPath)}` : "";
    navigate(
      `/${panel}/${moduleId}/${encodeURIComponent(row[config.idKey])}${q}`
    );
  }

  function closeSettings() {
    navigate(symbolNavLink(panel, folderPath));
  }

  function toggleView(key) {
    setView((v) => ({ ...v, [key]: !v[key] }));
  }

  function toggleColumn(colId) {
    setView((v) => ({
      ...v,
      hiddenCols: { ...v.hiddenCols, [colId]: !v.hiddenCols[colId] },
    }));
  }

  async function handleAdd() {
    const q = folderPath ? `?folder=${encodeURIComponent(folderPath)}` : "";
    navigate(`/${panel}/${moduleId}/new${q}`);
  }

  async function handleAddCopy() {
    if (!selectedRows.length) return;
    const postfix = window.prompt("Postfix for copies (e.g. .x):", ".x");
    if (!postfix) return;
    const subfolder = window.prompt("Subfolder name (optional):", "") || "";
    for (const row of selectedRows) {
      const detail = await fetchSymbol(row.symbol_id);
      if (!detail.data) continue;
      const copyName = row.symbol + postfix;
      const folder = subfolder
        ? folderPath
          ? `${folderPath}\\${subfolder}`
          : subfolder
        : folderPath;
      const created = await createSymbol(folder, copyName);
      if (created.data) {
        await saveSymbol(created.data.symbol_id, {
          ...detail.data,
          symbol: copyName,
          path: created.data.path,
          symbol_id: created.data.symbol_id,
        });
      }
    }
    await reload();
    showToast(`Created ${selectedRows.length} symbol copy(ies)`);
  }

  async function handleDelete() {
    if (!selectedRows.length) return;
    const names = selectedRows.map((r) => r.symbol).join(", ");
    if (!window.confirm(`Delete symbol(s): ${names}?`)) return;
    await deleteSymbols(selectedRows.map((r) => r.symbol_id));
    await reload();
    if (recordId) closeSettings();
    showToast(`Deleted ${selectedRows.length} symbol(s)`);
  }

  async function handleMove(direction) {
    const row = selectedRows[0];
    if (!row) return;
    const ok = await moveSymbol(row.symbol_id, direction);
    if (ok.ok) {
      await reload();
      showToast(`Moved ${row.symbol} ${direction}`);
    }
  }

  function handleExportFile() {
    if (!selectedRows.length) return;
    const payload = selectedRows.map((r) => ({
      symbol_id: r.symbol_id,
      symbol: r.symbol,
      path: r.path,
      description: r.description,
      digits: r.digits,
      exec_mode: r.exec_mode,
      trade_mode: r.trade_mode,
    }));
    const blob = new Blob([JSON.stringify(payload, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "symbols-export.json";
    a.click();
    URL.revokeObjectURL(url);
    showToast(`Exported ${selectedRows.length} symbol(s)`);
  }

  function handleImportFile() {
    fileInputRef.current?.click();
  }

  async function handleImportChange(e) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    try {
      const text = await file.text();
      const data = JSON.parse(text);
      const list = Array.isArray(data) ? data : [data];
      for (const item of list) {
        if (item.symbol_id) {
          await saveSymbol(item.symbol_id, item);
        } else if (item.symbol) {
          const created = await createSymbol(folderPath, item.symbol);
          if (created.data) await saveSymbol(created.data.symbol_id, item);
        }
      }
      await reload();
      showToast(`Imported ${list.length} symbol(s)`);
    } catch {
      showToast("Import failed — invalid JSON");
    }
  }

  function openContextMenu(e) {
    const items = buildSymbolTableMenu({
      selectedRows,
      hasSelection: selectedRows.length > 0,
      singleSelection: selectedRows.length === 1,
      columnSort,
      view,
      toast: showToast,
      handlers: {
        add: handleAdd,
        edit: () => openDetail(selectedRows[selectedRows.length - 1]),
        delete: handleDelete,
        addCopy: handleAddCopy,
        moveUp: () => handleMove("up"),
        moveDown: () => handleMove("down"),
        sort: () =>
          showToast("Sort Alphabetically — server-side sort not implemented"),
        exportFile: handleExportFile,
        importFile: handleImportFile,
        importServer: () => showToast("Import from Server — not implemented"),
        journal: () => showToast("Journal — open Toolbox › Journal"),
        charts: () => showToast("Charts — not implemented"),
        ticks: () => showToast("Ticks — not implemented"),
        toggleAutoArrange: () => toggleView("autoArrange"),
        toggleGrid: () => toggleView("grid"),
        toggleColumn: (ctx) => toggleColumn(ctx.meta?.columnId),
      },
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
      } else if (key === "m") {
        e.preventDefault();
        if (selectedRows.length) handleAddCopy();
      } else if (key === "f") {
        e.preventDefault();
        document.querySelector("[data-symbol-search]")?.focus();
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
        editLabel={config.editLabel}
        canOpen={selectedRows.length > 0}
      />
      <SearchField
        value={search}
        onChange={setSearch}
        onSubmit={reload}
        placeholder={config.searchPlaceholder}
        inputProps={{ "data-symbol-search": true }}
      />
      <div className="module-split-main symbols-table-pane" data-symbol-pane>
        <SymbolTable
          rows={rows}
          selected={selected}
          onToggleSelect={toggleSelect}
          onOpenRow={openDetail}
          onContextMenu={openContextMenu}
          view={view}
          onColumnSortChange={setColumnSort}
        />
      </div>
      {recordId && (
        <SymbolSettingsModule
          symbolId={recordId}
          onClose={() => {
            closeSettings();
            reload();
          }}
        />
      )}
      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          items={menu.items}
          onClose={close}
        />
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

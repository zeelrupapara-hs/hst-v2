import { useMemo, useRef, useState, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useSession } from "@/hooks/useSession.js";
import { useToolbox } from "@/hooks/useToolbox.jsx";
import { useListShortcuts } from "@/hooks/useListShortcuts.js";
import { useRegisterToolbarActions } from "@/hooks/useToolbarActions.jsx";
import { ContextMenu, listMenuHead } from "@/components/ui/ContextMenu.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { ExecMode_name, symbolBackgroundCss } from "@/constants/symbols.js";
import { filterSymbolsByFolder, topType } from "@/lib/symbolTree.js";
import { filterSymbolsByMasks, maskFilterLabel } from "@/lib/symbolMasks.js";
import { deleteSymbol } from "@/api/endpoints/symbols.js";
import { SymbolDialog } from "./SymbolDialog.jsx";
import { SymbolBulkDialog } from "./SymbolBulkDialog.jsx";
import { CloneSymbolsDialog } from "./CloneSymbolsDialog.jsx";

const COLUMNS = [
  { key: "symbol", label: "Symbol" },
  { key: "description", label: "Description" },
  { key: "type", label: "Type" },
  { key: "exec_mode", label: "Execution" },
  { key: "digits", label: "Digits", num: true },
];

const cellValue = (row, key) =>
  key === "type" ? topType(row.path) : key === "exec_mode" ? ExecMode_name[row.exec_mode] : row[key];

function DeleteConfirm({ selectedRows, selected, onConfirm, onClose }) {
  const close = useDialogStack(onClose);
  return (
    <div className="config-window sym-confirm">
      <div className="config-title">
        <span className="config-title-text">Delete</span>
      </div>
      <div className="config-body">
        <p>
          Delete {selected.length > 1 ? `${selected.length} symbols` : `symbol '${selectedRows[0]?.symbol}'`}?
        </p>
      </div>
      <div className="config-actions">
        <button type="button" className="config-ok" onClick={() => { onConfirm(); close(); }}>
          OK
        </button>
        <button type="button" onClick={close}>
          Cancel
        </button>
      </div>
    </div>
  );
}

/** Symbols list: table filtered by ?folder, mask search, dialog on double-click. */
export function SymbolsModule() {
  const { symbols, loading, reload } = useSymbols();
  const session = useSession();
  const toolbox = useToolbox();
  const rootRef = useRef(null);
  const [params] = useSearchParams();
  const folder = params.get("folder") || "";
  const [search, setSearch] = useState("");
  const [applied, setApplied] = useState("");
  const [sort, setSort] = useState(null);
  const [selected, setSelected] = useState([]);
  const [dialog, setDialog] = useState(null);
  const [bulk, setBulk] = useState(null);
  const [menu, setMenu] = useState(null);
  const [confirm, setConfirm] = useState(null);
  const [clone, setClone] = useState(null);
  const canEdit = session.can?.right_cfg_symbols !== false;

  const all = useMemo(() => filterSymbolsByFolder(symbols, folder), [symbols, folder]);

  const rows = useMemo(() => {
    const list = applied.trim()
      ? filterSymbolsByMasks(all, applied)
      : all.slice();
    if (sort) {
      list.sort((a, b) => {
        const x = cellValue(a, sort.key) ?? "";
        const y = cellValue(b, sort.key) ?? "";
        const cmp = typeof x === "number" ? x - y : String(x).localeCompare(String(y));
        return sort.desc ? -cmp : cmp;
      });
    }
    return list;
  }, [all, applied, sort]);

  const selectedRows = useMemo(
    () => selected.map((i) => rows[i]).filter(Boolean),
    [selected, rows],
  );

  function pick(i, e) {
    if (e.ctrlKey || e.metaKey) {
      setSelected((prev) => (prev.includes(i) ? prev.filter((n) => n !== i) : [...prev, i]));
    } else if (e.shiftKey && selected.length) {
      const last = selected[selected.length - 1];
      const [lo, hi] = [Math.min(last, i), Math.max(last, i)];
      setSelected(Array.from({ length: hi - lo + 1 }, (_, n) => lo + n));
    } else {
      setSelected([i]);
    }
  }

  const openAdd = useCallback(() => setDialog({ id: "new" }), []);

  const openEditAction = useCallback(() => {
    if (!selectedRows.length) return;
    if (selectedRows.length === 1) {
      setDialog({ id: selectedRows[0].symbol_id });
    } else {
      setBulk(selectedRows.map((r) => r.symbol_id));
    }
  }, [selectedRows]);

  const openDelete = useCallback(() => setConfirm(true), []);

  async function doDelete() {
    const targets = selectedRows;
    for (const row of targets) {
      const res = await deleteSymbol(row.symbol_id);
      if (!res.ok) window.alert(res.message || "delete failed");
    }
    setSelected([]);
    reload();
    session.refreshNav?.();
  }

  function saved() {
    reload();
    session.refreshNav?.();
  }

  function openJournal() {
    const row = selectedRows[0];
    if (!row) return;
    toolbox?.openJournal?.(row.symbol);
  }

  const hasSelection = selected.length > 0;
  const head = listMenuHead({
    onAdd: openAdd,
    onEdit: openEditAction,
    onDelete: openDelete,
    hasSelection,
  });

  useListShortcuts(rootRef, {
    onAdd: canEdit ? openAdd : undefined,
    onEdit: canEdit && hasSelection ? openEditAction : undefined,
    onDelete: canEdit && hasSelection ? openDelete : undefined,
    onFind: () => rootRef.current?.querySelector(".sym-search-bar input")?.focus(),
  });

  useRegisterToolbarActions({
    onAdd: canEdit ? openAdd : undefined,
    onEdit: canEdit && hasSelection ? openEditAction : undefined,
    onDelete: canEdit && hasSelection ? openDelete : undefined,
    canAdd: canEdit,
    canEdit: canEdit && hasSelection,
    canDelete: canEdit && hasSelection,
  });

  const items = [
    { ...head[0], disabled: hasSelection && selected.length > 1 },
    {
      label: "Add Copy",
      shortcut: "Ctrl+M",
      onClick: () => setClone(selectedRows),
    },
    head[1],
    head[2],
    "sep",
    { label: "Move Up", disabled: true },
    { label: "Move Down", disabled: true },
    { label: "Sort Alphabetically", onClick: () => setSort({ key: "symbol", desc: false }) },
    "sep",
    { label: "Copy As", disabled: true, items: [{ label: "Text", disabled: true }] },
    { label: "Export to File", disabled: true },
    { label: "Import from File", disabled: true },
    { label: "Import from Server", disabled: true },
    "sep",
    {
      label: "Journal",
      disabled: !hasSelection || selected.length > 1,
      onClick: openJournal,
    },
    { label: "Charts", disabled: !hasSelection || selected.length > 1 },
    { label: "Ticks", disabled: !hasSelection || selected.length > 1 },
    "sep",
    { label: "Find", shortcut: "Ctrl+F", onClick: () => rootRef.current?.querySelector(".sym-search-bar input")?.focus() },
    { label: "Auto Arrange", checked: true },
    { label: "Grid", checked: true },
    { label: "Columns", disabled: true, items: COLUMNS.map((c) => ({ label: c.label, checked: true })) },
  ];

  const filterLabel = applied ? maskFilterLabel(applied) : null;

  return (
    <div className="module-root" ref={rootRef} tabIndex={-1}>
      <div
        className="table-wrap"
        onContextMenu={(e) => {
          e.preventDefault();
          setMenu({ x: e.clientX, y: e.clientY });
        }}
      >
        <table className="data-table symbols-table data-table-grid data-table-auto">
          <thead>
            <tr>
              {COLUMNS.map((c) => (
                <th
                  key={c.key}
                  className={c.num ? "num" : undefined}
                  onClick={() =>
                    setSort((prev) =>
                      prev?.key === c.key ? { key: c.key, desc: !prev.desc } : { key: c.key, desc: false },
                    )
                  }
                >
                  {c.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => {
              const bg = symbolBackgroundCss(row.color_background);
              const selectedRow = selected.includes(i);
              return (
              <tr
                key={row.symbol_id}
                className={[selectedRow ? "selected" : "", bg ? "sym-row-bg" : ""].filter(Boolean).join(" ") || undefined}
                style={bg && !selectedRow ? { "--sym-row-bg": bg } : undefined}
                onClick={(e) => pick(i, e)}
                onContextMenu={(e) => !selectedRow && pick(i, e)}
                onDoubleClick={() => canEdit && setDialog({ id: row.symbol_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <span
                      className={bg ? "sym-color-swatch sym-list-swatch" : "sym-coin-icon"}
                      style={bg ? { background: bg } : undefined}
                      aria-hidden="true"
                    />
                    {row.symbol}
                  </span>
                </td>
                <td className="col-description">{row.description || " "}</td>
                <td>{topType(row.path)}</td>
                <td>{ExecMode_name[row.exec_mode]}</td>
                <td className="num">{row.digits}</td>
              </tr>
            );
            })}
          </tbody>
        </table>
      </div>
      <div className="sym-list-tabs">
        <button type="button" className={applied ? "" : "active"} onClick={() => setApplied("")}>
          Symbols ({all.length})
        </button>
        {applied && (
          <button type="button" className="active">
            Symbols {filterLabel} ({rows.length} of {all.length})
          </button>
        )}
      </div>
      <div className="filter-bar sym-search-bar">
        <span className="sym-search-glyph" aria-hidden="true" />
        <input
          type="text"
          placeholder="EURUSD, EUR*, !GBPUSD"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              setApplied(search);
              setSelected([]);
            }
          }}
        />
      </div>
      {confirm && (
        <DialogOverlay>
          <DeleteConfirm
            selectedRows={selectedRows}
            selected={selected}
            onConfirm={doDelete}
            onClose={() => setConfirm(null)}
          />
        </DialogOverlay>
      )}
      {menu && canEdit && (
        <ContextMenu x={menu.x} y={menu.y} onClose={() => setMenu(null)} items={items} />
      )}
      {clone && (
        <CloneSymbolsDialog symbols={clone} folder={folder} onClose={() => setClone(null)} onDone={saved} />
      )}
      {bulk && (
        <SymbolBulkDialog symbolIds={bulk} onClose={() => setBulk(null)} onSaved={saved} />
      )}
      {dialog && (
        <SymbolDialog
          symbolId={dialog.id}
          folderPath={folder}
          onClose={() => setDialog(null)}
          onSaved={saved}
        />
      )}
    </div>
  );
}

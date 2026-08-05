import { useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useSession } from "@/hooks/useSession.js";
import { ExecMode_name } from "@/constants/symbols.js";
import { filterSymbolsByFolder, topType } from "@/lib/symbolTree.js";
import { deleteSymbol } from "@/api/endpoints/symbols.js";
import { SymbolDialog } from "./SymbolDialog.jsx";

const short = (label) => label?.replace(" Execution", "") ?? "";

/** Symbols list: table filtered by ?folder, bottom filter bar, dialog on double-click. */
export function SymbolsModule() {
  const { symbols, loading, reload } = useSymbols();
  const session = useSession();
  const [params] = useSearchParams();
  const folder = params.get("folder") || "";
  const [filter, setFilter] = useState("");
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const canEdit = session.can?.right_cfg_symbols !== false;

  const rows = useMemo(() => {
    let list = filterSymbolsByFolder(symbols, folder);
    const masks = filter
      .split(",")
      .map((m) => m.trim())
      .filter(Boolean);
    if (masks.length) {
      const match = (sym, mask) =>
        new RegExp(`^${mask.replaceAll("*", ".*")}$`, "i").test(sym);
      list = list.filter((r) => {
        let ok = false;
        for (const mask of masks) {
          if (mask.startsWith("!")) {
            if (match(r.symbol, mask.slice(1))) return false;
          } else if (match(r.symbol, mask)) ok = true;
        }
        return ok || !masks.some((m) => !m.startsWith("!"));
      });
    }
    return list.sort((a, b) => a.symbol.localeCompare(b.symbol));
  }, [symbols, folder, filter]);

  async function onDelete(row) {
    if (!window.confirm(`Delete symbol '${row.symbol}'?`)) return;
    const res = await deleteSymbol(row.symbol_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    reload();
    session.refreshNav?.();
  }

  function saved() {
    reload();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="module-toolbar">
        {canEdit && (
          <>
            <button type="button" onClick={() => setDialog({ id: "new" })}>Add</button>
            <button
              type="button"
              disabled={selected == null}
              onClick={() => setDialog({ id: rows[selected]?.symbol_id })}
            >
              Edit
            </button>
            <button
              type="button"
              disabled={selected == null}
              onClick={() => onDelete(rows[selected])}
            >
              Delete
            </button>
          </>
        )}
        <span className="module-note">
          {loading ? "Loading…" : `${rows.length} symbols${folder ? ` in ${folder}` : ""}`}
        </span>
      </div>
      <div className="table-wrap">
        <table className="data-table symbols-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Symbol</th>
              <th>Description</th>
              <th>Type</th>
              <th>Execution</th>
              <th>Digits</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr
                key={row.symbol_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: row.symbol_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <span className="sym-coin-icon" aria-hidden="true" />
                    {row.symbol}
                  </span>
                </td>
                <td className="col-description">{row.description || " "}</td>
                <td>{topType(row.path)}</td>
                <td>{short(ExecMode_name[row.exec_mode])}</td>
                <td>{row.digits}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="filter-bar">
        <span>Filter:</span>
        <input
          type="text"
          placeholder="EUR*, !GBPUSD"
          value={filter}
          onChange={(e) => {
            setFilter(e.target.value);
            setSelected(null);
          }}
        />
      </div>
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

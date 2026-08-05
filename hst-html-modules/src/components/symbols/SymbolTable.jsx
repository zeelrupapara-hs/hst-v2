import { fmtExecMode } from "../../lib/formatters.js";
import { topType } from "../../lib/symbolTree.js";

const COLS = [
  { id: "symbol", label: "Symbol" },
  { id: "description", label: "Description", className: "col-description" },
  { id: "digits", label: "Digits" },
  { id: "type", label: "Type" },
  { id: "execution", label: "Execution" },
];

export function SymbolTable({
  rows,
  selected,
  onToggleSelect,
  onOpenRow,
  onContextMenu,
  view = { grid: true, hiddenCols: {} },
  onColumnSortChange,
}) {
  const isSelected = (i) =>
    selected instanceof Set ? selected.has(i) : selected.includes(i);

  const hidden = view.hiddenCols || {};

  return (
    <div className="table-wrap" onContextMenu={onContextMenu}>
      <table
        className={`data-table symbols-table${view.grid ? " data-table-grid" : ""}${view.autoArrange !== false ? " data-table-auto" : ""}`}
      >
        <thead>
          <tr>
            {COLS.filter((c) => !hidden[c.id]).map((col) => (
              <th
                key={col.id}
                className={col.className}
                title={col.title}
                onClick={() => onColumnSortChange?.(true)}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={row.symbol_id}
              className={isSelected(i) ? "selected" : ""}
              onClick={(e) =>
                onToggleSelect(i, e.ctrlKey || e.metaKey, e.shiftKey)
              }
              onDoubleClick={() => onOpenRow?.(row)}
            >
              {!hidden.symbol && (
                <td>
                  <span className="sym-symbol-cell">
                    <span className="sym-coin-icon" aria-hidden="true" />
                    {row.symbol}
                  </span>
                </td>
              )}
              {!hidden.description && (
                <td className="col-description" title={row.description || ""}>
                  {row.description || "\u00a0"}
                </td>
              )}
              {!hidden.digits && <td>{row.digits}</td>}
              {!hidden.type && <td>{topType(row.path)}</td>}
              {!hidden.execution && <td>{fmtExecMode(row.exec_mode)}</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export { COLS as SYMBOL_TABLE_COLS };

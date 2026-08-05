import { fmtExecMode } from "../../lib/formatters.js";
import { topType } from "../../lib/symbolTree.js";
import { SymbolIconCell } from "./SymbolIconCell.jsx";

export function SymbolTable({
  rows,
  sessionMap,
  selected,
  onToggleSelect,
  onOpenRow,
}) {
  const isSelected = (i) =>
    selected instanceof Set ? selected.has(i) : selected.includes(i);

  return (
    <div className="table-wrap">
      <table className="data-table symbols-table">
        <thead>
          <tr>
            <th className="col-icon" title="Trade mode &amp; session" />
            <th>Symbol</th>
            <th>Type</th>
            <th>Execution</th>
            <th>Digits</th>
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
              <td>
                <SymbolIconCell
                  row={row}
                  sessions={sessionMap[row.symbol_id]}
                />
              </td>
              <td>{row.symbol}</td>
              <td>{topType(row.path)}</td>
              <td>{fmtExecMode(row.exec_mode)}</td>
              <td>{row.digits}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

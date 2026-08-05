import {
  fmtDatafeedLastActive,
  fmtDatafeedSource,
  fmtDatafeedState,
  fmtDatafeedSymbols,
} from "../../lib/datafeedFormatters.js";
import { DATAFEED_TABLE_COLUMNS } from "../../lib/moduleContextMenus.js";

export function DatafeedTable({
  rows,
  selected,
  onToggleSelect,
  onOpenRow,
  onContextMenu,
  view = { grid: true, hiddenCols: {} },
}) {
  const isSelected = (i) =>
    selected instanceof Set ? selected.has(i) : selected.includes(i);
  const hidden = view.hiddenCols || {};

  const cols = [
    { id: "name", render: (r) => r.name },
    { id: "source", render: (r) => fmtDatafeedSource(r) },
    { id: "server", render: (r) => r.feed_server || "—" },
    { id: "symbols", render: (r) => fmtDatafeedSymbols(r) },
    { id: "last_active", render: (r) => fmtDatafeedLastActive(r) },
    { id: "state", render: (r) => fmtDatafeedState(r) },
  ];

  return (
    <div className="table-wrap" onContextMenu={onContextMenu}>
      <table
        className={`data-table df-table${view.grid ? " data-table-grid" : ""}${view.autoArrange !== false ? " data-table-auto" : ""}`}
      >
        <thead>
          <tr>
            {DATAFEED_TABLE_COLUMNS.filter((c) => !hidden[c.id]).map((col) => (
              <th key={col.id}>{col.label}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={row.datafeed_id}
              className={isSelected(i) ? "selected" : ""}
              onClick={(e) =>
                onToggleSelect(i, e.ctrlKey || e.metaKey, e.shiftKey)
              }
              onDoubleClick={() => onOpenRow?.(row)}
            >
              {cols
                .filter((c) => !hidden[c.id])
                .map((col) => (
                  <td key={col.id}>{col.render(row)}</td>
                ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

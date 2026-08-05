import { useEffect, useRef, useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";

export function DatafeedTabIntro({ children }) {
  return (
    <div className="sym-sessions-intro df-tab-intro">
      <span className="sym-tab-intro-icon sym-tab-intro-icon-sm" aria-hidden="true">
        <Icon id="datafeeds" size={40} />
      </span>
      <p>{children}</p>
    </div>
  );
}

export function DfRowIcon() {
  return (
    <span className="df-row-icon" aria-hidden="true">
      $
    </span>
  );
}

export function ParamTypeIcon({ type = 0 }) {
  return (
    <span className="df-param-type" title={type === 1 ? "Integer" : "String"}>
      {type === 1 ? "123" : "ab"}
    </span>
  );
}

function CellEditor({ col, value, onCommit, onCancel }) {
  const inputRef = useRef(null);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select?.();
  }, []);

  if (col.editor === "yesno") {
    return (
      <PropSelect
        fill
        value={value ? "Yes" : "No"}
        options={["Yes", "No"]}
        onChange={(v) => onCommit(v === "Yes")}
      />
    );
  }

  return (
    <input
      ref={inputRef}
      type="text"
      className="df-cell-input"
      defaultValue={value ?? ""}
      onBlur={(e) => onCommit(col.editor === "number" ? Number(e.target.value) || 0 : e.target.value)}
      onKeyDown={(e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          onCommit(col.editor === "number" ? Number(e.currentTarget.value) || 0 : e.currentTarget.value);
        } else if (e.key === "Escape") {
          e.preventDefault();
          onCancel();
        }
      }}
      onClick={(e) => e.stopPropagation()}
      onDoubleClick={(e) => e.stopPropagation()}
    />
  );
}

/**
 * Editable rows panel shared by the Symbols / Translations / Parameters tabs:
 * Up/Down/Add/Edit/Delete rail over an inline-edit grid; edits land in the dialog draft.
 */
export function DfListPanel({
  intro,
  columns,
  rows,
  rowKey,
  canEdit,
  onChangeRows,
  beforeTable,
  emptyLabel = "No items",
  renderCellPrefix,
  showMove = true,
  headless = false,
  showAddRow = false,
  onDefault,
}) {
  const [selected, setSelected] = useState(rows.length ? 0 : -1);
  const [editing, setEditing] = useState(null);

  useEffect(() => {
    setSelected((prev) => (rows.length ? Math.min(Math.max(prev, 0), rows.length - 1) : -1));
  }, [rows.length]);

  const hasSelection = selected >= 0 && selected < rows.length;

  function updateCell(rowIndex, colId, value) {
    onChangeRows(rows.map((row, i) => (i === rowIndex ? { ...row, [colId]: value } : row)));
  }

  function move(delta) {
    const target = selected + delta;
    if (!hasSelection || target < 0 || target >= rows.length) return;
    const next = rows.slice();
    [next[selected], next[target]] = [next[target], next[selected]];
    onChangeRows(next);
    setSelected(target);
  }

  function addRow() {
    const blank = Object.fromEntries(columns.map((c) => [c.id, c.defaultValue ?? ""]));
    onChangeRows([...rows, blank]);
    setSelected(rows.length);
    setEditing({ row: rows.length, col: columns[0].id });
  }

  function removeSelected() {
    if (!hasSelection) return;
    setEditing(null);
    onChangeRows(rows.filter((_, i) => i !== selected));
  }

  const fmt = (row, col) => {
    const raw = row[col.id];
    if (col.editor === "yesno") return raw ? "Yes" : "No";
    return raw == null || raw === "" ? (col.editor === "number" ? "0" : "") : String(raw);
  };

  return (
    <>
      <DatafeedTabIntro>{intro}</DatafeedTabIntro>
      {beforeTable}
      <div className="df-table-panel">
        <div className="df-table-toolbar">
          {showMove && (
            <>
              <button type="button" disabled={!canEdit || selected <= 0} onClick={() => move(-1)}>
                Up
              </button>
              <button
                type="button"
                disabled={!canEdit || !hasSelection || selected >= rows.length - 1}
                onClick={() => move(1)}
              >
                Down
              </button>
              <span className="df-toolbar-gap" aria-hidden="true" />
            </>
          )}
          <button type="button" disabled={!canEdit} onClick={addRow}>
            Add
          </button>
          <button
            type="button"
            disabled={!canEdit || !hasSelection}
            onClick={() => setEditing({ row: selected, col: columns[0].id })}
          >
            Edit
          </button>
          <button type="button" disabled={!canEdit || !hasSelection} onClick={removeSelected}>
            Delete
          </button>
          {onDefault && (
            <button type="button" disabled={!canEdit} onClick={onDefault}>
              Default
            </button>
          )}
        </div>
        <div className="df-table-main">
          <table className="data-table data-table-grid df-sub-table">
            {!headless && (
              <thead>
                <tr>
                  {columns.map((col) => (
                    <th key={col.id} className={col.align}>
                      {col.label}
                    </th>
                  ))}
                </tr>
              </thead>
            )}
            <tbody>
              {rows.length || showAddRow ? (
                rows.map((row, i) => (
                  <tr
                    key={rowKey(row, i)}
                    className={i === selected ? "selected" : ""}
                    onClick={() => setSelected(i)}
                  >
                    {columns.map((col) => (
                      <td
                        key={col.id}
                        className={[canEdit && "df-cell-editable", col.align].filter(Boolean).join(" ")}
                        onDoubleClick={() => canEdit && setEditing({ row: i, col: col.id })}
                      >
                        {editing?.row === i && editing?.col === col.id ? (
                          <CellEditor
                            col={col}
                            value={row[col.id]}
                            onCommit={(v) => {
                              updateCell(i, col.id, v);
                              setEditing(null);
                            }}
                            onCancel={() => setEditing(null)}
                          />
                        ) : (
                          <>
                            {renderCellPrefix?.(row, col)}
                            {fmt(row, col)}
                          </>
                        )}
                      </td>
                    ))}
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={columns.length} className="df-empty">
                    {emptyLabel}
                  </td>
                </tr>
              )}
              {showAddRow && (
                <tr className="df-add-row" onClick={() => canEdit && addRow()}>
                  <td colSpan={columns.length}>
                    <span className="df-add-plus" aria-hidden="true">
                      +
                    </span>
                    click to add...
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

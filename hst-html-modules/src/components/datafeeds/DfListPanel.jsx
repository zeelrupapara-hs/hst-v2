import { useEffect, useRef, useState } from "react";
import { Icon } from "../ui/Icon.jsx";
import { PropSelect } from "../ui/PropSelect.jsx";
import { feedSymbolOptionsWithValue } from "../../lib/feedSymbolOptions.js";

function DatafeedTabIntro({ children }) {
  return (
    <div className="sym-sessions-intro df-tab-intro">
      <span className="sym-tab-intro-icon sym-tab-intro-icon-sm" aria-hidden="true">
        <Icon name="datafeeds" size={40} />
      </span>
      <p>{children}</p>
    </div>
  );
}

function ParamTypeIcon({ type = 0 }) {
  const label = type === 1 ? "123" : "ab";
  return (
    <span className="df-param-type" title={type === 1 ? "Integer" : "String"}>
      {label}
    </span>
  );
}

function formatCellValue(row, col) {
  const raw = row[col.id];
  if (col.editor === "yesno") return raw ? "Yes" : "No";
  if (raw == null || raw === "") {
    return col.editor === "number" ? "0" : "";
  }
  return String(raw);
}

function DfCellEditor({ col, value, onCommit, onCancel }) {
  const inputRef = useRef(null);

  useEffect(() => {
    inputRef.current?.focus();
    if (col.editor === "text" || col.editor === "number") {
      inputRef.current?.select();
    }
  }, [col.editor]);

  if (col.editor === "yesno") {
    return (
      <PropSelect
        narrow
        fill
        value={value ? "Yes" : "No"}
        options={["Yes", "No"]}
        onChange={(v) => onCommit(v === "Yes" ? 1 : 0)}
      />
    );
  }

  if (col.editor === "select") {
    const opts = feedSymbolOptionsWithValue(col.selectOptions, value);
    return (
      <PropSelect
        narrow
        fill
        initialOpen
        value={value ?? ""}
        options={opts}
        onChange={(v) => onCommit(v)}
      />
    );
  }

  return (
    <input
      ref={inputRef}
      type={col.editor === "number" ? "text" : "text"}
      className="df-cell-input"
      defaultValue={formatCellValue({ [col.id]: value }, col)}
      onBlur={(e) => {
        const v =
          col.editor === "number"
            ? Number(e.target.value) || 0
            : e.target.value;
        onCommit(v);
      }}
      onKeyDown={(e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          const v =
            col.editor === "number"
              ? Number(e.currentTarget.value) || 0
              : e.currentTarget.value;
          onCommit(v);
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

export function DfListPanel({
  intro,
  columns,
  rows,
  rowKey,
  canEdit,
  showDefault,
  onChangeRows,
  beforeTable,
  emptyLabel = "No items",
  renderCellPrefix,
}) {
  const [selected, setSelected] = useState(rows.length ? 0 : -1);
  const [editing, setEditing] = useState(null);

  useEffect(() => {
    setSelected((prev) => {
      if (!rows.length) return -1;
      if (prev < 0 || prev >= rows.length) return 0;
      return prev;
    });
  }, [rows.length]);

  const hasSelection = selected >= 0 && selected < rows.length;
  const atTop = !hasSelection || selected === 0;
  const atBottom = !hasSelection || selected >= rows.length - 1;

  function commit(next) {
    onChangeRows?.(next);
  }

  function updateCell(rowIndex, colId, value) {
    const next = rows.map((row, i) =>
      i === rowIndex ? { ...row, [colId]: value } : row
    );
    commit(next);
  }

  function startEdit(rowIndex, colId) {
    if (!canEdit) return;
    const col = columns.find((c) => c.id === colId);
    if (!col || col.editable === false) return;
    setSelected(rowIndex);
    setEditing({ row: rowIndex, col: colId });
  }

  function move(delta) {
    if (!hasSelection || !canEdit) return;
    setEditing(null);
    const next = rows.slice();
    const target = selected + delta;
    if (target < 0 || target >= next.length) return;
    [next[selected], next[target]] = [next[target], next[selected]];
    commit(next);
    setSelected(target);
  }

  function removeSelected() {
    if (!hasSelection || !canEdit) return;
    setEditing(null);
    const next = rows.filter((_, i) => i !== selected);
    commit(next);
    setSelected(Math.min(selected, next.length - 1));
  }

  function addRow() {
    if (!canEdit) return;
    const blank = columns.reduce((acc, col) => {
      acc[col.id] = col.defaultValue ?? "";
      return acc;
    }, {});
    const next = [...rows, blank];
    commit(next);
    const newIndex = next.length - 1;
    setSelected(newIndex);
    const firstEditable = columns.find((c) => c.editable !== false);
    if (firstEditable) {
      setEditing({ row: newIndex, col: firstEditable.id });
    }
  }

  function editSelected() {
    if (!hasSelection || !canEdit) return;
    const firstEditable = columns.find((c) => c.editable !== false);
    if (firstEditable) startEdit(selected, firstEditable.id);
  }

  function restoreDefault() {
    if (!canEdit || !showDefault) return;
    setEditing(null);
    commit([]);
    setSelected(-1);
  }

  return (
    <>
      <DatafeedTabIntro>{intro}</DatafeedTabIntro>
      {beforeTable}
      <div className="df-table-panel">
        <div className="df-table-toolbar">
          <button type="button" disabled={!canEdit || atTop} onClick={() => move(-1)}>
            Up
          </button>
          <button type="button" disabled={!canEdit || atBottom} onClick={() => move(1)}>
            Down
          </button>
          <button type="button" disabled={!canEdit} onClick={addRow}>
            Add
          </button>
          <button
            type="button"
            disabled={!canEdit || !hasSelection}
            onClick={editSelected}
          >
            Edit
          </button>
          <button
            type="button"
            disabled={!canEdit || !hasSelection}
            onClick={removeSelected}
          >
            Delete
          </button>
          {showDefault ? (
            <button type="button" disabled={!canEdit} onClick={restoreDefault}>
              Default
            </button>
          ) : null}
        </div>
        <div className="df-table-main">
          <table className="data-table data-table-grid df-sub-table">
            <thead>
              <tr>
                {columns.map((col) => (
                  <th key={col.id}>{col.label}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.length ? (
                rows.map((row, i) => (
                  <tr
                    key={rowKey(row, i)}
                    className={i === selected ? "selected" : ""}
                    onClick={() => {
                      setSelected(i);
                      if (editing && editing.row !== i) setEditing(null);
                    }}
                  >
                    {columns.map((col) => {
                      const isEditing =
                        editing?.row === i && editing?.col === col.id;
                      const editable = canEdit && col.editable !== false;

                      return (
                        <td
                          key={col.id}
                          className={editable ? "df-cell-editable" : undefined}
                          onDoubleClick={(e) => {
                            e.stopPropagation();
                            startEdit(i, col.id);
                          }}
                        >
                          {isEditing ? (
                            <DfCellEditor
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
                              {renderCellPrefix?.(row, col, i)}
                              {formatCellValue(row, col)}
                            </>
                          )}
                        </td>
                      );
                    })}
                  </tr>
                ))
              ) : (
                <tr className="df-table-empty">
                  <td colSpan={columns.length}>{emptyLabel}</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

export { ParamTypeIcon, DatafeedTabIntro };

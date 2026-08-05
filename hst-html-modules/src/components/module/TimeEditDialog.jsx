import { useCallback, useEffect, useRef, useState } from "react";
import { TIME_DAY_NAMES } from "../../lib/timeFormat.js";

export function TimeEditDialog({ dayIndex, hours, onSave, onClose }) {
  const [draft, setDraft] = useState(() => hours.slice());
  const [focusedRow, setFocusedRow] = useState(dayIndex ?? 0);
  const dragRef = useRef({ active: false, value: true });

  useEffect(() => {
    setDraft(hours.map((day) => day.slice()));
    setFocusedRow(dayIndex ?? 0);
  }, [hours, dayIndex]);

  const setHour = useCallback((row, col, value) => {
    setDraft((prev) => {
      const next = prev.map((day) => day.slice());
      next[row][col] = value;
      return next;
    });
  }, []);

  const toggleRowAll = useCallback((row) => {
    setDraft((prev) => {
      const next = prev.map((day) => day.slice());
      const allOn = next[row].every(Boolean);
      next[row] = next[row].map(() => !allOn);
      return next;
    });
  }, []);

  const handlePointerDown = (row, col) => {
    setFocusedRow(row);
    dragRef.current = { active: true, value: !draft[row][col] };
    setHour(row, col, dragRef.current.value);
  };

  const handlePointerEnter = (row, col) => {
    if (!dragRef.current.active) return;
    setHour(row, col, dragRef.current.value);
  };

  const handlePointerUp = () => {
    dragRef.current.active = false;
  };

  useEffect(() => {
    window.addEventListener("mouseup", handlePointerUp);
    return () => window.removeEventListener("mouseup", handlePointerUp);
  }, []);

  function handleOk() {
    onSave(draft.map((day) => day.slice()));
    onClose();
  }

  return (
    <div className="time-edit-overlay" onClick={onClose}>
      <div
        className="time-edit-dialog"
        role="dialog"
        aria-labelledby="time-edit-title"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="config-title" id="time-edit-title">
          Time
        </div>
        <div className="time-edit-header">
          <span className="time-edit-clock" aria-hidden="true">
            🕐
          </span>
          <p className="time-edit-instruction">
            The adjustment of trading time of server by days of the week. Please
            mark working hours with the blue color.
          </p>
        </div>
        <div className="time-hour-table-wrap">
          <table className="time-hour-table">
          <tbody>
            {TIME_DAY_NAMES.map((name, row) => (
              <tr
                key={name}
                className={focusedRow === row ? "focused" : undefined}
              >
                <th scope="row" className="time-hour-label">
                  {name}:
                </th>
                {Array.from({ length: 24 }, (_, col) => (
                  <td key={col}>
                    <button
                      type="button"
                      className={`time-hour-cell${draft[row][col] ? " active" : ""}`}
                      onMouseDown={(e) => {
                        e.preventDefault();
                        handlePointerDown(row, col);
                      }}
                      onMouseEnter={() => handlePointerEnter(row, col)}
                    >
                      {String(col).padStart(2, "0")}
                    </button>
                  </td>
                ))}
                <td className="time-hour-star-cell">
                  <button
                    type="button"
                    className="time-hour-cell time-hour-star"
                    title="Toggle all hours"
                    onClick={() => toggleRowAll(row)}
                  >
                    *
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
        <div className="config-actions time-edit-footer">
          <button type="button" className="config-ok" onClick={handleOk}>
            OK
          </button>
          <button type="button" onClick={onClose}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}

import { Icon } from "../ui/Icon.jsx";
import { PropSelect } from "../ui/PropSelect.jsx";
import { TIME_ZONE_OPTIONS } from "../../lib/timeZones.js";
import { isDayRow, isSectionRow } from "../../features/modules/settings/time.js";

export function PropertyTable({
  rows,
  selectedRowId,
  onSelectRow,
  onFieldChange,
  onDayDoubleClick,
  onContextMenu,
}) {
  return (
    <table className="prop-table data-table data-table-grid">
      <thead>
        <tr>
          <th className="prop-col-label">Property</th>
          <th className="prop-col-value">Value</th>
        </tr>
      </thead>
      <tbody onContextMenu={onContextMenu}>
        {rows.map((row) => {
          if (isSectionRow(row)) {
            return (
              <tr key={row.id} className="prop-row-section">
                <td colSpan={2}>{row.label}</td>
              </tr>
            );
          }

          const selected = selectedRowId === row.id;
          return (
            <tr
              key={row.id}
              className={selected ? "selected" : undefined}
              onClick={() => onSelectRow(row.id)}
              onDoubleClick={() => {
                if (isDayRow(row)) onDayDoubleClick?.(row);
              }}
            >
              <td className="prop-col-label">
                <span className="prop-row-icon">
                  <Icon name={row.icon || "time"} />
                </span>
                {row.label}
              </td>
              <td className="prop-col-value">
                <PropertyValueCell
                  row={row}
                  selected={selected}
                  onFieldChange={onFieldChange}
                />
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}

function PropertyValueCell({ row, selected, onFieldChange }) {
  if (row.editor === "timezone") {
    return (
      <PropSelect
        wide
        value={row.value}
        options={TIME_ZONE_OPTIONS}
        onChange={(v) => onFieldChange("time_zone", v)}
      />
    );
  }

  if (row.editor === "yesno") {
    return (
      <PropSelect
        narrow
        value={row.value ? "Yes" : "No"}
        options={["Yes", "No"]}
        onChange={(v) => onFieldChange("daylight_saving", v === "Yes")}
      />
    );
  }

  if (row.editor === "text") {
    if (selected) {
      return (
        <input
          type="text"
          className="prop-value-input"
          value={row.value}
          autoFocus
          onClick={(e) => e.stopPropagation()}
          onChange={(e) => onFieldChange("sync_servers", e.target.value)}
        />
      );
    }
    return <span className="prop-value-text">{row.value}</span>;
  }

  if (row.editor === "day") {
    return <span className="prop-value-text">{row.value || ""}</span>;
  }

  return <span className="prop-value-text">{row.value ?? "—"}</span>;
}

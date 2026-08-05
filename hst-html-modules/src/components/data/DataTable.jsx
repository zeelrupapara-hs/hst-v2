import { useNavigate } from "react-router-dom";

export function DataTable({
  columns,
  rows,
  idKey,
  selected,
  onToggleSelect,
  onOpenRow,
}) {
  const navigate = useNavigate();

  return (
    <table className="data-table data-table-grid">
      <thead>
        <tr>
          {columns.map((col) => (
            <th key={col.key || col.label}>{col.label ?? col.key}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {!rows.length ? (
          <tr>
            <td colSpan={columns.length}>No records</td>
          </tr>
        ) : (
          rows.map((row, index) => (
            <tr
              key={idKey ? row[idKey] : index}
              className={selected.has(index) ? "selected" : undefined}
              onClick={(e) =>
                onToggleSelect(index, e.ctrlKey || e.metaKey, e.shiftKey)
              }
              onDoubleClick={() => onOpenRow?.(row)}
            >
              {columns.map((col) => {
                const raw = col.render ? col.render(row, index) : row[col.key];
                if (col.link) {
                  const href = col.link(row);
                  return (
                    <td key={col.key}>
                      <a
                        href={href}
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          navigate(href);
                        }}
                      >
                        {raw}
                      </a>
                    </td>
                  );
                }
                if (col.html) {
                  return (
                    <td
                      key={col.key}
                      dangerouslySetInnerHTML={{ __html: raw ?? "" }}
                    />
                  );
                }
                return <td key={col.key}>{raw ?? "—"}</td>;
              })}
            </tr>
          ))
        )}
      </tbody>
    </table>
  );
}

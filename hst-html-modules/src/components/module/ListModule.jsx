import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { ApiBar, SearchField } from "../ui/index.js";
import { DataTable, ModuleToolbar } from "../data/index.js";
import { useListData, useRowSelection } from "../../hooks/index.js";

export function ListModule({ config, detailPath }) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [search, setSearch] = useState("");
  const { rows: allRows, apiState, reload } = useListData(config, search);

  // Selecting a folder in the navigator narrows the list to what is under it, the way a folder
  // does anywhere else. The folder is a path prefix, and the record that sits at that path
  // belongs to it as much as the ones beneath.
  const folder = searchParams.get("folder") || "";
  const pathKey = config.pathKey;
  const rows =
    folder && pathKey
      ? allRows.filter((r) => {
          const path = String(r?.[pathKey] ?? "");
          return path === folder || path.startsWith(`${folder}\\`);
        })
      : allRows;

  const { selected, toggleSelect, resetSelection, selectedRows } = useRowSelection(rows);

  useEffect(() => {
    resetSelection(rows.length);
  }, [rows.length, resetSelection]);

  function openDetail(row) {
    if (!detailPath || !row) return;
    navigate(detailPath(row[config.idKey]));
  }

  return (
    <div className="module-root">
      <ApiBar state={apiState} />
      {config.toolbar !== false && (
        <ModuleToolbar
          onReload={reload}
          onOpen={detailPath ? () => openDetail(selectedRows[selectedRows.length - 1]) : null}
          editLabel={config.editLabel}
          canOpen={selectedRows.length > 0}
        />
      )}
      {config.note && (
        <p className="module-note" dangerouslySetInnerHTML={{ __html: config.note }} />
      )}
      {config.search !== false && (
        <SearchField
          value={search}
          onChange={setSearch}
          onSubmit={reload}
          placeholder={config.searchPlaceholder}
        />
      )}
      <div className="module-content table-wrap">
        <DataTable
          columns={config.columns}
          rows={rows}
          idKey={config.idKey}
          selected={selected}
          onToggleSelect={toggleSelect}
          onOpenRow={detailPath ? openDetail : undefined}
        />
      </div>
    </div>
  );
}

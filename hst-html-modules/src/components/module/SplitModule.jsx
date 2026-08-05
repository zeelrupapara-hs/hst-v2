import { useEffect, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { ApiBar, SearchField } from "../ui/index.js";
import { ModuleToolbar } from "../data/index.js";
import { SymbolTable } from "../symbols/SymbolTable.jsx";
import { SymbolSettingsModule } from "../symbols/SymbolSettingsModule.jsx";
import { useSymbolsList, useRowSelection } from "../../hooks/index.js";
import { symbolNavLink } from "../../lib/symbolNavTree.js";

export function SplitModule({ config, panel = "admin", moduleId = "symbols" }) {
  const { recordId } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const folderPath = searchParams.get("folder") || "";
  const [search, setSearch] = useState("");
  const { rows, sessionMap, apiState, reload } = useSymbolsList(
    config,
    search,
    folderPath
  );
  const { selected, toggleSelect, resetSelection, selectedRows, selectRow } =
    useRowSelection(rows);

  useEffect(() => {
    if (recordId && rows.length) {
      const idx = rows.findIndex(
        (r) => String(r[config.idKey]) === String(recordId)
      );
      if (idx >= 0) {
        selectRow(idx);
        return;
      }
    }
    resetSelection(rows.length);
  }, [rows, rows.length, recordId, config.idKey, selectRow, resetSelection]);

  function openDetail(row) {
    if (!row) return;
    const q = folderPath ? `?folder=${encodeURIComponent(folderPath)}` : "";
    navigate(
      `/${panel}/${moduleId}/${encodeURIComponent(row[config.idKey])}${q}`
    );
  }

  function closeSettings() {
    navigate(symbolNavLink(panel, folderPath));
  }

  return (
    <div className="module-root module-split-root">
      <ApiBar state={apiState} />
      <ModuleToolbar
        onReload={reload}
        onOpen={() => openDetail(selectedRows[selectedRows.length - 1])}
        editLabel={config.editLabel}
        canOpen={selectedRows.length > 0}
      />
      <SearchField
        value={search}
        onChange={setSearch}
        onSubmit={reload}
        placeholder={config.searchPlaceholder}
      />
      <div className="module-split-main symbols-table-pane" data-symbol-pane>
        <SymbolTable
          rows={rows}
          sessionMap={sessionMap}
          selected={selected}
          onToggleSelect={toggleSelect}
          onOpenRow={openDetail}
        />
      </div>
      {recordId && (
        <SymbolSettingsModule symbolId={recordId} onClose={closeSettings} />
      )}
    </div>
  );
}

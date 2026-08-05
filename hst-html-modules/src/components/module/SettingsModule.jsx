import { useMemo, useRef, useState } from "react";
import { PropertyTable } from "../data/PropertyTable.jsx";
import { ApiBar } from "../ui/ApiBar.jsx";
import { ModuleToolbar } from "../data/ModuleToolbar.jsx";
import { TimeEditDialog } from "./TimeEditDialog.jsx";
import { useTimeSettings } from "../../hooks/useTimeSettings.js";
import {
  buildTimePropertyRows,
  isDayRow,
} from "../../features/modules/settings/time.js";

export function SettingsModule({ config }) {
  const {
    settings,
    selectedRowId,
    setSelectedRowId,
    apiState,
    toast,
    reload,
    updateField,
    updateDaySchedule,
    importSettings,
  } = useTimeSettings();

  const [editDayIndex, setEditDayIndex] = useState(null);
  const [menu, setMenu] = useState(null);
  const fileInputRef = useRef(null);

  const rows = useMemo(
    () => (settings ? buildTimePropertyRows(settings) : []),
    [settings]
  );

  const selectedRow = rows.find((r) => r.id === selectedRowId);
  const canEditDay = selectedRow && isDayRow(selectedRow);

  function openEditDialog(row) {
    const target = row ?? selectedRow;
    if (!target || !isDayRow(target)) return;
    setEditDayIndex(target.dayIndex);
  }

  function handleContextMenu(e) {
    e.preventDefault();
    const row = rows.find((r) => r.id === selectedRowId);
    setMenu({ x: e.clientX, y: e.clientY, canEdit: row && isDayRow(row) });
  }

  function closeMenu() {
    setMenu(null);
  }

  function exportSettings() {
    if (!settings) return;
    const blob = new Blob([JSON.stringify(settings, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "hst-time-settings.json";
    a.click();
    URL.revokeObjectURL(url);
    closeMenu();
  }

  function handleImportFile(e) {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      try {
        const data = JSON.parse(reader.result);
        importSettings(data);
      } catch {
        alert("Invalid JSON file");
      }
    };
    reader.readAsText(file);
    e.target.value = "";
    closeMenu();
  }

  if (!settings && apiState.loading) {
    return (
      <div className="module-root">
        <ApiBar state={apiState} />
        <p className="module-note">Loading…</p>
      </div>
    );
  }

  return (
    <div className="module-root" onClick={closeMenu}>
      <ApiBar state={apiState} />
      <ModuleToolbar
        onReload={reload}
        onOpen={canEditDay ? () => openEditDialog() : null}
        editLabel="✎ Edit"
        canOpen={canEditDay}
      />
      <div className="module-content table-wrap">
        <PropertyTable
          rows={rows}
          selectedRowId={selectedRowId}
          onSelectRow={setSelectedRowId}
          onFieldChange={updateField}
          onDayDoubleClick={openEditDialog}
          onContextMenu={handleContextMenu}
        />
      </div>

      {editDayIndex != null && settings && (
        <TimeEditDialog
          dayIndex={editDayIndex}
          hours={settings.schedule}
          onSave={updateDaySchedule}
          onClose={() => setEditDayIndex(null)}
        />
      )}

      {menu && (
        <div
          className="ctx-menu prop-ctx-menu"
          style={{ left: menu.x, top: menu.y }}
          onClick={(e) => e.stopPropagation()}
        >
          <div
            className={`ctx-item${menu.canEdit ? "" : " disabled"}`}
            onClick={() => {
              if (menu.canEdit) {
                openEditDialog();
                closeMenu();
              }
            }}
          >
            Edit
          </div>
          <div className="ctx-item" onClick={exportSettings}>
            Export to File
          </div>
          <div
            className="ctx-item"
            onClick={() => fileInputRef.current?.click()}
          >
            Import from File
          </div>
        </div>
      )}

      <input
        ref={fileInputRef}
        type="file"
        accept="application/json,.json"
        hidden
        onChange={handleImportFile}
      />

      {toast && <div className="time-toast">{toast}</div>}
    </div>
  );
}

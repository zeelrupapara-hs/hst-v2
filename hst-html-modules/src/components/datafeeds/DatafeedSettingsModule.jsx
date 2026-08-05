import { useEffect, useState } from "react";
import { SettingsDialog } from "../ui/SettingsDialog.jsx";
import { useDatafeedSettings } from "../../hooks/useDatafeedSettings.js";
import { useDialogDrag } from "../../hooks/useDialogDrag.js";
import { isNewDatafeedId } from "../../lib/datafeedDraft.js";
import { isNewsMode } from "../../lib/datafeedModules.js";
import { DATAFEED_TABS } from "./DatafeedTabPanels.jsx";

/** Data feed properties — draggable modal over the list. */
export function DatafeedSettingsModule({ datafeedId, onClose }) {
  const isNew = isNewDatafeedId(datafeedId);
  const [activeTab, setActiveTab] = useState("common");
  const { offset, onTitlePointerDown } = useDialogDrag(datafeedId);
  const {
    datafeed,
    modules,
    isNew: creating,
    updateField,
    changeMode,
    saveAll,
    toast,
  } = useDatafeedSettings(datafeedId);

  useEffect(() => {
    setActiveTab("common");
  }, [datafeedId]);

  async function handleOk() {
    const saved = await saveAll();
    if (creating && !saved) return;
    onClose?.(saved);
  }

  const title = creating
    ? "Data Feed: New"
    : `Data Feed: ${datafeed?.name || "…"}`;

  const isNewsOnly = isNewsMode(datafeed?.mode);

  return (
    <div className="dialog-overlay" onClick={() => onClose?.()} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          mode="overlay"
          draggable
          onTitlePointerDown={onTitlePointerDown}
          className="sym-config-window df-config-window"
          title={title}
          tabs={
            <div className="config-tabs df-config-tabs">
              {DATAFEED_TABS.filter(
                (tab) =>
                  !isNewsOnly || !["symbols", "translations"].includes(tab.id)
              ).map((tab) => (
                <button
                  key={tab.id}
                  type="button"
                  className={activeTab === tab.id ? "active" : ""}
                  onClick={() => setActiveTab(tab.id)}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions df-config-actions">
              <button type="button" className="config-ok" onClick={handleOk}>
                OK
              </button>
              <button type="button" onClick={() => onClose?.()}>
                Cancel
              </button>
              <button
                type="button"
                onClick={() =>
                  alert("Data feed help — see the administrator documentation.")
                }
              >
                Help
              </button>
            </div>
          }
        >
          {DATAFEED_TABS.filter(
            (tab) => !isNewsOnly || !["symbols", "translations"].includes(tab.id)
          ).map((tab) => (
            <div
              key={tab.id}
              className={`config-panel${activeTab === tab.id ? " active" : ""}`}
            >
              {tab.Panel && (
                <tab.Panel
                  d={datafeed}
                  modules={modules}
                  onFieldChange={updateField}
                  onModeChange={changeMode}
                  isNew={creating}
                />
              )}
            </div>
          ))}
        </SettingsDialog>
        {toast && <div className="module-toast sym-settings-toast">{toast}</div>}
      </div>
    </div>
  );
}

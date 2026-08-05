import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { SettingsDialog } from "../ui/SettingsDialog.jsx";
import { useSymbolSettings } from "../../hooks/useSymbolSettings.js";
import { useDialogDrag } from "../../hooks/useDialogDrag.js";
import { symbolNavLink } from "../../lib/symbolNavTree.js";
import { symbolFolder } from "../../lib/symbolTree.js";
import { isNewSymbolId } from "../../lib/symbolDraft.js";
import { SYMBOL_TABS } from "./SymbolTabPanels.jsx";
import { SymbolSessionsTab } from "./SymbolSessionsTab.jsx";

/** Symbol properties — draggable modal over the symbols list. */
export function SymbolSettingsModule({ symbolId, onClose }) {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const folderPath = searchParams.get("folder") || "";
  const isNew = isNewSymbolId(symbolId);
  const [activeTab, setActiveTab] = useState("common");
  const { offset, onTitlePointerDown } = useDialogDrag(symbolId);
  const {
    symbol,
    isNew: creating,
    updateField,
    updateSessions,
    updateTimeLimits,
    saveAll,
    toast,
  } = useSymbolSettings(symbolId, folderPath);

  useEffect(() => {
    setActiveTab("common");
  }, [symbolId]);

  function close() {
    if (onClose) {
      onClose();
      return;
    }
    const folder =
      searchParams.get("folder") ||
      (symbol?.path ? symbolFolder(symbol.path) : "");
    navigate(symbolNavLink("admin", folder));
  }

  async function handleOk() {
    const saved = await saveAll();
    if (creating && !saved) return;
    close();
  }

  const titlePath = creating
    ? folderPath || "New"
    : symbol?.path || "…";

  return (
    <div
      className="dialog-overlay"
      onClick={close}
      role="presentation"
    >
      <div
        className="dialog-positioner"
        style={{
          transform: `translate(${offset.x}px, ${offset.y}px)`,
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          mode="overlay"
          draggable
          onTitlePointerDown={onTitlePointerDown}
          className="sym-config-window"
          title={`Symbol: ${titlePath}`}
          tabs={
            <div className="config-tabs sym-config-tabs">
              {SYMBOL_TABS.map((tab) => (
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
            <div className="config-actions">
              <button type="button" className="config-ok" onClick={handleOk}>
                OK
              </button>
              <button type="button" onClick={close}>
                Cancel
              </button>
              <button
                type="button"
                onClick={() =>
                  alert(
                    "Symbol settings help — see the administrator documentation."
                  )
                }
              >
                Help
              </button>
            </div>
          }
        >
          {SYMBOL_TABS.map((tab) => (
            <div
              key={tab.id}
              className={`config-panel${activeTab === tab.id ? " active" : ""}`}
            >
              {tab.id === "sessions" ? (
                symbol && (
                  <SymbolSessionsTab
                    symbol={symbol}
                    onUpdateSessions={updateSessions}
                    onUpdateTimeLimits={updateTimeLimits}
                  />
                )
              ) : tab.id === "common" ? (
                tab.Panel && (
                  <tab.Panel
                    s={symbol}
                    onFieldChange={updateField}
                    isNew={creating}
                  />
                )
              ) : (
                tab.Panel && <tab.Panel s={symbol} />
              )}
            </div>
          ))}
        </SettingsDialog>
        {toast && <div className="module-toast sym-settings-toast">{toast}</div>}
      </div>
    </div>
  );
}

import { useEffect, useMemo, useRef, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { fetchSymbol, updateSymbol } from "@/api/endpoints/symbols.js";
import {
  CommonTab,
  CurrencyTab,
  ExecutionTab,
  QuotesTab,
  SwapsTab,
  TradeTab,
} from "./SymbolTabs.jsx";

const TABS = [
  { id: "common", label: "Common", Panel: CommonTab },
  { id: "currency", label: "Currency", Panel: CurrencyTab },
  { id: "quotes", label: "Quotes", Panel: QuotesTab },
  { id: "trade", label: "Trade", Panel: TradeTab },
  { id: "execution", label: "Execution", Panel: ExecutionTab },
  { id: "swaps", label: "Swaps", Panel: SwapsTab },
];

const BULK_LOCKED = new Set(["symbol", "path", "symbol_id", "sessions"]);

function sameValue(a, b) {
  return JSON.stringify(a) === JSON.stringify(b);
}

/** Fields that differ across the selection cannot be bulk-edited (MT5 group-work rule). */
function buildLockField(rows) {
  const locked = new Set(BULK_LOCKED);
  if (rows.length < 2) return (key) => locked.has(key);

  for (const key of Object.keys(rows[0])) {
    if (BULK_LOCKED.has(key)) continue;
    const first = rows[0][key];
    for (let i = 1; i < rows.length; i++) {
      if (!sameValue(first, rows[i][key])) {
        locked.add(key);
        break;
      }
    }
  }
  return (key) => locked.has(key);
}

function buildPatch(draft, baseline, touched) {
  const patch = {};
  for (const key of touched) {
    if (BULK_LOCKED.has(key)) continue;
    if (!sameValue(draft[key], baseline[key])) patch[key] = draft[key];
  }
  return patch;
}

/**
 * Bulk symbol settings — only touched fields that are common across the selection are written.
 * @param {{symbolIds: number[], onClose: Function, onSaved: Function}} props
 */
export function SymbolBulkDialog({ symbolIds, onClose, onSaved }) {
  const [activeTab, setActiveTab] = useState("common");
  const [rows, setRows] = useState(null);
  const [draft, setDraft] = useState(null);
  const [baseline, setBaseline] = useState(null);
  const [error, setError] = useState("");
  const touched = useRef(new Set());
  const { offset, onTitlePointerDown } = useDialogDrag(`bulk-${symbolIds.join(",")}`);
  const close = useDialogStack(onClose);

  useEffect(() => {
    Promise.all(symbolIds.map((id) => fetchSymbol(id))).then((results) => {
      const loaded = results.filter((r) => r.ok).map((r) => r.data);
      if (!loaded.length) {
        setError("failed to load symbols");
        return;
      }
      setRows(loaded);
      setDraft({ ...loaded[0] });
      setBaseline({ ...loaded[0] });
    });
  }, [symbolIds]);

  const lockField = useMemo(() => (rows ? buildLockField(rows) : () => true), [rows]);

  function set(key, value) {
    if (lockField(key)) return;
    touched.current.add(key);
    setDraft((prev) => ({ ...prev, [key]: value }));
  }

  async function handleOk() {
    if (!draft || !baseline || !rows) return;
    const patch = buildPatch(draft, baseline, touched.current);
    if (!Object.keys(patch).length) {
      close();
      return;
    }

    let failed = false;
    for (const row of rows) {
      const res = await updateSymbol(row.symbol_id, patch);
      if (!res.ok) {
        failed = true;
        setError(res.message || `update failed for ${row.symbol}`);
        break;
      }
    }
    if (failed) return;
    onSaved();
    close();
  }

  const title =
    rows?.length === 1
      ? `Symbol: ${rows[0].path}`
      : `${rows?.length ?? symbolIds.length} symbols selected`;

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          className="sym-config-window sym-bulk-window"
          width={787}
          height={620}
          title={title}
          tabs={
            <div className="config-tabs sym-config-tabs">
              {TABS.map((tab) => (
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
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>
                OK
              </button>
              <button type="button" onClick={close}>
                Cancel
              </button>
            </div>
          }
        >
          {!draft && !error && <p className="module-note">Loading…</p>}
          {draft &&
            TABS.map(({ id, Panel }) => (
              <div key={id} className={`config-panel${activeTab === id ? " active" : ""}`}>
                <Panel s={draft} set={set} lockField={lockField} />
              </div>
            ))}
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

import { useEffect, useRef, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack, prevTabEscape } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { createSymbol, fetchSymbol, updateSymbol } from "@/api/endpoints/symbols.js";
import {
  deriveSymbolCurrencies,
} from "@/lib/symbolCurrency.js";
import {
  CommonTab,
  CurrencyTab,
  ExecutionTab,
  MarginRatesTab,
  MarginTab,
  QuotesTab,
  SwapsTab,
  TradeTab,
} from "./SymbolTabs.jsx";
import { SymbolSessionsTab } from "./SymbolSessionsTab.jsx";
import { symbolFolder } from "@/lib/symbolTree.js";

const TABS = [
  { id: "common", label: "Common", Panel: CommonTab },
  { id: "currency", label: "Currency", Panel: CurrencyTab },
  { id: "quotes", label: "Quotes", Panel: QuotesTab },
  { id: "trade", label: "Trade", Panel: TradeTab },
  { id: "execution", label: "Execution", Panel: ExecutionTab },
  { id: "margin", label: "Margin", Panel: MarginTab },
  { id: "margin_rates", label: "Margin Rates", Panel: MarginRatesTab },
  { id: "swaps", label: "Swaps", Panel: SwapsTab },
  { id: "sessions", label: "Sessions", Panel: SymbolSessionsTab },
];

function newDraft(folderPath) {
  return {
    symbol: "",
    path: folderPath ? `${folderPath}\\` : "",
    description: "",
    currency_base: "USD",
    currency_profit: "USD",
    currency_margin: "USD",
    currency_base_digits: 2,
    currency_profit_digits: 2,
    currency_margin_digits: 2,
    digits: 5,
    trade_mode: 4,
    calc_mode: 0,
    exec_mode: 2,
    gtc_mode: 0,
    contract_size: 100000,
    fill_flags: 3,
    expir_flags: 15,
    order_flags: 127,
    tick_flags: 1,
    volume_min: 10000,
    volume_max: 100000000,
    volume_step: 10000,
    sessions: [],
  };
}

const sessionRow = ({ type, day, open, close }) => ({ type, day, open, close });

/** Changed fields only; sessions compare structurally and replace wholesale. */
function diff(draft, original) {
  const patch = {};
  for (const key of Object.keys(draft)) {
    if (key === "sessions") {
      const a = JSON.stringify((draft.sessions || []).map(sessionRow));
      const b = JSON.stringify((original.sessions || []).map(sessionRow));
      if (a !== b) patch.sessions = (draft.sessions || []).map(sessionRow);
    } else if (draft[key] !== original[key]) {
      patch[key] = draft[key];
    }
  }
  return patch;
}

/**
 * Symbol properties dialog. symbolId "new" creates; OK writes changed fields only.
 * @param {{symbolId: number|"new", folderPath?: string, onClose: Function, onSaved: Function}} props
 */
export function SymbolDialog({ symbolId, folderPath = "", onClose, onSaved }) {
  const isNew = symbolId === "new";
  const [activeTab, setActiveTab] = useState("common");
  const [draft, setDraft] = useState(isNew ? newDraft(folderPath) : null);
  const [original, setOriginal] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(symbolId);
  const close = useDialogStack(onClose, {
    onEscape: () => prevTabEscape(TABS, activeTab, setActiveTab),
  });

  useEffect(() => {
    if (isNew) return;
    fetchSymbol(symbolId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load symbol");
        return;
      }
      setDraft(res.data);
      setOriginal(res.data);
    });
  }, [symbolId, isNew]);

  // the stored currencies stand on open; only a new name or calculation re-derives them
  const derivedFor = useRef(null);
  useEffect(() => {
    if (!draft) return;
    const key = `${draft.symbol}|${draft.calc_mode}`;
    if (derivedFor.current === null && !isNew) {
      derivedFor.current = key;
      return;
    }
    if (derivedFor.current === key) return;
    derivedFor.current = key;
    const patch = deriveSymbolCurrencies({
      symbol: draft.symbol,
      calc_mode: draft.calc_mode,
      current: draft,
    });
    if (!patch) return;
    setDraft((prev) => {
      if (!prev) return prev;
      let changed = false;
      const next = { ...prev };
      for (const [k, v] of Object.entries(patch)) {
        if (next[k] !== v) {
          next[k] = v;
          changed = true;
        }
      }
      return changed ? next : prev;
    });
  }, [draft?.symbol, draft?.calc_mode]);

  function set(keyOrFields, value) {
    setDraft((prev) => {
      if (typeof keyOrFields === "object" && keyOrFields !== null) {
        return { ...prev, ...keyOrFields };
      }
      const next = { ...prev, [keyOrFields]: value };
      if (keyOrFields === "symbol") {
        const name = String(value ?? "");
        if (isNew) {
          next.path = folderPath ? `${folderPath}\\${name}` : name;
        } else {
          const folder = symbolFolder(prev.path || "");
          next.path = folder ? `${folder}\\${name}` : name;
        }
      }
      return next;
    });
  }

  async function handleOk() {
    if (!draft) return;
    const trimmed = String(draft.symbol ?? "").trim();
    const folder = symbolFolder(draft.path || "");
    const normalized =
      trimmed === draft.symbol
        ? draft
        : { ...draft, symbol: trimmed, path: folder ? `${folder}\\${trimmed}` : trimmed };
    // the reference's clone-by-rename: editing the name creates a new symbol, the original stays
    if (isNew || trimmed !== original?.symbol) {
      if (!trimmed) {
        setError("Symbol name is required");
        return;
      }
      const res = await createSymbol({ ...normalized, sessions: (normalized.sessions || []).map(sessionRow) });
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
    } else {
      const patch = diff(normalized, original);
      if (Object.keys(patch).length) {
        const res = await updateSymbol(symbolId, patch);
        if (!res.ok) {
          setError(res.message || "save failed");
          return;
        }
      }
    }
    onSaved();
    close();
  }

  const title = `Symbol: ${isNew ? draft.path || "New" : draft?.path || "…"}`;

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
          className="sym-config-window"
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
              <button type="button" className="config-help" disabled>
                Help
              </button>
            </div>
          }
        >
          {draft &&
            TABS.map(({ id, Panel }) => (
              <div key={id} className={`config-panel${activeTab === id ? " active" : ""}`}>
                <Panel s={draft} set={set} isNew={isNew} />
              </div>
            ))}
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

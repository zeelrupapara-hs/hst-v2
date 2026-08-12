import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { SymbolTreeSelectField } from "@/components/ui/SymbolTreeSelectField.jsx";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";
import { useSymbols } from "@/hooks/useSymbols.js";
import { throwQuote } from "@/api/endpoints/quotes.js";
import { price } from "@/lib/format.js";

/** Throw a manual quote into the feed, MT5's Quotes dialog (F4). */
export function QuotesDialog({ symbol: initialSymbol, onClose }) {
  const { symbols } = useSymbols();
  const ticks = useMarketFeed();
  const [symbol, setSymbol] = useState(initialSymbol);
  const digits = symbols.find((s) => s.symbol === symbol)?.digits ?? 5;
  const tick = ticks.get(symbol);
  const [bid, setBid] = useState("");
  const [ask, setAsk] = useState("");
  const [status, setStatus] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag("quotes");
  const close = useDialogStack(onClose);

  const point = Math.pow(10, -digits);

  // Update inserts the current prices, and a symbol change starts from them
  const update = () => {
    const t = ticks.get(symbol);
    setBid(t ? t.bid.toFixed(digits) : "");
    setAsk(t ? t.ask.toFixed(digits) : "");
  };
  useEffect(update, [symbol]);

  // Shift = 5 points, Ctrl/Cmd = 10, Ctrl+Shift = 50 — MT5's modifiers
  const mult = (e) => ((e.ctrlKey || e.metaKey) && e.shiftKey ? 50 : e.ctrlKey || e.metaKey ? 10 : e.shiftKey ? 5 : 1);
  const nudge = (setter, dir, m) =>
    setter((v) => (Math.max(0, (Number(v) || 0) + dir * m * point)).toFixed(digits));
  const step = (setter) => (dir) => (e) => nudge(setter, dir, mult(e));
  // Up and Down move both prices together, one point (with the same modifiers)
  const both = (dir) => (e) => {
    const m = mult(e);
    nudge(setBid, dir, m);
    nudge(setAsk, dir, m);
  };

  const Steppers = ({ onStep }) => (
    <span className="trade-steppers">
      <button type="button" tabIndex={-1} onClick={onStep(1)}>▲</button>
      <button type="button" tabIndex={-1} onClick={onStep(-1)}>▼</button>
    </span>
  );

  async function handleSend() {
    const b = Number(bid);
    const a = Number(ask);
    if (!(b > 0) || !(a >= b)) {
      setStatus("need ask >= bid > 0");
      return;
    }
    const res = await throwQuote(symbol, b, a);
    setStatus(res.ok ? "quote accepted" : res.message || "quote rejected");
  }

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
          width={430}
          height={300}
          title="Quotes"
          footer={
            <div className="config-actions">
              {status && <span className="login-error">{status}</span>}
              <button type="button" className="config-ok" onClick={handleSend}>Send</button>
              <button type="button" onClick={close}>Close</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="form-grid">
              <SymbolTreeSelectField label="Symbol" value={symbol} onChange={setSymbol} />
              <label>Bid</label>
              <span className="trade-stepper-wrap">
                <input type="text" value={bid} autoFocus onChange={(e) => setBid(e.target.value)} />
                <Steppers onStep={step(setBid)} />
              </span>
              <label>Ask</label>
              <span className="trade-stepper-wrap">
                <input type="text" value={ask} onChange={(e) => setAsk(e.target.value)} />
                <Steppers onStep={step(setAsk)} />
              </span>
              <label>Current</label>
              <span className="grp-suffix">
                {tick ? `${price(tick.bid, digits)} / ${price(tick.ask, digits)}` : "no quote"}
              </span>
            </div>
            <div className="quotes-updown">
              <button type="button" className="quotes-down" onClick={both(-1)}>Down</button>
              <button type="button" onClick={update}>Update</button>
              <button type="button" className="quotes-up" onClick={both(1)}>Up</button>
            </div>
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

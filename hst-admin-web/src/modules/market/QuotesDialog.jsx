import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";
import { throwQuote } from "@/api/endpoints/quotes.js";
import { price } from "@/lib/format.js";

/** Throw a manual quote into the feed for one symbol, MT5's Quotes dialog. */
export function QuotesDialog({ symbol, digits, onClose }) {
  const ticks = useMarketFeed();
  const tick = ticks.get(symbol);
  const [bid, setBid] = useState(tick ? tick.bid.toFixed(digits) : "");
  const [ask, setAsk] = useState(tick ? tick.ask.toFixed(digits) : "");
  const [status, setStatus] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(`quotes-${symbol}`);
  const close = useDialogStack(onClose);

  const point = Math.pow(10, -digits);

  // Shift = 5 points, Ctrl/Cmd = 10 points — MT5's modifiers
  const step = (setter) => (dir) => (e) => {
    const mult = e.ctrlKey || e.metaKey ? 10 : e.shiftKey ? 5 : 1;
    setter((v) => (Math.max(0, (Number(v) || 0) + dir * mult * point)).toFixed(digits));
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
          width={380}
          height={240}
          title={`Quotes: ${symbol}`}
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
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { bulkClose } from "@/api/endpoints/trades.js";
import { useSymbols } from "@/hooks/useSymbols.js";
import { SymbolTreeSelectField } from "@/components/ui/SymbolTreeSelectField.jsx";
import { GroupTreeSelect } from "@/components/ui/GroupTreeSelect.jsx";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";
import { money } from "@/lib/format.js";
import { OrderType_name } from "@/constants/trades.js";

const px = (v, d = 5) => (v ? Number(v).toFixed(d) : "0.000");

/** MT5's Close All Positions By Symbol: pick the selection, tick what to do, preview fills itself. */
export function BulkCloseDialog({ initialSymbol, onClose, onDone }) {
  const { symbols } = useSymbols();
  const ticks = useMarketFeed();
  const [symbol, setSymbol] = useState(initialSymbol || "");
  const [mask, setMask] = useState("*");
  const [comment, setComment] = useState("");
  const [doClose, setDoClose] = useState(true);
  const [doOrders, setDoOrders] = useState(false);
  const [doSltp, setDoSltp] = useState(false);
  const [preview, setPreview] = useState(null);
  const [results, setResults] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag("bulk-close");
  const close = useDialogStack(onClose);

  useEffect(() => {
    if (!symbol && symbols.length) {
      setSymbol((symbols.find((s) => s.symbol === "EURUSD") ?? symbols[0]).symbol);
    }
  }, [symbols]);

  const digits = symbols.find((s) => s.symbol === symbol)?.digits ?? 5;
  const tick = ticks.get(symbol);
  const modes = [
    ...(doClose ? ["close_positions"] : []),
    ...(doOrders ? ["delete_orders"] : []),
    ...(doSltp ? ["clear_sltp"] : []),
  ];

  // the preview shows positions and/or pending orders, following the ticked operations
  async function refresh(sym = symbol, m = mask) {
    setError("");
    setResults(null);
    const rows = [];
    if (doClose || doSltp) {
      const res = await bulkClose({ symbol: sym, group_mask: m || "*", mode: "close_positions", preview: true });
      if (!res.ok) {
        setError(res.message || "preview failed");
        return;
      }
      rows.push(...(res.data || []).map((r) => ({ ...r, kind: "position" })));
    }
    if (doOrders) {
      const res = await bulkClose({ symbol: sym, group_mask: m || "*", mode: "delete_orders", preview: true });
      if (res.ok) rows.push(...(res.data || []).map((r) => ({ ...r, kind: "order" })));
    }
    setPreview(rows);
  }

  useEffect(() => {
    if (symbol) refresh();
  }, [symbol, mask, doClose, doOrders, doSltp]);

  async function handleProcess() {
    if (!modes.length) {
      setError("Tick at least one operation");
      return;
    }
    if (!window.confirm(`${modes.join(" + ")} over ${preview?.length ?? 0} positions on ${symbol} (${mask}). Are you sure?`)) return;
    setError("");
    const all = [];
    for (const mode of modes) {
      const res = await bulkClose({ symbol, group_mask: mask || "*", mode, comment });
      if (!res.ok) {
        setError(res.message || `${mode} failed`);
        return;
      }
      all.push(...(res.data || []).map((r) => ({ ...r, mode })));
    }
    setResults(all);
    setPreview(null);
    onDone?.();
  }

  return (
    <DialogOverlay>
      <div className="dialog-positioner" style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}>
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          width={660}
          height={520}
          className="balance-dialog"
          title="Close All Positions By Symbol"
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleProcess}>Close</button>
              <button type="button" onClick={close}>Cancel</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="bulk-close-head">
              <div className="form-grid bulk-close-form">
                <SymbolTreeSelectField label="Symbol" value={symbol} onChange={setSymbol} />
                <label>Bid / Ask</label>
                <span className="bulk-close-prices">
                  <input type="text" readOnly value={px(tick?.bid, digits)} />
                  <input type="text" readOnly value={px(tick?.ask, digits)} />
                  <button type="button" onClick={() => refresh()}>Update</button>
                </span>
                <label>Groups</label>
                <GroupTreeSelect maskable value={mask} onChange={setMask} onPick={setMask} />
                <label>Comment</label>
                <input type="text" maxLength={31} value={comment} onChange={(e) => setComment(e.target.value)} />
              </div>
              <div className="bulk-close-checks">
                <label><input type="checkbox" checked={doClose} onChange={(e) => setDoClose(e.target.checked)} /> Close positions</label>
                <label><input type="checkbox" checked={doOrders} onChange={(e) => setDoOrders(e.target.checked)} /> Delete pending orders</label>
                <label><input type="checkbox" checked={doSltp} onChange={(e) => setDoSltp(e.target.checked)} /> Clear Stop Loss and Take Profit</label>
              </div>
            </div>
            <div className="bulk-results">
              {results ? (
                results.length ? (
                  results.map((r, i) => (
                    <div key={i} className={r.ok ? "acc-profit" : "acc-loss"}>
                      {r.mode}: {r.login} #{r.id} — {r.ok ? "✓" : r.message || "failed"}
                    </div>
                  ))
                ) : (
                  <div className="df-empty">No rows matched</div>
                )
              ) : (
                <table className="data-table data-table-grid">
                  <thead>
                    <tr>
                      <th>Login</th><th>Type</th><th>Volume</th><th>Price</th>
                      <th>S / L</th><th>T / P</th><th>Price</th><th>Swap</th><th>Profit</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(preview || []).map((r, i) => (
                      <tr key={i}>
                        <td>{r.login}</td>
                        <td>{r.kind === "order" ? (OrderType_name[r.action] ?? r.action) : r.action === 0 ? "buy" : "sell"}</td>
                        <td>{(r.volume ?? 0).toFixed(2)}</td>
                        <td>{px(r.price, digits)}</td>
                        <td>{px(r.price_sl, digits)}</td>
                        <td>{px(r.price_tp, digits)}</td>
                        <td>{px(r.action === 0 ? tick?.bid : tick?.ask, digits)}</td>
                        <td>{money(r.storage)}</td>
                        <td className={r.profit < 0 ? "acc-loss" : "acc-profit"}>{money(r.profit)}</td>
                      </tr>
                    ))}
                    {preview !== null && !preview.length && (
                      <tr><td colSpan={9} className="df-empty">No positions match</td></tr>
                    )}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

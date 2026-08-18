import { useEffect, useRef, useState } from "react";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { SymbolTreeSelectField } from "@/components/ui/SymbolTreeSelectField.jsx";
import {
  closePosition,
  createOrder,
  fetchAccountPositions,
  modifyPosition,
} from "@/api/endpoints/trades.js";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";
import { useSymbols } from "@/hooks/useSymbols.js";
import { onEvent } from "@/api/socket.js";
import { money } from "@/lib/format.js";

const px = (v, digits = 5) => (v == null ? "—" : Number(v).toFixed(digits));

const ORDER_TYPES = [
  { value: "market", label: "Market Order" },
  { value: "2", label: "Buy Limit" },
  { value: "3", label: "Sell Limit" },
  { value: "4", label: "Buy Stop" },
  { value: "5", label: "Sell Stop" },
];

const FILL_POLICIES = [
  { value: "0", label: "Fill or Kill" },
  { value: "1", label: "Immediate or Cancel" },
  { value: "2", label: "Return" },
];

/** The MT5 tick chart: the last minutes of bid/ask, stepped, with price chips at the edge. */
function TickChart({ symbol, tick, digits }) {
  const canvasRef = useRef(null);
  const points = useRef([]);
  const lastSymbol = useRef(symbol);

  if (lastSymbol.current !== symbol) {
    lastSymbol.current = symbol;
    points.current = [];
  }

  useEffect(() => {
    if (tick?.bid) {
      const last = points.current.at(-1);
      if (!last || last.bid !== tick.bid || last.ask !== tick.ask) {
        points.current.push({ bid: tick.bid, ask: tick.ask });
        if (points.current.length > 160) points.current.shift();
      }
    }

    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    const { width: w, height: h } = canvas;
    const pts = points.current;

    ctx.fillStyle = "#fff";
    ctx.fillRect(0, 0, w, h);
    ctx.strokeStyle = "#000";
    ctx.strokeRect(0.5, 0.5, w - 1, h - 1);
    ctx.font = "10px sans-serif";
    ctx.fillStyle = "#000";
    ctx.fillText(symbol, 6, 14);
    if (pts.length < 2) return;

    let lo = Infinity;
    let hi = -Infinity;
    for (const p of pts) {
      lo = Math.min(lo, p.bid);
      hi = Math.max(hi, p.ask);
    }
    const pad = (hi - lo) * 0.15 || Math.pow(10, -digits);
    lo -= pad;
    hi += pad;
    const X = (i) => 4 + (i * (w - 46)) / (pts.length - 1);
    const Y = (v) => h - 16 - ((v - lo) * (h - 34)) / (hi - lo);

    // dotted midlines, the MT5 grid whisper
    ctx.strokeStyle = "#ddd";
    ctx.setLineDash([2, 3]);
    for (const f of [0.25, 0.5, 0.75]) {
      const y = 18 + f * (h - 34);
      ctx.beginPath();
      ctx.moveTo(4, y);
      ctx.lineTo(w - 42, y);
      ctx.stroke();
    }
    ctx.setLineDash([]);

    const step = (key, color) => {
      ctx.strokeStyle = color;
      ctx.beginPath();
      pts.forEach((p, i) => {
        const x = X(i);
        const y = Y(p[key]);
        if (i === 0) ctx.moveTo(x, y);
        else {
          ctx.lineTo(x, Y(pts[i - 1][key]));
          ctx.lineTo(x, y);
        }
      });
      ctx.stroke();
    };
    step("bid", "#c0392b");
    step("ask", "#2465b4");

    const chip = (v, color) => {
      const y = Math.min(h - 12, Math.max(12, Y(v)));
      ctx.fillStyle = color;
      ctx.fillRect(w - 44, y - 7, 42, 13);
      ctx.fillStyle = "#fff";
      ctx.fillText(Number(v).toFixed(digits), w - 42, y + 3);
    };
    const last = pts.at(-1);
    chip(last.bid, "#c0392b");
    chip(last.ask, "#2465b4");
  }, [tick, symbol, digits]);

  return <canvas ref={canvasRef} width={250} height={380} className="trade-chart" />;
}

/** The dealer trades for the client from inside the account dialog; every deal carries the dealer. */
export function AccountTradeTab({ login }) {
  const { symbols } = useSymbols();
  const ticks = useMarketFeed();
  const [symbol, setSymbol] = useState("");
  const [type, setType] = useState("market");
  const [form, setForm] = useState({ volume: "0.01", price: "", sl: "", tp: "", comment: "" });
  const [fill, setFill] = useState("0");
  const [autoPrice, setAutoPrice] = useState(true);
  const [positions, setPositions] = useState(null);
  const [editing, setEditing] = useState(null);
  const [error, setError] = useState("");
  const { confirm, confirmElement } = useConfirm();
  const [notice, setNotice] = useState("");

  const reload = () => fetchAccountPositions(login).then((res) => res.ok && setPositions(res.data || []));

  useEffect(() => {
    reload();
    // the engine speaks when a position opens, changes or closes — refresh on its word
    const offs = ["position_create", "position_update", "position_close"].map((t) =>
      onEvent(t, reload),
    );
    return () => offs.forEach((off) => off?.());
  }, [login]);

  useEffect(() => {
    if (!symbol && symbols.length) {
      const first = symbols.find((s) => s.symbol === "EURUSD") ?? symbols[0];
      setSymbol(first.symbol);
    }
  }, [symbols]);

  const spec = symbols.find((s) => s.symbol === symbol);
  const digits = spec?.digits ?? 5;
  const tick = ticks.get(symbol);
  const isMarket = type === "market";
  const pending = !isMarket;
  const buySide = type === "2" || type === "4";
  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));

  // Auto keeps At Price on the market, exactly MT5's checkbox
  useEffect(() => {
    if (autoPrice && tick) {
      setForm((f) => ({ ...f, price: Number(buySide || isMarket ? tick.ask : tick.bid).toFixed(digits) }));
    }
  }, [tick?.bid, tick?.ask, autoPrice, type]);

  const notional = () => {
    const lots = Number(form.volume) || 0;
    const size = spec?.contract_size || 100000;
    return `${Math.round(lots * size).toLocaleString("en-US").replace(/,/g, " ")} ${symbol.slice(0, 3)}`;
  };

  async function send(orderType) {
    setError("");
    setNotice("");
    const res = await createOrder({
      login,
      symbol,
      type: orderType,
      volume: Number(form.volume) || 0,
      price: pending ? Number(form.price) || 0 : 0,
      price_sl: Number(form.sl) || 0,
      price_tp: Number(form.tp) || 0,
      type_fill: Number(fill),
      comment: form.comment.slice(0, 31),
    });
    if (!res.ok) {
      setError(res.message || "order refused");
      return;
    }
    setNotice(res.data?.message || "order requested");
    setTimeout(reload, 700);
  }

  async function close(p) {
    if (!(await confirm({ title: "Trade", message: `Close position #${p.position_id} (${p.volume.toFixed(2)} ${p.symbol})?` }))) return;
    const res = await closePosition(p.position_id, { login, position_id: p.position_id });
    if (!res.ok) window.alert(res.message || "close refused");
    setTimeout(reload, 700);
  }

  async function saveLevels() {
    const res = await modifyPosition(editing.position_id, {
      login,
      position_id: editing.position_id,
      price_sl: Number(editing.sl) || 0,
      price_tp: Number(editing.tp) || 0,
    });
    if (!res.ok) {
      window.alert(res.message || "modify refused");
      return;
    }
    setEditing(null);
    setTimeout(reload, 700);
  }

  const stepper = (key, step) => (dir) =>
    setForm((f) => ({ ...f, [key]: (Math.max(0, (Number(f[key]) || 0) + dir * step)).toFixed(key === "volume" ? 2 : digits) }));

  const Steppers = ({ onStep }) => (
    <span className="trade-steppers">
      <button type="button" tabIndex={-1} onClick={() => onStep(1)}>▲</button>
      <button type="button" tabIndex={-1} onClick={() => onStep(-1)}>▼</button>
    </span>
  );

  const point = Math.pow(10, -digits);

  return (
    <>
      <div className="trade-layout">
        <TickChart symbol={symbol} tick={tick} digits={digits} />
        <div className="trade-form">
          <div className="trade-grid">
            <SymbolTreeSelectField label="Symbol:" value={symbol} onChange={setSymbol} />
            <label>Type:</label>
            <PropSelect fill value={type} options={ORDER_TYPES} onChange={setType} />
            <label>Volume:</label>
            <span className="trade-stepper-wrap">
              <input type="text" value={form.volume} onChange={set("volume")} />
              <Steppers onStep={stepper("volume", 0.01)} />
              <span className="grp-suffix">{notional()}</span>
            </span>
            <label>Fill Policy:</label>
            <PropSelect value={fill} options={FILL_POLICIES} onChange={setFill} />
            <label>At Price:</label>
            <span className="trade-stepper-wrap">
              <input type="text" value={form.price} disabled={autoPrice} onChange={set("price")} />
              <Steppers onStep={stepper("price", point)} />
              <label className="trade-auto">
                <input type="checkbox" checked={autoPrice} onChange={(e) => setAutoPrice(e.target.checked)} />
                Auto
              </label>
            </span>
            <label>Stop Loss:</label>
            <span className="trade-stepper-wrap">
              <input type="text" value={form.sl} onChange={set("sl")} placeholder={Number(0).toFixed(digits)} />
              <Steppers onStep={stepper("sl", point)} />
              <label className="trade-tp-label">Take Profit:</label>
              <input type="text" value={form.tp} onChange={set("tp")} placeholder={Number(0).toFixed(digits)} />
              <Steppers onStep={stepper("tp", point)} />
            </span>
            <label>Comment:</label>
            <input type="text" maxLength={31} value={form.comment} onChange={set("comment")} />
          </div>

          <div className="trade-quote">
            <span className="trade-bid">{px(tick?.bid, digits)}</span>
            <span className="trade-sep"> / </span>
            <span className="trade-ask">{px(tick?.ask, digits)}</span>
          </div>

          {isMarket ? (
            <div className="trade-buttons">
              <button type="button" className="trade-sell" disabled={!tick} onClick={() => send(1)}>
                Sell at {px(tick?.bid, digits)}
              </button>
              <button type="button" className="trade-buy" disabled={!tick} onClick={() => send(0)}>
                Buy at {px(tick?.ask, digits)}
              </button>
            </div>
          ) : (
            <div className="trade-buttons">
              <button
                type="button"
                className={buySide ? "trade-buy" : "trade-sell"}
                disabled={!Number(form.price)}
                onClick={() => send(Number(type))}
              >
                Place {ORDER_TYPES.find((t) => t.value === type)?.label} at {form.price || "—"}
              </button>
            </div>
          )}

          <div className="trade-status">
            {error ? <span className="login-error">{error}</span> : <span>{notice}</span>}
          </div>
        </div>
      </div>

      <table className="data-table data-table-grid dealing-context-positions">
        <thead>
          <tr>
            <th>Ticket</th><th>Symbol</th><th>Type</th><th>Volume</th><th>Open</th>
            <th>S / L</th><th>T / P</th><th>Profit</th><th></th>
          </tr>
        </thead>
        <tbody>
          {(positions || []).map((p) =>
            editing?.position_id === p.position_id ? (
              <tr key={p.position_id} className="selected">
                <td>{p.position_id}</td>
                <td>{p.symbol}</td>
                <td>{p.action === 0 ? "buy" : "sell"}</td>
                <td>{p.volume.toFixed(2)}</td>
                <td>{px(p.price_open, digits)}</td>
                <td><input type="text" style={{ width: 70 }} value={editing.sl}
                  onChange={(e) => setEditing({ ...editing, sl: e.target.value })} /></td>
                <td><input type="text" style={{ width: 70 }} value={editing.tp}
                  onChange={(e) => setEditing({ ...editing, tp: e.target.value })} /></td>
                <td className={p.profit < 0 ? "acc-loss" : "acc-profit"}>{money(p.profit)}</td>
                <td>
                  <button type="button" onClick={saveLevels}>Save</button>{" "}
                  <button type="button" onClick={() => setEditing(null)}>Cancel</button>
                </td>
              </tr>
            ) : (
              <tr key={p.position_id}>
                <td>{p.position_id}</td>
                <td>{p.symbol}</td>
                <td>{p.action === 0 ? "buy" : "sell"}</td>
                <td>{p.volume.toFixed(2)}</td>
                <td>{px(p.price_open, digits)}</td>
                <td>{p.price_sl ? px(p.price_sl, digits) : "—"}</td>
                <td>{p.price_tp ? px(p.price_tp, digits) : "—"}</td>
                <td className={p.profit < 0 ? "acc-loss" : "acc-profit"}>{money(p.profit)}</td>
                <td>
                  <button type="button" onClick={() => setEditing({
                    position_id: p.position_id, sl: p.price_sl || "", tp: p.price_tp || "",
                  })}>Modify</button>{" "}
                  <button type="button" onClick={() => close(p)}>Close</button>
                </td>
              </tr>
            ),
          )}
          {positions !== null && !positions.length && (
            <tr><td colSpan={9} className="df-empty">No open positions</td></tr>
          )}
        </tbody>
      </table>
      {confirmElement}
    </>
  );
}

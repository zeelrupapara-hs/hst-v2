import { useEffect, useRef, useState } from "react";
import { onEvent } from "@/api/socket.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import {
  confirmRequest,
  connectDealer,
  disconnectDealer,
  fetchDealerState,
  fetchDealingQueue,
  heartbeatDealer,
  rejectRequest,
  requoteRequest,
} from "@/api/endpoints/dealing.js";
import { OrderState_name, OrderType_name } from "@/constants/trades.js";
import { fetchAccountPositions } from "@/api/endpoints/trades.js";
import { fetchUser } from "@/api/endpoints/users.js";
import { formatNs } from "@/lib/time.js";
import { money } from "@/lib/format.js";
import { useLiveAccount } from "@/hooks/useLiveAccounts.js";
import { useMarketFeed } from "@/hooks/useMarketFeed.js";

const REQUEST_STATE = { 7: "new order", 8: "modification", 9: "cancellation" };
const ANSWER_WINDOW_SECONDS = 30;

// seconds the desk still has before the engine rejects the request itself
const secondsLeft = (timeSetupNs, now) =>
  Math.max(0, ANSWER_WINDOW_SECONDS - Math.floor((now - timeSetupNs / 1e6) / 1000));

/** The client behind the selected request: live money above, open positions below. */
function AccountContext({ login, symbol }) {
  const [positions, setPositions] = useState(null);
  const [stored, setStored] = useState(null);
  const live = useLiveAccount(login);

  useEffect(() => {
    setPositions(null);
    setStored(null);
    fetchAccountPositions(login).then((res) => res.ok && setPositions(res.data || []));
    // a flat account never speaks on the live stream, so its stored money stands in
    fetchUser(login).then((res) => res.ok && setStored(res.data));
  }, [login]);

  return (
    <div className="dealing-context">
      <div className="dealing-context-money">
        <span>Account <b>{login}</b></span>
        {live ? (
          <>
            <span>Balance: <b>{money(live.balance)}</b></span>
            <span>Equity: <b>{money(live.equity)}</b></span>
            <span>Free: <b>{money(live.free)}</b></span>
            <span>Level: <b>{live.margin > 0 ? `${live.level.toFixed(2)} %` : "—"}</b></span>
            <span className={live.profit < 0 ? "acc-loss" : "acc-profit"}>P/L: <b>{money(live.profit)}</b></span>
          </>
        ) : stored ? (
          <>
            <span>Balance: <b>{money(stored.balance)}</b></span>
            <span>Credit: <b>{money(stored.credit)}</b></span>
            <span className="grp-suffix">stored values — no live line yet</span>
          </>
        ) : null}
      </div>
      <table className="data-table data-table-grid dealing-context-positions">
        <thead>
          <tr><th>Ticket</th><th>Symbol</th><th>Type</th><th>Volume</th><th>Open</th><th>Profit</th></tr>
        </thead>
        <tbody>
          {(positions || []).map((p) => (
            <tr key={p.position_id} className={p.symbol === symbol ? "selected" : ""}>
              <td>{p.position_id}</td>
              <td>{p.symbol}</td>
              <td>{p.action === 0 ? "buy" : "sell"}</td>
              <td>{(p.volume ?? 0).toFixed(2)}</td>
              <td>{p.price_open}</td>
              <td className={p.profit < 0 ? "acc-loss" : "acc-profit"}>{money(p.profit)}</td>
            </tr>
          ))}
          {positions !== null && !positions.length && (
            <tr><td colSpan={6} className="df-empty">No open positions</td></tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

/** Confirm at request/market price, requote at a named price, or reject with a reason. */
function ActionDialog({ action, row, marketPrice, onClose, onDone }) {
  const [value, setValue] = useState(
    action === "reject" ? "" : String(marketPrice ?? row.price_current ?? ""),
  );
  const [error, setError] = useState("");
  const close = useDialogStack(onClose);

  async function run() {
    let res;
    if (action === "confirm") res = await confirmRequest(String(row.order_id), row.login, Number(value) || 0);
    else if (action === "requote") {
      if (!Number(value)) {
        setError("A requote needs a price");
        return;
      }
      res = await requoteRequest(String(row.order_id), row.login, Number(value));
    } else res = await rejectRequest(String(row.order_id), row.login, value.slice(0, 31));
    if (!res.ok) {
      setError(res.message || "request failed");
      return;
    }
    onDone();
    close();
  }

  const label = { confirm: "Price (0 = request price)", requote: "New price", reject: "Reason" }[action];

  return (
    <DialogOverlay className="sym-session-dialog-overlay" onClose={onClose}>
      <div className="sym-session-dialog dealing-action" role="dialog">
        <div className="sym-session-dialog-title">
          {action[0].toUpperCase() + action.slice(1)}: order #{row.order_id} — {row.login} {row.symbol}
        </div>
        <div className="sym-session-dialog-body">
          <div className="form-grid">
            <label>{label}</label>
            <input type="text" autoFocus value={value} onChange={(e) => setValue(e.target.value)} onKeyDown={(e) => e.key === "Enter" && run()} />
          </div>
          {error && <p className="login-error">{error}</p>}
        </div>
        <div className="sym-session-dialog-footer">
          <span />
          <div className="sym-session-dialog-actions">
            <button type="button" className="sym-session-ok" onClick={run}>OK</button>
            <button type="button" className="sym-session-cancel" onClick={close}>Cancel</button>
          </div>
        </div>
      </div>
    </DialogOverlay>
  );
}

const AUTO_KINDS = [
  { state: 7, label: "New orders" },
  { state: 8, label: "Modifications" },
  { state: 9, label: "Cancellations" },
];

/** MT5-style automation popover: per-kind enable + max lots (0 = unlimited). */
function AutoPopover({ auto, setAuto }) {
  return (
    <div className="dealing-auto-pop">
      {AUTO_KINDS.map(({ state, label }) => (
        <label key={state} className="dealing-auto-row">
          <input
            type="checkbox"
            checked={auto[state].on}
            onChange={(e) => setAuto({ ...auto, [state]: { ...auto[state], on: e.target.checked } })}
          />
          <span>{label}</span>
          <input
            type="text"
            className="dealing-auto-max"
            value={auto[state].max}
            onChange={(e) => setAuto({ ...auto, [state]: { ...auto[state], max: e.target.value } })}
          />
        </label>
      ))}
      <div className="dealing-auto-hint grp-suffix">max lots, 0 = unlimited — resets on reload</div>
    </div>
  );
}

/** The dealer desk: connect to work the queue, confirm, requote or reject each request. */
export function DealingModule() {
  const [onDesk, setOnDesk] = useState(false);
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [menu, setMenu] = useState(null);
  const [action, setAction] = useState(null);
  const [now, setNow] = useState(Date.now());
  // deliberately not persisted — MT5 automation resets on restart
  const [auto, setAuto] = useState({
    7: { on: false, max: "0" },
    8: { on: false, max: "0" },
    9: { on: false, max: "0" },
  });
  const [autoOpen, setAutoOpen] = useState(false);
  const [autoBusy, setAutoBusy] = useState(() => new Set());
  const autoDone = useRef(new Set());
  const heartbeat = useRef(0);
  const ticks = useMarketFeed();

  // auto-confirm freshly loaded rows whose kind is armed and volume fits
  useEffect(() => {
    for (const r of rows || []) {
      const cfg = auto[r.state];
      if (!cfg?.on || autoDone.current.has(r.order_id)) continue;
      const max = Number(cfg.max) || 0;
      if (max > 0 && (r.volume ?? 0) > max) continue;
      autoDone.current.add(r.order_id);
      setAutoBusy((s) => new Set(s).add(r.order_id));
      const price = ticks.get(r.symbol)?.[r.type === 1 ? "bid" : "ask"] ?? r.price_current ?? 0;
      confirmRequest(String(r.order_id), r.login, Number(price) || 0).then(() => {
        setAutoBusy((s) => { const n = new Set(s); n.delete(r.order_id); return n; });
        load();
      });
    }
  }, [rows, auto]);

  // the countdown column ticks once a second
  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);

  const load = () => fetchDealingQueue().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    fetchDealerState().then((res) => res.ok && setOnDesk(!!res.data?.online));
    load();
    const off = onEvent("*", (e) => {
      if (/^dealer|^order_dealer/.test(e.type)) load();
    });
    return () => off();
  }, []);

  // the desk drops dealers that stop breathing
  useEffect(() => {
    clearInterval(heartbeat.current);
    if (onDesk) heartbeat.current = setInterval(heartbeatDealer, 15000);
    return () => clearInterval(heartbeat.current);
  }, [onDesk]);

  async function toggleDesk() {
    const res = onDesk ? await disconnectDealer() : await connectDealer();
    if (!res.ok) {
      window.alert(res.message || "desk unavailable");
      return;
    }
    setOnDesk(!onDesk);
    load();
  }

  const row = selected != null ? rows?.[selected] : null;

  return (
    <div className="module-root">
      <div className="dealing-bar">
        <Icon id="dealing" />
        <span className={onDesk ? "acc-profit" : "grp-suffix"}>
          {onDesk ? "On the desk — requests routed to you" : "Off the desk — watching the queue"}
        </span>
        <button type="button" onClick={toggleDesk}>{onDesk ? "Leave desk" : "Join desk"}</button>
        <button type="button" onClick={load}>Refresh</button>
        <span className="dealing-auto-anchor">
          <button
            type="button"
            className={Object.values(auto).some((c) => c.on) ? "acc-profit" : ""}
            onClick={() => setAutoOpen(!autoOpen)}
          >
            Automation
          </button>
          {autoOpen && <AutoPopover auto={auto} setAuto={setAuto} />}
        </span>
      </div>
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Request</th>
              <th>Login</th>
              <th>Symbol</th>
              <th>Type</th>
              <th>Volume</th>
              <th>Price</th>
              <th>Current</th>
              <th>Time</th>
              <th>Left</th>
              <th>Comment</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((r, i) => (
              <tr
                key={r.order_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => setAction("confirm")}
              >
                <td>
                  {REQUEST_STATE[r.state] ?? OrderState_name[r.state]}
                  {autoBusy.has(r.order_id) && <span className="grp-suffix dealing-auto-tag">auto</span>}
                </td>
                <td>{r.login}</td>
                <td>{r.symbol}</td>
                <td>{OrderType_name[r.type] ?? r.type}</td>
                <td>{(r.volume ?? 0).toFixed(2)}</td>
                <td>{r.price_order || ""}</td>
                <td>{r.price_current || ""}</td>
                <td>{formatNs(r.time_setup)}</td>
                <td className={secondsLeft(r.time_setup, now) <= 10 ? "acc-loss" : ""}>
                  {secondsLeft(r.time_setup, now)}s
                </td>
                <td>{r.comment}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr>
                <td colSpan={10} className="df-empty">The queue is empty</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {row && <AccountContext login={row.login} symbol={row.symbol} />}
      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            { label: "Confirm", disabled: !row, onClick: () => setAction("confirm") },
            { label: "Requote", disabled: !row, onClick: () => setAction("requote") },
            "sep",
            { label: "Reject", disabled: !row, onClick: () => setAction("reject") },
          ]}
        />
      )}
      {action && row && (
        <ActionDialog
          action={action}
          row={row}
          marketPrice={ticks.get(row.symbol)?.[row.type === 1 ? "bid" : "ask"]}
          onClose={() => setAction(null)}
          onDone={load}
        />
      )}
    </div>
  );
}

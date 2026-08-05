import { useEffect, useRef, useState } from "react";
import { onEvent } from "@/api/socket.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
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
import { formatNs } from "@/lib/time.js";

const REQUEST_STATE = { 7: "new order", 8: "modification", 9: "cancellation" };

/** Confirm at request/market price, requote at a named price, or reject with a reason. */
function ActionDialog({ action, row, onClose, onDone }) {
  const [value, setValue] = useState(action === "requote" ? String(row.price_current || "") : "");
  const [error, setError] = useState("");

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
    onClose();
  }

  const label = { confirm: "Price (0 = request price)", requote: "New price", reject: "Reason" }[action];

  return (
    <div className="sym-session-dialog-overlay" onClick={onClose}>
      <div className="sym-session-dialog dealing-action" role="dialog" onClick={(e) => e.stopPropagation()}>
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
            <button type="button" className="sym-session-cancel" onClick={onClose}>Cancel</button>
          </div>
        </div>
      </div>
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
  const heartbeat = useRef(0);

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
                <td>{REQUEST_STATE[r.state] ?? OrderState_name[r.state]}</td>
                <td>{r.login}</td>
                <td>{r.symbol}</td>
                <td>{OrderType_name[r.type] ?? r.type}</td>
                <td>{(r.volume ?? 0).toFixed(2)}</td>
                <td>{r.price_order || ""}</td>
                <td>{r.price_current || ""}</td>
                <td>{formatNs(r.time_setup)}</td>
                <td>{r.comment}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr>
                <td colSpan={9} className="df-empty">The queue is empty</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
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
        <ActionDialog action={action} row={row} onClose={() => setAction(null)} onDone={load} />
      )}
    </div>
  );
}

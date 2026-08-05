import { useEffect, useState } from "react";
import { fetchAccountOrders, fetchAccountPositions } from "@/api/endpoints/trades.js";
import { OrderType_name } from "@/constants/trades.js";
import { formatNs } from "@/lib/time.js";

const px = (v, digits = 5) => (v ? v.toFixed(digits) : "");
const money = (v) => (v ?? 0).toFixed(2);

/** Overview: open positions, the account state row, pending orders — as the reference lays it out. */
export function AccountOverviewTab({ login, user }) {
  const [positions, setPositions] = useState(null);
  const [orders, setOrders] = useState(null);

  useEffect(() => {
    fetchAccountPositions(login).then((res) => res.ok && setPositions(res.data || []));
    fetchAccountOrders(login).then((res) => res.ok && setOrders(res.data || []));
  }, [login]);

  const floating = (positions || []).reduce((s, p) => s + (p.profit || 0), 0);

  return (
    <div className="acc-overview">
      <div className="acc-grid-title">Open Positions</div>
      <div className="table-wrap acc-grid">
        <table className="data-table data-table-grid df-sub-table">
          <thead>
            <tr>
              <th>Symbol</th>
              <th>Ticket</th>
              <th>Time</th>
              <th>Type</th>
              <th>Volume</th>
              <th>Price</th>
              <th>S / L</th>
              <th>T / P</th>
              <th>Price</th>
              <th>Profit</th>
            </tr>
          </thead>
          <tbody>
            {(positions || []).map((p) => (
              <tr key={p.position_id}>
                <td>{p.symbol}</td>
                <td>{p.position_id}</td>
                <td>{formatNs(p.time_create)}</td>
                <td>{p.action === 0 ? "buy" : "sell"}</td>
                <td>{(p.volume ?? 0).toFixed(2)}</td>
                <td>{px(p.price_open, p.digits)}</td>
                <td>{px(p.price_sl, p.digits)}</td>
                <td>{px(p.price_tp, p.digits)}</td>
                <td>{px(p.price_current, p.digits)}</td>
                <td className={p.profit < 0 ? "acc-loss" : "acc-profit"}>{money(p.profit)}</td>
              </tr>
            ))}
            {positions?.length === 0 && (
              <tr>
                <td colSpan={10} className="df-empty">No open positions</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <div className="acc-state-row">
        <span>Balance: <b>{money(user?.balance)}</b></span>
        <span>Credit: <b>{money(user?.credit)}</b></span>
        <span>Floating P/L: <b className={floating < 0 ? "acc-loss" : "acc-profit"}>{money(floating)}</b></span>
      </div>

      <div className="acc-grid-title">Pending Orders</div>
      <div className="table-wrap acc-grid">
        <table className="data-table data-table-grid df-sub-table">
          <thead>
            <tr>
              <th>Symbol</th>
              <th>Ticket</th>
              <th>Time</th>
              <th>Type</th>
              <th>Volume</th>
              <th>Price</th>
              <th>S / L</th>
              <th>T / P</th>
            </tr>
          </thead>
          <tbody>
            {(orders || []).map((o) => (
              <tr key={o.order_id}>
                <td>{o.symbol}</td>
                <td>{o.order_id}</td>
                <td>{formatNs(o.time_setup)}</td>
                <td>{OrderType_name[o.type] ?? o.type}</td>
                <td>{(o.volume_initial ?? o.volume ?? 0).toFixed(2)}</td>
                <td>{px(o.price_order, o.digits)}</td>
                <td>{px(o.price_sl, o.digits)}</td>
                <td>{px(o.price_tp, o.digits)}</td>
              </tr>
            ))}
            {orders?.length === 0 && (
              <tr>
                <td colSpan={8} className="df-empty">No pending orders</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

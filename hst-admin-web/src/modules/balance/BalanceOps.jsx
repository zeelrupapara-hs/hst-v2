import { useEffect, useState } from "react";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { createBalance } from "@/api/endpoints/balance.js";
import { fetchAccountDeals } from "@/api/endpoints/trades.js";
import { onEvent } from "@/api/socket.js";
import { useLiveAccount } from "@/hooks/useLiveAccounts.js";
import { formatNs } from "@/lib/time.js";
import { money } from "@/lib/format.js";

// The balance-type deals, as the engine accepts them.
const OPERATIONS = [
  { value: 2, label: "Balance" },
  { value: 3, label: "Credit" },
  { value: 4, label: "Charge" },
  { value: 5, label: "Correction" },
  { value: 6, label: "Bonus" },
  { value: 7, label: "Commission" },
];

const OPERATION_NAMES = {
  2: "balance", 3: "credit", 4: "charge", 5: "correction",
  6: "bonus", 7: "commission", 8: "commission", 9: "commission",
  12: "interest", 19: "so compensation", 20: "so compensation",
};

const isBalanceDeal = (deal) => OPERATION_NAMES[deal.action] !== undefined;

/** The account's balance desk: run an operation, watch the history and the live totals. */
export function BalanceOps({ login }) {
  const [action, setAction] = useState(2);
  const [amount, setAmount] = useState("");
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");
  const [deals, setDeals] = useState(null);
  const live = useLiveAccount(login);

  const loadDeals = () =>
    fetchAccountDeals(login).then((res) => {
      if (res.ok) setDeals((res.data || []).filter(isBalanceDeal));
    });

  useEffect(() => {
    if (!login) return;
    setDeals(null);
    loadDeals();
    // a finished operation announces itself, so the history refreshes on its own
    const off = onEvent("*", (event) => {
      if (event.type === "money_change" || event.type === "deal_create") loadDeals();
    });
    return off;
  }, [login]);

  async function run(sign) {
    setError("");
    const value = Number(amount);
    if (!value || value <= 0) {
      setError("Amount must be a positive number");
      return;
    }
    const res = await createBalance(Number(login), action, sign * value, comment);
    if (!res.ok) {
      setError(res.message || "operation refused");
      return;
    }
    setAmount("");
    loadDeals();
  }

  if (!login) return null;

  return (
    <div className="balance-ops">
      <div className="form-grid balance-ops-form">
        <label>Operation:</label>
        <PropSelect value={action} options={OPERATIONS} onChange={(v) => setAction(Number(v))} />
        <label>Amount:</label>
        <input type="number" min="0" step="0.01" value={amount} onChange={(e) => setAmount(e.target.value)} />
        <label>Comment:</label>
        <input type="text" className="wide" maxLength={64} value={comment} onChange={(e) => setComment(e.target.value)} />
      </div>
      <div className="balance-ops-actions">
        <button type="button" className="balance-deposit" onClick={() => run(1)}>Deposit</button>
        <button type="button" className="balance-withdraw" onClick={() => run(-1)}>Withdrawal</button>
        {error && <span className="login-error">{error}</span>}
      </div>
      <div className="table-wrap balance-ops-history">
        <table className="data-table data-table-grid">
          <thead>
            <tr>
              <th>Time</th>
              <th>Deal</th>
              <th>Type</th>
              <th>Amount</th>
              <th>Comment</th>
            </tr>
          </thead>
          <tbody>
            {(deals || []).map((deal) => (
              <tr key={deal.deal_id}>
                <td>{formatNs(deal.time)}</td>
                <td>{deal.deal_id}</td>
                <td>{OPERATION_NAMES[deal.action]}</td>
                <td className={deal.profit < 0 ? "acc-loss" : "acc-profit"}>{money(deal.profit)}</td>
                <td>{deal.comment}</td>
              </tr>
            ))}
            {deals !== null && !deals.length && (
              <tr><td colSpan={5} className="df-empty">No balance operations</td></tr>
            )}
          </tbody>
        </table>
      </div>
      {live && (
        <div className="balance-ops-totals">
          <span>Balance: <b>{money(live.balance)}</b></span>
          <span>Credit: <b>{money(live.credit)}</b></span>
          <span>Equity: <b>{money(live.equity)}</b></span>
          <span>Margin: <b>{money(live.margin)}</b></span>
          <span>Free: <b>{money(live.free)}</b></span>
          <span>Level: <b>{live.margin > 0 ? `${live.level.toFixed(2)} %` : "—"}</b></span>
        </div>
      )}
    </div>
  );
}

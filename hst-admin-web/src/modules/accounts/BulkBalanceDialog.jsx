import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { bulkBalance } from "@/api/endpoints/balance.js";
import { useConfirm } from "@/hooks/useConfirm.jsx";

// Engine deal actions: deposit and withdrawal are both a balance deal, signed.
const OPS = [
  { value: "deposit", label: "Deposit", action: 2, sign: 1 },
  { value: "withdrawal", label: "Withdrawal", action: 2, sign: -1 },
  { value: "credit", label: "Credit", action: 3, sign: 1 },
  { value: "correction", label: "Correction", action: 5, sign: 1 },
  { value: "bonus", label: "Bonus", action: 6, sign: 1 },
  { value: "commission", label: "Commission", action: 7, sign: 1 },
];

/** Bulk balance operations: one action + amount on many logins, or a login;amount CSV. */
export function BulkBalanceDialog({ logins, onClose, onSaved }) {
  const [op, setOp] = useState("deposit");
  const [amount, setAmount] = useState("");
  const [comment, setComment] = useState("");
  const [csv, setCsv] = useState("");
  const [results, setResults] = useState(null);
  const [error, setError] = useState("");
  const { confirm, confirmElement } = useConfirm();
  const { offset, onTitlePointerDown } = useDialogDrag("bulk-balance");
  const close = useDialogStack(onClose);

  async function handleProcess() {
    setError("");
    let operations;
    const kind = OPS.find((o) => o.value === op);
    if (csv.trim()) {
      operations = csv.trim().split(/\n+/).map((line) => {
        const [login, amt] = line.split(";");
        return { login: Number(login), action: kind.action, amount: Number(amt) * kind.sign };
      });
      const bad = operations.find((o) => !o.login || !o.amount);
      if (bad) {
        setError("CSV must be login;amount per line");
        return;
      }
    } else {
      const value = Number(amount);
      if (!value) {
        setError("Amount is required");
        return;
      }
      operations = logins.map((login) => ({ login, action: kind.action, amount: value * kind.sign }));
    }
    if (!(await confirm({ title: "Balance Operations", message: `Apply ${operations.length} balance operations?` }))) return;
    const res = await bulkBalance({ operations, comment });
    if (!res.ok) {
      setError(res.message || "bulk operation failed");
      return;
    }
    setResults(res.data || []);
    onSaved?.();
  }

  return (
    <DialogOverlay>
      <div className="dialog-positioner" style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}>
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          width={460}
          height={520}
          className="balance-dialog"
          title={`Bulk Balance Operation — ${logins.length} accounts`}
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleProcess}>Process</button>
              <button type="button" onClick={close}>Close</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="balance" size={40} />
              </span>
              <p>
                Books one balance deal per account. Targets: <b>{logins.join(", ")}</b>
              </p>
            </div>
            <div className="form-grid">
              <label>Operation</label>
              <PropSelect fill value={op} options={OPS} onChange={setOp} />
              <label>Amount</label>
              <input type="text" value={amount} onChange={(e) => setAmount(e.target.value)} autoFocus />
              <label>Comment</label>
              <input type="text" maxLength={64} value={comment} onChange={(e) => setComment(e.target.value)} />
              <label>CSV</label>
              <textarea
                className="bulk-csv"
                rows={4}
                placeholder="login;amount per line — overrides the selection and amount when filled"
                value={csv}
                onChange={(e) => setCsv(e.target.value)}
              />
            </div>
            {results && (
              <div className="bulk-results">
                {results.map((r, i) => (
                  <div key={i} className={r.ok ? "acc-profit" : "acc-loss"}>
                    {r.login} — {r.ok ? "✓" : r.message || "failed"}
                  </div>
                ))}
                {!results.length && <div className="df-empty">No results</div>}
              </div>
            )}
          </div>
        </SettingsDialog>
      </div>
      {confirmElement}
    </DialogOverlay>
  );
}

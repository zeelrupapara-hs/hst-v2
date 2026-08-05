import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createCorrection,
  createCredit,
  createDeposit,
  createWithdrawal,
} from "@/api/endpoints/balance.js";

const OPS = {
  deposit: { label: "Deposit", run: createDeposit },
  withdrawal: { label: "Withdrawal", run: createWithdrawal },
  credit: { label: "Credit", run: createCredit },
  correction: { label: "Correction", run: createCorrection },
};

/** The balance operation window: pick the operation, the amount and a deal comment. */
export function BalanceDialog({ user, onClose, onSaved }) {
  const [op, setOp] = useState("deposit");
  const [amount, setAmount] = useState("");
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(user.login);

  async function handleOk() {
    const value = Number(amount);
    if (!value) {
      setError("Amount is required");
      return;
    }
    const res = await OPS[op].run(user.login, value, comment);
    if (!res.ok) {
      setError(res.message || "operation failed");
      return;
    }
    onSaved();
    onClose();
  }

  return (
    <div className="dialog-overlay" onClick={onClose} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          draggable
          onTitlePointerDown={onTitlePointerDown}
          width={420}
          className="balance-dialog"
          title={`Balance: ${user.login} — ${user.name}`}
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={onClose}>Cancel</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="balance" size={40} />
              </span>
              <p>
                The operation books a deal on the account. Current balance:{" "}
                <b>{(user.balance ?? 0).toFixed(2)}</b>, credit <b>{(user.credit ?? 0).toFixed(2)}</b>.
              </p>
            </div>
            <div className="form-grid">
              <label>Operation</label>
              <PropSelect
                fill
                value={op}
                options={Object.entries(OPS).map(([value, o]) => ({ value, label: o.label }))}
                onChange={setOp}
              />
              <label>Amount</label>
              <input type="text" value={amount} onChange={(e) => setAmount(e.target.value)} autoFocus />
              <label>Comment</label>
              <input type="text" maxLength={64} value={comment} onChange={(e) => setComment(e.target.value)} />
            </div>
          </div>
        </SettingsDialog>
      </div>
    </div>
  );
}

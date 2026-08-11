import { useState } from "react";
import { fetchUser } from "@/api/endpoints/users.js";
import { Icon } from "@/components/ui/Icon.jsx";
import { BalanceOps } from "./BalanceOps.jsx";

/** Balance operations: pick an account by login, then work its money. */
export function BalanceModule() {
  const [typed, setTyped] = useState("");
  const [account, setAccount] = useState(null);
  const [error, setError] = useState("");

  async function pick() {
    setError("");
    const login = Number(typed);
    if (!login) return;
    const res = await fetchUser(login);
    if (!res.ok) {
      setAccount(null);
      setError("No account with that login");
      return;
    }
    setAccount(res.data);
  }

  return (
    <div className="module-root balance-module">
      <div className="sym-sessions-intro">
        <span className="sym-tab-intro-icon" aria-hidden="true">
          <Icon id="balance" size={48} />
        </span>
        <p>
          Deposits, withdrawals, credits and corrections land on the account as deals. Pick the
          account by its login; the totals below follow the market live.
        </p>
      </div>
      <div className="balance-picker">
        <label>Account:</label>
        <input
          type="text"
          value={typed}
          placeholder="login"
          onChange={(e) => setTyped(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && pick()}
        />
        <button type="button" onClick={pick}>Open</button>
        {account && <span className="balance-picker-name">{account.name} — {account.group}</span>}
        {error && <span className="login-error">{error}</span>}
      </div>
      {account && <BalanceOps login={account.login} />}
    </div>
  );
}

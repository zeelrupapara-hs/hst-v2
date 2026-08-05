import { useState } from "react";
import { Link } from "react-router-dom";
import { CenteredCard } from "../components/ui/CenteredCard.jsx";
import { getApiBase, login, setApiBase } from "../lib/api.js";

export function LoginPage() {
  const [apiBase, setApiBaseState] = useState(getApiBase());
  const [loginId, setLoginId] = useState("1000");
  const [password, setPassword] = useState("Bootstrap-Admin-2026!");
  const [connectionType, setConnectionType] = useState("33");
  const [msg, setMsg] = useState(null);

  async function handleLogin() {
    setApiBase(apiBase.trim());
    const result = await login(loginId.trim(), password, connectionType);
    if (result.ok) {
      setMsg({
        type: "ok",
        text: `Connected as login ${result.data.data.login}. Open admin or manager panel.`,
      });
    } else {
      setMsg({
        type: "err",
        text: `Failed (${result.status}): ${JSON.stringify(result.data)}`,
      });
    }
  }

  return (
    <CenteredCard
      className="login-card"
      title="Connect to HST API"
      description="Sign in once — modules load live data from your backend."
    >
      <div className="field">
        <label htmlFor="apiBase">API base URL</label>
        <input
          id="apiBase"
          type="url"
          value={apiBase}
          onChange={(e) => setApiBaseState(e.target.value)}
        />
      </div>
      <div className="field">
        <label htmlFor="login">Login</label>
        <input
          id="login"
          type="text"
          value={loginId}
          onChange={(e) => setLoginId(e.target.value)}
        />
      </div>
      <div className="field">
        <label htmlFor="password">Password</label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>
      <div className="field">
        <label htmlFor="connectionType">Panel</label>
        <select
          id="connectionType"
          value={connectionType}
          onChange={(e) => setConnectionType(e.target.value)}
        >
          <option value="32">Administrator (32)</option>
          <option value="33">Manager (33)</option>
        </select>
      </div>

      <div className="actions">
        <button type="button" className="primary" onClick={handleLogin}>
          Sign in
        </button>
        <Link className="btn" to="/">
          Back
        </Link>
      </div>
      {msg && <div className={`msg ${msg.type}`}>{msg.text}</div>}
    </CenteredCard>
  );
}

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { signIn } from "@/api/endpoints/auth.js";
import { useSession } from "@/hooks/useSession.js";

/** The sign-in window, drawn as a settings dialog because that is the platform's one dialog. */
export function LoginPage() {
  const navigate = useNavigate();
  const session = useSession();
  const [form, setForm] = useState({ login: "", password: "", connectionType: 32 });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const set = (key) => (e) => setForm({ ...form, [key]: e.target.value });

  async function submit(e) {
    e.preventDefault();
    setBusy(true);
    setError("");

    const res = await signIn(form.login.trim(), form.password, Number(form.connectionType));
    setBusy(false);

    if (!res.ok) {
      setError(res.message);
      return;
    }

    await session.reload();
    navigate(Number(form.connectionType) === 32 ? "/admin" : "/manager", { replace: true });
  }

  return (
    <div className="login-host">
      <form className="config-window settings-dialog login-window" onSubmit={submit}>
        <div className="config-title">Connect to Trade Server</div>
        <div className="config-body">
          <div className="config-panel active">
            <div className="login-grid">
              <label htmlFor="login">Login</label>
              <input id="login" autoFocus value={form.login} onChange={set("login")} />
              <label htmlFor="password">Password</label>
              <input id="password" type="password" value={form.password} onChange={set("password")} />
              <label htmlFor="panel">Terminal</label>
              <select id="panel" value={form.connectionType} onChange={set("connectionType")}>
                <option value={32}>Administrator</option>
                <option value={33}>Manager</option>
              </select>
            </div>
            {error && <p className="login-error">{error}</p>}
          </div>
        </div>
        <div className="config-actions">
          <button type="submit" className="config-ok" disabled={busy}>
            {busy ? "Connecting…" : "OK"}
          </button>
        </div>
      </form>
    </div>
  );
}

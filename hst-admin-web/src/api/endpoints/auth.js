import { clearSession, getApiBase, request, setSession } from "@/api/client.js";

/**
 * Staff login: HTTP Basic credentials + the panel's connection type in the body.
 * @param {string} login @param {string} password @param {32|33} connectionType
 */
export async function signIn(login, password, connectionType) {
  let res;
  try {
    res = await fetch(getApiBase() + "/auth/v1/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Basic " + btoa(`${login}:${password}`),
      },
      body: JSON.stringify({ connection_type: connectionType }),
    });
  } catch {
    return { ok: false, message: "server unreachable" };
  }

  const envelope = await res.json().catch(() => ({}));
  if (!res.ok || envelope.success === false || !envelope.data?.access_token) {
    return { ok: false, message: envelope.message || envelope.error || `login failed (${res.status})` };
  }

  setSession(envelope.data);
  return { ok: true, data: envelope.data };
}

export async function signOut() {
  await request("/api/v1/auth/logout", { method: "POST" });
  clearSession();
}

export const fetchMe = () => request("/api/v1/auth/me");

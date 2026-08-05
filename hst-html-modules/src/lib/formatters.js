export function esc(s) {
  if (s == null) return "";
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function fmtAuthMode(n) {
  const m = { 0: "Normal", 1: "RSA 1024", 2: "RSA 2048", 3: "Custom SSL" };
  return m[n] != null ? m[n] : String(n);
}

export function fmtExecMode(n) {
  const m = { 0: "Request", 1: "Instant", 2: "Market", 3: "Exchange" };
  return m[n] != null ? m[n] : String(n);
}

export function fmtClientStatus(n) {
  const m = { 0: "Registered", 1: "Active", 2: "Suspended", 3: "Closed" };
  return m[n] != null ? m[n] : String(n);
}

export function fmtKycStatus(n) {
  const m = { 0: "None", 1: "Pending", 2: "Approved", 3: "Rejected" };
  return m[n] != null ? m[n] : String(n);
}

export function fmtTs(sec) {
  if (sec == null || sec === 0) return "—";
  const d = new Date(Number(sec) * (String(sec).length > 10 ? 1 : 1000));
  if (isNaN(d.getTime())) return String(sec);
  return d.toISOString().replace("T", " ").slice(0, 19);
}

/** Unix nanoseconds (API symbol date_modified). */
export function fmtNsTs(ns) {
  if (ns == null || ns === 0) return "—";
  const ms = Number(ns) / 1_000_000;
  const d = new Date(ms);
  if (isNaN(d.getTime())) return String(ns);
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}.${pad(d.getMonth() + 1)}.${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export function buildQuery(params) {
  const parts = [];
  Object.keys(params || {}).forEach((k) => {
    if (params[k] != null && params[k] !== "") {
      parts.push(encodeURIComponent(k) + "=" + encodeURIComponent(params[k]));
    }
  });
  return parts.length ? "?" + parts.join("&") : "";
}

export function statusMessage(title, res, count) {
  const base = title + ": " + count + " row(s)";
  if (res.demo) return base + " — demo data";
  return base + " — " + res.elapsed + " ms";
}

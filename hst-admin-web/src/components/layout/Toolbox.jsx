import { Fragment, useCallback, useEffect, useState } from "react";
import { fetchJournal, requestJournal } from "@/api/endpoints/journal.js";
import { useToolbox } from "@/hooks/useToolbox.jsx";
import { formatNs } from "@/lib/time.js";
import { JournalType_name } from "@/constants/journal.js";
import { SummaryPanel } from "@/components/layout/SummaryPanel.jsx";
import { ExposureTab } from "@/modules/toolbox/ExposureTab.jsx";
import { useSession } from "@/hooks/useSession.js";

const toNs = (local) => (local ? new Date(local).getTime() * 1e6 : undefined);

const dayOf = (ns) => formatNs(ns).slice(0, 10);

const JOURNAL_MODES = [
  ["full", "Full"],
  ["without_logins", "Without logins"],
  ["errors_only", "Errors only"],
];

// Today, yesterday or a custom window, in the way the reference journal asks for a period.
function periodBounds(period, from, to) {
  const startOfDay = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime() * 1e6;
  const now = new Date();
  if (period === "today") return { from: startOfDay(now) };
  if (period === "yesterday") {
    const y = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1);
    return { from: startOfDay(y), to: startOfDay(now) - 1 };
  }
  return { from: toNs(from), to: toNs(to) };
}

const rowText = (r) =>
  [formatNs(r.created_at), r.login ? r.login : "server", r.ip, r.message].join("\t");

function exportCsv(rows) {
  const esc = (v) => `"${String(v ?? "").replace(/"/g, '""')}"`;
  const lines = rows.map((r) =>
    [formatNs(r.created_at), r.login || "server", JournalType_name[r.type] ?? r.type, r.channel, r.ip, r.message].map(esc).join(","),
  );
  const blob = new Blob(["Time,Source,Type,Channel,IP,Message\n" + lines.join("\n")], { type: "text/csv" });
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = "journal.csv";
  a.click();
  URL.revokeObjectURL(a.href);
}

function JournalRows({ rows, message, allColumns, selected, onSelect }) {
  const cols = allColumns ? 6 : 4;
  let day = "";
  return (
    <table className="journal-table">
      <thead>
        <tr>
          <th>Time</th>
          <th>Source</th>
          {allColumns && <th>Type</th>}
          {allColumns && <th>Channel</th>}
          <th>IP</th>
          <th>Message</th>
        </tr>
      </thead>
      <tbody>
        {(rows ?? []).map((r) => {
          const d = dayOf(r.created_at);
          const separator = d !== day;
          day = d;
          return (
            <Fragment key={r.journal_id}>
              {separator && (
                <tr className="journal-day">
                  <td colSpan={cols}>{d}</td>
                </tr>
              )}
              <tr
                className={selected === r.journal_id ? "selected" : undefined}
                onClick={() => onSelect(r.journal_id)}
              >
                <td>{formatNs(r.created_at)}</td>
                <td>{r.login ? r.login : "server"}</td>
                {allColumns && <td>{JournalType_name[r.type] ?? r.type}</td>}
                {allColumns && <td>{r.channel}</td>}
                <td>{r.ip}</td>
                <td>{r.message}</td>
              </tr>
            </Fragment>
          );
        })}
        {rows !== null && rows.length === 0 && (
          <tr>
            <td colSpan={cols}>{message || "No journal entries."}</td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

/** The server journal: a keyword, a period, a mode and a module type, then Request. */
function JournalPanel({ query, tick }) {
  const [form, setForm] = useState({ search: "", period: "today", from: "", to: "", mode: "full", type: 0 });
  const [rows, setRows] = useState(null);
  const [message, setMessage] = useState("");
  const [find, setFind] = useState("");
  const [allColumns, setAllColumns] = useState(false);
  const [selected, setSelected] = useState(null);
  const [busy, setBusy] = useState(false);

  const run = useCallback(async (f) => {
    setBusy(true);
    const res = await requestJournal({
      search: f.search.trim(),
      ...periodBounds(f.period, f.from, f.to),
      mode: f.mode,
      type: f.type,
      scope: "server",
    });
    setBusy(false);
    if (res.ok) {
      setRows(res.data);
      setMessage(res.truncated ? "The first 5000 entries are shown, narrow the request." : "");
    } else {
      setRows([]);
      setMessage(res.message || "journal request failed");
    }
  }, []);

  // a module's Journal command lands here with its name as the keyword and requests at once
  useEffect(() => {
    const q = query.trim();
    if (!q) {
      fetchJournal({ limit: 100 }).then((res) => {
        if (res.ok) setRows((res.data ?? []).slice().reverse());
        else setMessage(res.message || "journal unavailable");
      });
      return;
    }
    const next = { search: q, period: "today", from: "", to: "", mode: "full", type: 0 };
    setForm(next);
    run(next);
  }, [query, tick, run]);

  const shown = find ? (rows ?? []).filter((r) => r.message.toLowerCase().includes(find.toLowerCase())) : rows;
  const set = (k) => (e) => setForm({ ...form, [k]: e.target.value });

  function copySelected() {
    const r = (rows ?? []).find((x) => x.journal_id === selected);
    if (r) navigator.clipboard?.writeText(rowText(r));
  }

  return (
    <>
      <div className="toolbox-search-bar">
        <input
          type="text"
          placeholder="Keyword, e.g. EURUSD or a login…"
          value={form.search}
          onChange={set("search")}
          onKeyDown={(e) => e.key === "Enter" && run(form)}
        />
        <select value={form.mode} onChange={set("mode")}>
          {JOURNAL_MODES.map(([v, l]) => <option key={v} value={v}>{l}</option>)}
        </select>
        <select value={form.type} onChange={(e) => setForm({ ...form, type: Number(e.target.value) })}>
          <option value={0}>All</option>
          {Object.entries(JournalType_name).filter(([k]) => k !== "0").map(([k, l]) => (
            <option key={k} value={k}>{l}</option>
          ))}
        </select>
        <select value={form.period} onChange={set("period")}>
          <option value="today">Today</option>
          <option value="yesterday">Yesterday</option>
          <option value="custom">Custom</option>
        </select>
        {form.period === "custom" && (
          <>
            <input type="datetime-local" value={form.from} onChange={set("from")} />
            <input type="datetime-local" value={form.to} onChange={set("to")} />
          </>
        )}
        <button type="button" onClick={() => run(form)} disabled={busy}>{busy ? "Requesting…" : "Request"}</button>
        {rows !== null && <span className="module-note">{shown.length} entries</span>}
      </div>
      <div className="toolbox-search-bar">
        <input type="text" placeholder="Find in results…" value={find} onChange={(e) => setFind(e.target.value)} />
        <label>
          <input type="checkbox" checked={allColumns} onChange={(e) => setAllColumns(e.target.checked)} /> Columns
        </label>
        <button type="button" onClick={copySelected} disabled={selected === null}>Copy</button>
        <button type="button" onClick={() => exportCsv(shown ?? [])} disabled={!shown?.length}>Export</button>
        {message && <span className="module-note">{message}</span>}
      </div>
      <JournalRows rows={shown} message={message} allColumns={allColumns} selected={selected} onSelect={setSelected} />
    </>
  );
}

/** Bottom service panel: the server journal, and the desk views on the manager side. */
export function Toolbox() {
  const session = useSession();
  const toolbox = useToolbox();
  const tab = toolbox?.tab ?? "journal";
  const setTab = toolbox?.setTab ?? (() => {});
  const journalQuery = toolbox?.journalQuery ?? "";
  const journalTick = toolbox?.tick ?? 0;

  const isManagerPanel = session.terminal !== "administrator";

  return (
    <div className="toolbox">
      <div className="toolbox-tabs">
        {[
          { id: "journal", label: "Journal" },
          ...(isManagerPanel ? [{ id: "summary", label: "Summary" }] : []),
          ...(isManagerPanel && session.can?.right_risk_manager !== false
            ? [{ id: "exposure", label: "Exposure" }]
            : []),
        ].map((t) => (
          <button
            key={t.id}
            type="button"
            className={tab === t.id ? "active" : undefined}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "journal" && (
        <div className="toolbox-panel active">
          <JournalPanel query={journalQuery} tick={journalTick} />
        </div>
      )}

      {tab === "summary" && (
        <div className="toolbox-panel active">
          <SummaryPanel />
        </div>
      )}

      {tab === "exposure" && (
        <div className="toolbox-panel active">
          <ExposureTab />
        </div>
      )}

    </div>
  );
}

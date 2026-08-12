import { useEffect, useState } from "react";
import { fetchJournal, searchJournal } from "@/api/endpoints/journal.js";
import { useToolbox } from "@/hooks/useToolbox.jsx";
import { formatNs } from "@/lib/time.js";
import { JournalType_name } from "@/constants/journal.js";
import { SummaryPanel } from "@/components/layout/SummaryPanel.jsx";
import { ExposureTab } from "@/modules/toolbox/ExposureTab.jsx";
import { useSession } from "@/hooks/useSession.js";

const toNs = (local) => (local ? new Date(local).getTime() * 1e6 : undefined);

function JournalRows({ rows, message }) {
  return (
    <table className="journal-table">
      <thead>
        <tr>
          <th>Time</th>
          <th>Source</th>
          <th>Type</th>
          <th>Channel</th>
          <th>IP</th>
          <th>Message</th>
        </tr>
      </thead>
      <tbody>
        {(rows ?? []).map((r) => (
          <tr key={r.journal_id}>
            <td>{formatNs(r.created_at)}</td>
            <td>{r.login ? r.login : "server"}</td>
            <td>{JournalType_name[r.type] ?? r.type}</td>
            <td>{r.channel}</td>
            <td>{r.ip}</td>
            <td>{r.message}</td>
          </tr>
        ))}
        {rows !== null && rows.length === 0 && (
          <tr>
            <td colSpan={6}>{message || "No journal entries."}</td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

function SearchPanel() {
  const [query, setQuery] = useState({ search: "", from: "", to: "", errorsOnly: false });
  const [rows, setRows] = useState(null);
  const [message, setMessage] = useState("");

  async function run() {
    const res = await searchJournal({
      search: query.search.trim(),
      from: toNs(query.from),
      to: toNs(query.to),
      errorsOnly: query.errorsOnly,
    });
    if (res.ok) {
      setRows(res.data ?? []);
      setMessage("");
    } else {
      setRows([]);
      setMessage(res.message || "search failed");
    }
  }

  return (
    <>
      <div className="toolbox-search-bar">
        <input
          type="text"
          placeholder="Message text…"
          value={query.search}
          onChange={(e) => setQuery({ ...query, search: e.target.value })}
          onKeyDown={(e) => e.key === "Enter" && run()}
        />
        <input
          type="datetime-local"
          value={query.from}
          onChange={(e) => setQuery({ ...query, from: e.target.value })}
        />
        <input
          type="datetime-local"
          value={query.to}
          onChange={(e) => setQuery({ ...query, to: e.target.value })}
        />
        <label>
          <input
            type="checkbox"
            checked={query.errorsOnly}
            onChange={(e) => setQuery({ ...query, errorsOnly: e.target.checked })}
          />{" "}
          Errors only
        </label>
        <button type="button" onClick={run}>Request</button>
        {rows !== null && <span className="module-note">{rows.length} entries</span>}
      </div>
      <JournalRows rows={rows} message={message || "Enter a query and press Request."} />
    </>
  );
}

/** Bottom service panel: Journal live from the server, Search lands with the modules. */
export function Toolbox() {
  const session = useSession();
  const toolbox = useToolbox();
  const tab = toolbox?.tab ?? "journal";
  const setTab = toolbox?.setTab ?? (() => {});
  const journalQuery = toolbox?.journalQuery ?? "";
  const journalTick = toolbox?.tick ?? 0;

  const [rows, setRows] = useState(null);
  const [message, setMessage] = useState("");
  const [journalSearch, setJournalSearch] = useState("");

  useEffect(() => {
    if (tab !== "journal") return;

    const q = journalQuery.trim();
    setJournalSearch(q);

    if (q) {
      searchJournal({ search: q, limit: 200 }).then((res) => {
        if (res.ok) {
          setRows(res.data ?? []);
          setMessage("");
        } else {
          setRows([]);
          setMessage(res.message || "journal search failed");
        }
      });
      return;
    }

    fetchJournal({ limit: 100 }).then((res) => {
      if (res.ok) setRows(res.data ?? []);
      else setMessage(res.message || "journal unavailable");
    });
  }, [tab, journalQuery, journalTick]);

  const isManagerPanel = session.terminal !== "administrator";

  return (
    <div className="toolbox">
      <div className="toolbox-tabs">
        {[
          { id: "journal", label: "Journal" },
          { id: "search", label: "Search" },
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
          {journalSearch && (
            <div className="toolbox-search-bar">
              <input type="text" value={journalSearch} readOnly />
              <span className="module-note">Filtered by configuration</span>
            </div>
          )}
          <JournalRows rows={rows} message={message} />
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

      {tab === "search" && (
        <div className="toolbox-panel active">
          <SearchPanel />
        </div>
      )}
    </div>
  );
}

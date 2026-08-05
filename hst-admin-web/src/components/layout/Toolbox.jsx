import { useEffect, useState } from "react";
import { fetchJournal } from "@/api/endpoints/journal.js";
import { formatNs } from "@/lib/time.js";

/** Bottom service panel: Journal live from the server, Search lands with the modules. */
export function Toolbox() {
  const [tab, setTab] = useState("journal");
  const [rows, setRows] = useState(null);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (tab !== "journal" || rows !== null) return;

    fetchJournal({ limit: 100 }).then((res) => {
      if (res.ok) setRows(res.data ?? []);
      else setMessage(res.message || "journal unavailable");
    });
  }, [tab, rows]);

  return (
    <div className="toolbox">
      <div className="toolbox-tabs">
        {[
          { id: "journal", label: "Journal" },
          { id: "search", label: "Search" },
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
          <table className="journal-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Channel</th>
                <th>IP</th>
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              {(rows ?? []).map((r) => (
                <tr key={r.journal_id}>
                  <td>{formatNs(r.created_at)}</td>
                  <td>{r.channel}</td>
                  <td>{r.ip}</td>
                  <td>{r.message}</td>
                </tr>
              ))}
              {rows !== null && rows.length === 0 && (
                <tr>
                  <td colSpan={4}>{message || "No journal entries."}</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {tab === "search" && (
        <div className="toolbox-panel active">
          <p style={{ padding: 8, color: "#666" }}>
            Search covers loaded module lists; it lands with the modules.
          </p>
        </div>
      )}
    </div>
  );
}

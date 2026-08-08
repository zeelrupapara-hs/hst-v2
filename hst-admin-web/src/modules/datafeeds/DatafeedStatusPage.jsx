import { useCallback, useEffect, useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import { updateDatafeed } from "@/api/endpoints/datafeeds.js";
import { searchJournal } from "@/api/endpoints/journal.js";
import { formatNs } from "@/lib/time.js";

const sourceText = (mode) => {
  const m = mode ?? 1;
  if (m & 1 && m & 2) return "Quotes and News";
  return m & 2 ? "News" : "Quotes";
};

// bytes as the reference prints traffic: a plain count under a KB, then KB/MB
const traffic = (bytes) => {
  const n = bytes ?? 0;
  if (n < 1024) return String(n);
  if (n < 1024 * 1024) return `${Math.round(n / 1024)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
};

// journal codes: 0 info, 1 warning, 2 error, 3 critical
const codeGlyph = (code) => (code >= 2 ? "✕" : code === 1 ? "!" : "•");

function FeedJournal({ feedId }) {
  const [rows, setRows] = useState(null);
  const [search, setSearch] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");

  const requestRows = useCallback(() => {
    const dayNs = (d, end) =>
      d ? new Date(`${d}T${end ? "23:59:59" : "00:00:00"}`).getTime() * 1e6 : undefined;
    searchJournal({
      channel: `datafeed:${feedId}`,
      search,
      from: dayNs(from),
      to: dayNs(to, true),
    }).then((res) => setRows(res.ok ? res.data || [] : []));
  }, [feedId, search, from, to]);

  useEffect(() => {
    requestRows();
  }, [requestRows]);

  return (
    <div className="df-status-journal">
      <div className="table-wrap">
        <table className="data-table data-table-grid data-table-auto">
          <thead>
            <tr>
              <th>Time</th>
              <th>Message</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((r) => (
              <tr key={r.journal_id}>
                <td className="df-journal-time">
                  <span className={`df-journal-code df-journal-code-${r.code >= 2 ? "err" : r.code === 1 ? "warn" : "info"}`}>
                    {codeGlyph(r.code)}
                  </span>
                  {formatNs(r.created_at).replace(/-/g, ".")}
                </td>
                <td>{r.message}</td>
              </tr>
            ))}
            {rows?.length === 0 && (
              <tr>
                <td colSpan={2} className="df-empty">No journal entries</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="df-journal-bar">
        <input
          type="text"
          placeholder="Search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && requestRows()}
        />
        <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
        <input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
        <button type="button" onClick={requestRows}>Request</button>
      </div>
    </div>
  );
}

/** One feed as MT-style Status page: enabled banner, configuration, live counters, journal. */
export function DatafeedStatusPage({ feed, canEdit, onSaved }) {
  const [tab, setTab] = useState("status");
  const enabled = feed.enable === 1;

  async function toggle() {
    const res = await updateDatafeed(feed.datafeed_id, { enable: enabled ? 0 : 1 });
    if (!res.ok) window.alert(res.message || "save failed");
    onSaved();
  }

  return (
    <div className="df-status-page">
      <div className="df-status-head">
        <div>
          <h2 className="df-section-title">
            <Icon id="datafeeds" size={18} /> {feed.name}
          </h2>
          <p className="df-section-sub">{feed.module}</p>
        </div>
        <div className="df-status-enable">
          <span>The feeder is currently {enabled ? "enabled" : "disabled"}</span>
          <button
            type="button"
            className={enabled ? "df-btn-disable" : "df-btn-enable"}
            disabled={!canEdit}
            onClick={toggle}
          >
            {enabled ? "Disable" : "Enable"}
          </button>
        </div>
      </div>

      {tab === "status" ? (
        <div className="df-status-body">
          <div className="df-status-block">
            <h3>Configuration</h3>
            <dl>
              <dt>Module:</dt>
              <dd>{feed.module}</dd>
              <dt>Source:</dt>
              <dd>{sourceText(feed.mode)}</dd>
              <dt>Feed server:</dt>
              <dd>{feed.feed_server || "—"}</dd>
              <dt>Gateway server:</dt>
              <dd>{feed.gateway_server || "—"}</dd>
            </dl>
          </div>
          <div className="df-status-block">
            <h3>Databases</h3>
            <dl>
              <dt>Price ticks:</dt>
              <dd>{feed.ticks_count ?? 0}</dd>
              <dt>Price books:</dt>
              <dd>{feed.books_count ?? 0}</dd>
              <dt>News:</dt>
              <dd>{feed.news_count ?? 0}</dd>
              <dt>Traffic:</dt>
              <dd>{traffic(feed.bytes_received)}</dd>
            </dl>
          </div>
        </div>
      ) : (
        <FeedJournal feedId={feed.datafeed_id} />
      )}

      <div className="df-pane-tabs">
        <button type="button" className={tab === "status" ? "active" : ""} onClick={() => setTab("status")}>
          Status
        </button>
        <button type="button" className={tab === "journal" ? "active" : ""} onClick={() => setTab("journal")}>
          Journal
        </button>
      </div>
    </div>
  );
}

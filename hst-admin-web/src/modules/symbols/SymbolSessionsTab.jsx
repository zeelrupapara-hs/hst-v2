import { useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import {
  DAY_NAMES,
  SESSION_QUOTE,
  SESSION_TRADE,
  formatDaySessions,
  mergeDaySessions,
} from "@/lib/symbolSessions.js";
import { formatUnixSec, parseMt5DateTimeToSec } from "@/lib/time.js";
import { SessionEditorDialog } from "./SessionEditorDialog.jsx";

const pad = (n) => String(n).padStart(2, "0");

const EPOCH_STAMP = "1970.01.01 00:00";

const formatStamp = (ns) => {
  if (!ns) return EPOCH_STAMP;
  const d = fromNs(ns);
  return `${d.getFullYear()}.${pad(d.getMonth() + 1)}.${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const parseStamp = (raw) => {
  const s = raw.trim();
  const m = s.match(/^(\d{4})\.(\d{2})\.(\d{2})\s+(\d{2}):(\d{2})$/);
  if (!m) return null;
  const d = new Date(+m[1], +m[2] - 1, +m[3], +m[4], +m[5]);
  if (Number.isNaN(d.getTime())) return null;
  return d.getTime() * 1e6;
};

/** Sessions grid: select days (Ctrl/Shift), Edit or double-click opens the timeline editor. */
export function SymbolSessionsTab({ s, set }) {
  const sessions = s.sessions || [];
  const [selectedDays, setSelectedDays] = useState([]);
  const [editorOpen, setEditorOpen] = useState(false);
  const [useLimits, setUseLimits] = useState(!!(s.time_start || s.time_expiration));

  function toggleDay(day, e) {
    if (e.ctrlKey || e.metaKey) {
      setSelectedDays((prev) =>
        prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day],
      );
    } else if (e.shiftKey && selectedDays.length) {
      const last = selectedDays[selectedDays.length - 1];
      const [lo, hi] = [Math.min(last, day), Math.max(last, day)];
      setSelectedDays(Array.from({ length: hi - lo + 1 }, (_, i) => lo + i));
    } else {
      setSelectedDays([day]);
    }
  }

  function saveDays(dayIndexes, quoteWins, tradeWins, separate) {
    set("sessions", mergeDaySessions(sessions, dayIndexes, quoteWins, tradeWins, separate));
  }

  return (
    <>
      <div className="sym-sessions-intro">
        <span className="sym-tab-intro-icon" aria-hidden="true">
          <Icon id="symbols" size={48} />
        </span>
        <p>
          The setting up of trade and quotation sessions of the symbol by days. During the
          quotation session it is possible to view the price dynamics but trading is prohibited.
        </p>
      </div>
      <div className="sym-sessions-main">
        <div className="sym-sessions-table-wrap">
          <table className="sym-sessions-table">
            <thead>
              <tr>
                <th>Day</th>
                <th>Quotes</th>
                <th>Trade</th>
              </tr>
            </thead>
            <tbody>
              {DAY_NAMES.map((name, day) => (
                <tr
                  key={name}
                  className={selectedDays.includes(day) ? "selected" : ""}
                  onClick={(e) => toggleDay(day, e)}
                  onDoubleClick={() => {
                    setSelectedDays([day]);
                    setEditorOpen(true);
                  }}
                >
                  <td>
                    <span className="sym-day-icon" aria-hidden="true" />
                    {name}
                  </td>
                  <td>{formatDaySessions(sessions, SESSION_QUOTE, day) || " "}</td>
                  <td>{formatDaySessions(sessions, SESSION_TRADE, day) || " "}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <button
          type="button"
          className="sym-sessions-edit"
          disabled={!selectedDays.length}
          onClick={() => setEditorOpen(true)}
        >
          Edit
        </button>
      </div>
      <div className="sym-sessions-limits">
        <label className="sym-check sym-sessions-limits-check">
          <input
            type="checkbox"
            checked={useLimits}
            onChange={(e) => {
              setUseLimits(e.target.checked);
              if (!e.target.checked) {
                set("time_start", 0);
                set("time_expiration", 0);
              }
            }}
          />
          <span>Use time limits</span>
        </label>
        <div className="sym-sessions-limits-fields">
          <label className="sym-sessions-limits-label">From:</label>
          <input
            type="text"
            className="sym-sessions-datetime"
            disabled={!useLimits}
            defaultValue={formatStamp(s.time_start)}
            key={`from-${s.time_start}-${useLimits}`}
            placeholder={EPOCH_STAMP}
            onBlur={(e) => {
              const ns = parseStamp(e.target.value);
              if (ns != null) set("time_start", ns);
              else e.target.value = formatStamp(s.time_start);
            }}
          />
          <label className="sym-sessions-limits-label">To:</label>
          <input
            type="text"
            className="sym-sessions-datetime"
            disabled={!useLimits}
            defaultValue={formatStamp(s.time_expiration)}
            key={`to-${s.time_expiration}-${useLimits}`}
            placeholder={EPOCH_STAMP}
            onBlur={(e) => {
              const ns = parseStamp(e.target.value);
              if (ns != null) set("time_expiration", ns);
              else e.target.value = formatStamp(s.time_expiration);
            }}
          />
        </div>
      </div>
      {editorOpen && (
        <SessionEditorDialog
          symbol={s.symbol}
          dayIndexes={selectedDays}
          sessions={sessions}
          onSave={saveDays}
          onClose={() => setEditorOpen(false)}
        />
      )}
    </>
  );
}

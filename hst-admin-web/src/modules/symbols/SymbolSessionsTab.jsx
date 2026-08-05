import { useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import {
  DAY_NAMES,
  SESSION_QUOTE,
  SESSION_TRADE,
  formatDaySessions,
  mergeDaySessions,
} from "@/lib/symbolSessions.js";
import { fromNs } from "@/lib/time.js";
import { SessionEditorDialog } from "./SessionEditorDialog.jsx";

const pad = (n) => String(n).padStart(2, "0");

const toLocal = (ns) => {
  if (!ns) return "";
  const d = fromNs(ns);
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const toNs = (local) => (local ? new Date(local).getTime() * 1e6 : 0);

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
          <Icon id="symbols-tree" size={48} />
        </span>
        <p>
          The setting up of trade and quotation sessions of the symbol by days. During the
          quotation session it is possible to view the price dynamics but trading is prohibited.
        </p>
      </div>
      <div className="sym-sessions-table-wrap sym-sessions-row">
        <div className="sym-sessions-gutter">
          <button
            type="button"
            className="sym-sessions-edit"
            disabled={!selectedDays.length}
            onClick={() => setEditorOpen(true)}
          >
            Edit
          </button>
        </div>
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
      <div className="sym-sessions-limits">
        <label className="sym-check">
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
        <label>From</label>
        <input
          type="datetime-local"
          disabled={!useLimits}
          value={toLocal(s.time_start)}
          onChange={(e) => set("time_start", toNs(e.target.value))}
        />
        <label>To</label>
        <input
          type="datetime-local"
          disabled={!useLimits}
          value={toLocal(s.time_expiration)}
          onChange={(e) => set("time_expiration", toNs(e.target.value))}
        />
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

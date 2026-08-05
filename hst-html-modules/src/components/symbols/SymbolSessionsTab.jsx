import { useState } from "react";
import { createPortal } from "react-dom";
import { Icon } from "../ui/Icon.jsx";
import {
  DAY_NAMES,
  SESSION_QUOTE,
  SESSION_TRADE,
  fmtMt5DateTime,
  formatDaySessions,
  parseMt5DateTime,
} from "../../lib/symbolSessions.js";
import { SessionEditorDialog } from "./SessionEditorDialog.jsx";

function emptyLimitValue(enabled, sec) {
  if (!enabled || sec == null || sec === 0) return "";
  return fmtMt5DateTime(sec);
}

export function SymbolSessionsTab({ symbol, onUpdateSessions, onUpdateTimeLimits }) {
  const sessions = symbol?.sessions || [];
  const [selectedDays, setSelectedDays] = useState([]);
  const [editorOpen, setEditorOpen] = useState(false);
  const useLimits = !!(symbol?.time_start || symbol?.time_expiration);
  const [limitsEnabled, setLimitsEnabled] = useState(useLimits);
  const [fromStr, setFromStr] = useState(() =>
    emptyLimitValue(useLimits, symbol?.time_start)
  );
  const [toStr, setToStr] = useState(() =>
    emptyLimitValue(useLimits, symbol?.time_expiration)
  );

  function toggleDay(day, e) {
    if (e.ctrlKey || e.metaKey) {
      setSelectedDays((prev) =>
        prev.includes(day) ? prev.filter((d) => d !== day) : [...prev, day]
      );
    } else if (e.shiftKey && selectedDays.length) {
      const last = selectedDays[selectedDays.length - 1];
      const lo = Math.min(last, day);
      const hi = Math.max(last, day);
      const range = [];
      for (let d = lo; d <= hi; d++) range.push(d);
      setSelectedDays(range);
    } else {
      setSelectedDays([day]);
    }
  }

  function openEditor() {
    if (!selectedDays.length) return;
    setEditorOpen(true);
  }

  async function handleSaveDays(dayIndexes, quoteWins, tradeWins, separate) {
    await onUpdateSessions(dayIndexes, quoteWins, tradeWins, separate);
    setEditorOpen(false);
  }

  async function applyLimits(enabled, from, to) {
    setLimitsEnabled(enabled);
    await onUpdateTimeLimits(enabled, parseMt5DateTime(from), parseMt5DateTime(to));
  }

  return (
    <>
      <div className="sym-sessions-intro">
        <span className="sym-tab-intro-icon" aria-hidden="true">
          <Icon name="time" size={48} />
        </span>
        <p>
          The setting up of trade and quotation sessions of the symbol by days.
          During the quotation session it is possible to view the price dynamics
          but trading is prohibited.
        </p>
      </div>
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
                <td>
                  {formatDaySessions(sessions, SESSION_QUOTE, day) || "\u00a0"}
                </td>
                <td>
                  {formatDaySessions(sessions, SESSION_TRADE, day) || "\u00a0"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="sym-sessions-toolbar">
        <button
          type="button"
          className="sym-sessions-edit"
          disabled={!selectedDays.length}
          onClick={openEditor}
        >
          Edit
        </button>
      </div>
      <div className="sym-sessions-limits">
        <label>
          <input
            type="checkbox"
            checked={limitsEnabled}
            onChange={(e) => {
              const en = e.target.checked;
              let from = fromStr;
              let to = toStr;
              if (en && !from) {
                from = fmtMt5DateTime(Math.floor(Date.now() / 1000));
                setFromStr(from);
              }
              if (en && !to) {
                to = fmtMt5DateTime(Math.floor(Date.now() / 1000) + 86400 * 365);
                setToStr(to);
              }
              applyLimits(en, from, to);
            }}
          />{" "}
          Use time limits
        </label>
        <label>From:</label>
        <input
          type="text"
          value={fromStr}
          disabled={!limitsEnabled}
          placeholder="YYYY.MM.DD HH:MM"
          onChange={(e) => setFromStr(e.target.value)}
          onBlur={() =>
            limitsEnabled && applyLimits(true, fromStr || "1970.01.01 00:00", toStr)
          }
        />
        <label>To:</label>
        <input
          type="text"
          value={toStr}
          disabled={!limitsEnabled}
          placeholder="YYYY.MM.DD HH:MM"
          onChange={(e) => setToStr(e.target.value)}
          onBlur={() =>
            limitsEnabled && applyLimits(true, fromStr, toStr || "1970.01.01 00:00")
          }
        />
      </div>
      {editorOpen &&
        createPortal(
          <SessionEditorDialog
            symbol={symbol.symbol}
            dayIndexes={selectedDays}
            sessions={sessions}
            onSave={handleSaveDays}
            onClose={() => setEditorOpen(false)}
          />,
          document.body
        )}
    </>
  );
}

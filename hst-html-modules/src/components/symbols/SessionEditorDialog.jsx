import { useCallback, useEffect, useRef, useState } from "react";
import { Icon } from "../ui/Icon.jsx";
import {
  dayLabel,
  hasSeparateTrade,
  minutesToTime,
  SESSION_QUOTE,
  SESSION_TRADE,
  tradeWithinQuote,
  windowsForDay,
} from "../../lib/symbolSessions.js";

const DEFAULT_DAY_WINDOW = [{ open: 0, close: 1440 }];

function cloneWindows(sessions, type, day) {
  const wins = windowsForDay(sessions, type, day);
  if (!wins.length) return DEFAULT_DAY_WINDOW.map((w) => ({ ...w }));
  return wins.map((w) => ({ open: w.open, close: w.close }));
}

function TimelineRow({ label, windows, onChange, readOnly, variant = "quote" }) {
  const dragRef = useRef({ active: false, index: 0, edge: "open" });

  function pct(m) {
    return `${(m / 1440) * 100}%`;
  }

  function handleMove(clientX, trackEl) {
    const rect = trackEl.getBoundingClientRect();
    let m = Math.round(((clientX - rect.left) / rect.width) * 1440);
    m = Math.max(0, Math.min(1440, m));
    if (dragRef.current.shift) m = Math.round(m / 5) * 5;
    const next = windows.map((w, i) => {
      if (i !== dragRef.current.index) return { ...w };
      const copy = { ...w };
      if (dragRef.current.edge === "open") {
        copy.open = Math.min(m, copy.close - 1);
      } else {
        copy.close = Math.max(m, copy.open + 1);
      }
      return copy;
    });
    onChange(next);
  }

  useEffect(() => {
    function onUp() {
      dragRef.current.active = false;
    }
    function onMove(e) {
      if (!dragRef.current.active || !dragRef.current.track) return;
      handleMove(e.clientX, dragRef.current.track);
    }
    window.addEventListener("mouseup", onUp);
    window.addEventListener("mousemove", onMove);
    return () => {
      window.removeEventListener("mouseup", onUp);
      window.removeEventListener("mousemove", onMove);
    };
  });

  const wins = windows.length ? windows : DEFAULT_DAY_WINDOW;

  return (
    <div className="sym-timeline-row">
      <span className="sym-timeline-label">{label}</span>
      <div className="sym-timeline-track-wrap">
        <div className="sym-timeline-scale">
          {[0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24].map((h) => (
            <span
              key={h}
              className={`sym-timeline-tick${h === 24 ? " end" : ""}`}
              style={{ left: `${(h / 24) * 100}%` }}
            >
              {h < 10 ? `0${h}` : h}
            </span>
          ))}
        </div>
        <div
          className={`sym-timeline-track sym-timeline-track-${variant}${readOnly ? " readonly" : ""}`}
          ref={(el) => {
            if (dragRef.current.active) dragRef.current.track = el;
          }}
        >
          {wins.map((w, i) => (
            <div key={i}>
              <div
                className="sym-timeline-seg"
                style={{ left: pct(w.open), width: pct(w.close - w.open) }}
              />
              {!readOnly && (
                <>
                  <div
                    className="sym-timeline-marker sym-timeline-marker-draggable"
                    style={{ left: pct(w.open) }}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      dragRef.current = {
                        active: true,
                        index: i,
                        edge: "open",
                        shift: e.shiftKey,
                        track: e.currentTarget.closest(".sym-timeline-track"),
                      };
                    }}
                  >
                    <span className="sym-timeline-marker-label">
                      {minutesToTime(w.open)}
                    </span>
                    <span className="sym-timeline-marker-pin" />
                  </div>
                  <div
                    className="sym-timeline-marker sym-timeline-marker-draggable"
                    style={{ left: pct(w.close) }}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      dragRef.current = {
                        active: true,
                        index: i,
                        edge: "close",
                        shift: e.shiftKey,
                        track: e.currentTarget.closest(".sym-timeline-track"),
                      };
                    }}
                  >
                    <span className="sym-timeline-marker-label">
                      {minutesToTime(w.close)}
                    </span>
                    <span className="sym-timeline-marker-pin" />
                  </div>
                </>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export function SessionEditorDialog({
  symbol,
  dayIndexes,
  sessions,
  onSave,
  onClose,
}) {
  const primaryDay = dayIndexes[0];
  const [separate, setSeparate] = useState(() =>
    hasSeparateTrade(sessions, primaryDay)
  );
  const [quoteWins, setQuoteWins] = useState(() =>
    cloneWindows(sessions, SESSION_QUOTE, primaryDay)
  );
  const [tradeWins, setTradeWins] = useState(() => {
    const t = windowsForDay(sessions, SESSION_TRADE, primaryDay);
    if (t.length) return t.map((w) => ({ open: w.open, close: w.close }));
    return cloneWindows(sessions, SESSION_QUOTE, primaryDay);
  });
  const [error, setError] = useState("");

  const titleDays = dayLabel(dayIndexes);

  function handleOk() {
    const trade = separate ? tradeWins : quoteWins;
    if (!tradeWithinQuote(quoteWins, trade)) {
      setError("Trade sessions must fall within quotation sessions.");
      return;
    }
    onSave(dayIndexes, quoteWins, trade, separate);
    onClose();
  }

  return (
    <div className="sym-session-dialog-overlay" onClick={onClose}>
      <div
        className="sym-session-dialog"
        role="dialog"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="sym-session-dialog-title">
          Sessions {symbol}: {titleDays}
        </div>
        <div className="sym-session-dialog-body">
          <div className="sym-sessions-intro sym-sessions-intro-compact">
            <span className="sym-tab-intro-icon sym-tab-intro-icon-sm" aria-hidden="true">
              <Icon name="time" size={40} />
            </span>
            <p>
              The setting up of sessions within the selected day. Trading
              sessions must be within the quotation ones.
            </p>
          </div>
          <TimelineRow
            label="Quotes:"
            windows={quoteWins}
            onChange={setQuoteWins}
            variant="quote"
          />
          <TimelineRow
            label="Trade:"
            windows={separate ? tradeWins : quoteWins}
            onChange={setTradeWins}
            readOnly={!separate}
            variant="trade"
          />
          {error && <p className="sym-session-error">{error}</p>}
        </div>
        <div className="sym-session-dialog-footer">
          <label className="sym-session-separate">
            <input
              type="checkbox"
              checked={separate}
              onChange={(e) => setSeparate(e.target.checked)}
            />{" "}
            Enable separate trading sessions
          </label>
          <div className="sym-session-dialog-actions">
            <button type="button" className="sym-session-ok" onClick={handleOk}>
              OK
            </button>
            <button type="button" className="sym-session-cancel" onClick={onClose}>
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

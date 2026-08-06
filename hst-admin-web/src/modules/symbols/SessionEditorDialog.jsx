import { useEffect, useRef, useState } from "react";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import {
  EXTEND_SCALE_HOURS,
  PRIMARY_SCALE_HOURS,
  SESSION_QUOTE,
  SESSION_TRADE,
  TAIL_SCALE_HOURS,
  canAddWindow,
  clampMarkerMinute,
  classifyWindows,
  dayLabel,
  displayPct,
  extendScaleLeft,
  hasSeparateTrade,
  hourTickClass,
  minuteFromTrack,
  minutesToTime,
  primaryScaleLeft,
  sessionOverlaps,
  tailScaleLeft,
  toDisplayMinute,
  tradeWithinQuoteExtended,
  windowsForDay,
} from "@/lib/symbolSessions.js";

const DEFAULT_DAY_WINDOW = [{ open: 0, close: 1440 }];

function cloneWindows(sessions, type, day) {
  const wins = windowsForDay(sessions, type, day);
  return (wins.length ? wins : DEFAULT_DAY_WINDOW).map((w) => ({ ...w }));
}

function collectMarkers(windows) {
  const classified = classifyWindows(windows.length ? windows : DEFAULT_DAY_WINDOW);
  const map = new Map();
  for (const w of classified) {
    const dOpen = toDisplayMinute(w.open, w.zone);
    const dClose = toDisplayMinute(w.close, w.zone);
    map.set(`${dOpen}-open`, {
      storageMin: w.open,
      displayMin: dOpen,
      end: false,
      zone: w.zone,
      edge: "open",
    });
    map.set(`${dClose}-close`, {
      storageMin: w.close,
      displayMin: dClose,
      end: w.close >= 1440 && w.zone === "primary",
      zone: w.zone,
      edge: "close",
    });
  }
  const markers = [...map.values()].sort((a, b) => a.displayMin - b.displayMin);
  let lastPct = -999;
  return markers.map((m, i) => {
    const pct = (m.displayMin / 2880) * 100;
    const tier = Math.abs(pct - lastPct) < 6 ? (i % 2) + 1 : 0;
    lastPct = pct;
    return { ...m, tier };
  });
}

function TimelineScale({ windows }) {
  return (
    <div className="sym-timeline-scale">
      {PRIMARY_SCALE_HOURS.map((h) => (
        <span
          key={`p-${h}`}
          className={hourTickClass(h, "primary", windows)}
          style={{ left: primaryScaleLeft(h) }}
        >
          {h < 10 ? `0${h}` : h}
        </span>
      ))}
      {EXTEND_SCALE_HOURS.map((h) => (
        <span
          key={`e-${h}`}
          className={hourTickClass(h, "extend", windows)}
          style={{ left: extendScaleLeft(h) }}
        >
          {h < 10 ? `0${h}` : h}
        </span>
      ))}
      {TAIL_SCALE_HOURS.map((h) => (
        <span
          key={`t-${h}`}
          className={hourTickClass(h, "tail", windows)}
          style={{ left: tailScaleLeft(h) }}
        >
          {h}
        </span>
      ))}
    </div>
  );
}

function TimelineRow({ label, windows, onChange, readOnly }) {
  const dragRef = useRef({ active: false, index: 0, edge: "open" });
  const createRef = useRef(null);
  const trackRef = useRef(null);
  const [createPreview, setCreatePreview] = useState(null);
  const wins = windows.length ? windows : DEFAULT_DAY_WINDOW;
  const classified = classifyWindows(wins);
  const markers = collectMarkers(wins);

  function findWindowIndex(storageMin, edge, zone) {
    return classified.findIndex(
      (w) => w.zone === zone && (edge === "open" ? w.open === storageMin : w.close === storageMin),
    );
  }

  function applyMarkerMove(clientX, trackEl, shiftKey) {
    const { minute } = minuteFromTrack(clientX, trackEl, shiftKey);
    const idx = dragRef.current.index;
    const edge = dragRef.current.edge;
    const zone = classified[idx]?.zone ?? "primary";
    onChange(
      wins.map((w, i) => {
        if (i !== idx) return { ...w };
        const next = clampMarkerMinute(minute, edge, w, zone);
        if (sessionOverlaps(wins, next.open, next.close, i)) return { ...w };
        return next;
      }),
    );
  }

  function nudgeMarker(idx, edge, delta) {
    const zone = classified[idx]?.zone ?? "primary";
    const w = wins[idx];
    const minute = edge === "open" ? w.open + delta : w.close + delta;
    onChange(
      wins.map((wi, i) => {
        if (i !== idx) return { ...wi };
        const next = clampMarkerMinute(minute, edge, wi, zone);
        if (sessionOverlaps(wins, next.open, next.close, i)) return { ...wi };
        return next;
      }),
    );
  }

  function finishCreate(clientX, shiftKey) {
    const ref = createRef.current;
    if (!ref) return;
    createRef.current = null;
    setCreatePreview(null);
    const end = minuteFromTrack(clientX, ref.track, shiftKey);
    const start = ref.start;
    if (start.zone !== end.zone) return;
    const lo = Math.min(start.minute, end.minute);
    const hi = Math.max(start.minute, end.minute);
    if (!canAddWindow(wins, lo, hi)) return;
    onChange([...wins, { open: lo, close: hi }].sort((a, b) => a.open - b.open));
  }

  useEffect(() => {
    const onUp = (e) => {
      if (createRef.current) finishCreate(e.clientX, e.shiftKey);
      dragRef.current.active = false;
    };
    function onMove(e) {
      if (createRef.current) {
        const end = minuteFromTrack(e.clientX, createRef.current.track, e.shiftKey);
        const start = createRef.current.start;
        if (start.zone !== end.zone) {
          setCreatePreview(null);
          return;
        }
        const loMin = Math.min(start.minute, end.minute);
        const hiMin = Math.max(start.minute, end.minute);
        const loDisplay = Math.min(start.displayMinute, end.displayMinute);
        const hiDisplay = Math.max(start.displayMinute, end.displayMinute);
        setCreatePreview({ open: loMin, close: hiMin, loDisplay, hiDisplay });
        return;
      }
      if (!dragRef.current.active || !dragRef.current.track) return;
      applyMarkerMove(e.clientX, dragRef.current.track, e.shiftKey);
    }
    window.addEventListener("mouseup", onUp);
    window.addEventListener("mousemove", onMove);
    return () => {
      window.removeEventListener("mouseup", onUp);
      window.removeEventListener("mousemove", onMove);
    };
  });

  function onTrackMouseDown(e) {
    if (readOnly || e.button !== 0) return;
    if (e.target.closest(".sym-timeline-marker-hit")) return;
    if (e.target.closest(".sym-timeline-track-tail")) return;
    const track = trackRef.current;
    if (!track) return;
    e.preventDefault();
    const start = minuteFromTrack(e.clientX, track, e.shiftKey);
    createRef.current = { start, track };
    setCreatePreview({
      open: start.minute,
      close: start.minute,
      loDisplay: start.displayMinute,
      hiDisplay: start.displayMinute,
    });
  }

  return (
    <div className="sym-timeline-row">
      <span className="sym-timeline-label">{label}</span>
      <div className="sym-timeline-track-wrap">
        <TimelineScale windows={wins} />
        <div
          ref={trackRef}
          className={`sym-timeline-track${readOnly ? " readonly" : " editable"}`}
          onMouseDown={onTrackMouseDown}
        >
          <div className="sym-timeline-track-zones" aria-hidden="true">
            <div className="sym-timeline-track-zone sym-timeline-track-primary" />
            <div className="sym-timeline-track-zone sym-timeline-track-extend" />
            <div className="sym-timeline-track-zone sym-timeline-track-tail" />
          </div>
          {classified.map((w, i) => (
            <div
              key={`${w.zone}-${i}-${w.open}`}
              className="sym-timeline-seg"
              style={{
                left: displayPct(toDisplayMinute(w.open, w.zone)),
                width: displayPct(toDisplayMinute(w.close, w.zone) - toDisplayMinute(w.open, w.zone)),
              }}
            />
          ))}
          {createPreview && createPreview.hiDisplay > createPreview.loDisplay && (
            <div
              className="sym-timeline-create-preview"
              style={{
                left: displayPct(createPreview.loDisplay),
                width: displayPct(createPreview.hiDisplay - createPreview.loDisplay),
              }}
            />
          )}
          {markers.map((m) => {
            const idx = findWindowIndex(m.storageMin, m.edge, m.zone);
            const draggable = !readOnly && idx >= 0;
            return (
              <div
                key={`${m.displayMin}-${m.edge}`}
                className={`sym-timeline-marker${draggable ? " sym-timeline-marker-draggable" : ""}`}
                style={{ left: displayPct(m.displayMin) }}
              >
                <span
                  className={`sym-timeline-marker-label tier-${m.tier}${m.end ? " end" : ""}`}
                >
                  {minutesToTime(m.storageMin)}
                </span>
                <span className="sym-timeline-marker-pin" />
                {draggable && (
                  <span
                    className="sym-timeline-marker-hit"
                    tabIndex={0}
                    role="slider"
                    aria-valuemin={0}
                    aria-valuemax={1440}
                    aria-valuenow={m.storageMin}
                    aria-label={`${m.edge} at ${minutesToTime(m.storageMin)}`}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      dragRef.current = {
                        active: true,
                        index: idx,
                        edge: m.edge,
                        track: trackRef.current,
                      };
                    }}
                    onKeyDown={(e) => {
                      const step = e.shiftKey ? 5 : 1;
                      if (e.key === "ArrowLeft") {
                        e.preventDefault();
                        nudgeMarker(idx, m.edge, -step);
                      } else if (e.key === "ArrowRight") {
                        e.preventDefault();
                        nudgeMarker(idx, m.edge, step);
                      }
                    }}
                  />
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

export function SessionEditorDialog({ symbol, dayIndexes, sessions, onSave, onClose }) {
  const primaryDay = dayIndexes[0];
  const [separate, setSeparate] = useState(() => hasSeparateTrade(sessions, primaryDay));
  const [quoteWins, setQuoteWins] = useState(() => cloneWindows(sessions, SESSION_QUOTE, primaryDay));
  const [tradeWins, setTradeWins] = useState(() => {
    const t = windowsForDay(sessions, SESSION_TRADE, primaryDay);
    return t.length ? t.map((w) => ({ ...w })) : cloneWindows(sessions, SESSION_QUOTE, primaryDay);
  });
  const [error, setError] = useState("");
  const close = useDialogStack(onClose);

  function handleOk() {
    const trade = separate ? tradeWins : quoteWins;
    if (!tradeWithinQuoteExtended(quoteWins, trade)) {
      setError("Trade sessions must fall within quotation sessions.");
      return;
    }
    onSave(dayIndexes, quoteWins, trade, separate);
    close();
  }

  return (
    <DialogOverlay className="sym-session-dialog-overlay" onClose={onClose}>
      <div className="sym-session-dialog" role="dialog">
        <div className="sym-session-dialog-title">
          Sessions {symbol}: {dayLabel(dayIndexes)}
        </div>
        <div className="sym-session-dialog-body">
          <div className="sym-sessions-intro sym-sessions-intro-compact">
            <span className="sym-sessions-clock-icon" aria-hidden="true" />
            <p>
              The setting up of sessions within the selected day. It is possible to determine several
              sessions of each type, trading sessions must be within the quotation ones. If specific
              trade sessions are not determined they will coincide with quotation ones.
            </p>
          </div>
          <TimelineRow label="Quotes:" windows={quoteWins} onChange={setQuoteWins} />
          <TimelineRow
            label="Trade:"
            windows={separate ? tradeWins : quoteWins}
            onChange={setTradeWins}
            readOnly={!separate}
          />
          <label className="sym-session-separate sym-session-separate-inline">
            <input type="checkbox" checked={separate} onChange={(e) => setSeparate(e.target.checked)} />
            Enable separate trading sessions
          </label>
          {error && <p className="sym-session-error">{error}</p>}
        </div>
        <div className="sym-session-dialog-footer">
          <div className="sym-session-dialog-actions">
            <button type="button" className="sym-session-ok" onClick={handleOk}>
              OK
            </button>
            <button type="button" className="sym-session-cancel" onClick={close}>
              Cancel
            </button>
          </div>
        </div>
      </div>
    </DialogOverlay>
  );
}

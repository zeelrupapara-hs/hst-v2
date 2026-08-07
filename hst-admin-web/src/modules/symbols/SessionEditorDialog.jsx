import { useEffect, useRef, useState } from "react";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import {
  DISPLAY_MIN,
  EXTEND_MIN,
  EXTEND_SCALE_HOURS,
  PRIMARY_MIN,
  PRIMARY_SCALE_HOURS,
  SESSION_QUOTE,
  SESSION_TRADE,
  TAIL_SCALE_HOURS,
  applyCloseDrag,
  applyOpenDrag,
  classifyWindows,
  dayLabel,
  displayPct,
  extendScaleLeft,
  formatScaleHour,
  gridDisplayMinutes,
  hasSeparateTrade,
  hourTickClass,
  insertSessionWindow,
  mergeSessionWindows,
  minuteFromTrack,
  minutesToTime,
  nudgeSessionMarker,
  primaryScaleLeft,
  tailScaleLeft,
  toDisplayMinute,
  tradeWithinQuoteExtended,
  usesOvernightScale,
  windowsForDay,
} from "@/lib/symbolSessions.js";

const DEFAULT_DAY_WINDOW = [{ open: 0, close: 1440 }];

function cloneWindows(sessions, type, day) {
  const wins = windowsForDay(sessions, type, day);
  return (wins.length ? wins : DEFAULT_DAY_WINDOW).map((w) => ({ ...w }));
}

const MARKER_CLUSTER_MIN = 36;

function collectMarkers(windows) {
  const classified = classifyWindows(windows.length ? windows : DEFAULT_DAY_WINDOW);
  const map = new Map();
  for (const w of classified) {
    const dOpen = toDisplayMinute(w.open, w.zone);
    const dClose = toDisplayMinute(w.close, w.zone);
    map.set(`${dOpen}-open-${w.zone}`, {
      storageMin: w.open,
      displayMin: dOpen,
      end: false,
      zone: w.zone,
      edge: "open",
    });
    map.set(`${dClose}-close-${w.zone}`, {
      storageMin: w.close,
      displayMin: dClose,
      end: w.close >= PRIMARY_MIN && w.zone === "primary",
      zone: w.zone,
      edge: "close",
    });
  }
  const markers = [...map.values()].sort((a, b) => {
    if (a.displayMin !== b.displayMin) return a.displayMin - b.displayMin;
    if (a.edge !== b.edge) return a.edge === "close" ? -1 : 1;
    if (a.end !== b.end) return a.end ? -1 : 1;
    return 0;
  });

  /** MT5: open labels above, close labels below; stack when markers crowd together. */
  const tiers = markers.map(() => 0);
  let clusterStart = 0;
  for (let i = 1; i <= markers.length; i++) {
    const span =
      i === markers.length ? Infinity : markers[i].displayMin - markers[clusterStart].displayMin;
    if (i === markers.length || span > MARKER_CLUSTER_MIN) {
      for (let j = clusterStart; j < i; j++) {
        tiers[j] = j - clusterStart;
      }
      clusterStart = i;
    }
  }

  return markers.map((m, i) => {
    const tier = tiers[i];
    const placement =
      tier > 0 ? (tier % 2 === 1 ? "below" : "above") : m.edge === "close" ? "below" : "above";
    const stack = Math.floor(tier / 2);
    return { ...m, tier, placement, stack };
  });
}

function TimelineScale({ windows }) {
  const overnight = usesOvernightScale(windows);
  const scaleTicks = [
    ...PRIMARY_SCALE_HOURS.map((h) => ({ key: `p-${h}`, left: primaryScaleLeft(h, windows) })),
    ...(overnight
      ? [
          ...EXTEND_SCALE_HOURS.map((h) => ({ key: `e-${h}`, left: extendScaleLeft(h, windows) })),
          ...TAIL_SCALE_HOURS.map((h) => ({ key: `t-${h}`, left: tailScaleLeft(h, windows) })),
        ]
      : []),
  ];
  return (
    <div className="sym-timeline-scale">
      <div className="sym-timeline-scale-ticks" aria-hidden="true">
        {scaleTicks.map(({ key, left }) => (
          <span key={key} className="sym-timeline-scale-tick" style={{ left }} />
        ))}
      </div>
      {PRIMARY_SCALE_HOURS.map((h) => (
        <span
          key={`p-${h}`}
          className={hourTickClass(h, "primary", windows)}
          style={{ left: primaryScaleLeft(h, windows) }}
        >
          {formatScaleHour(h, "primary")}
        </span>
      ))}
      {overnight &&
        EXTEND_SCALE_HOURS.map((h) => (
          <span
            key={`e-${h}`}
            className={hourTickClass(h, "extend", windows)}
            style={{ left: extendScaleLeft(h, windows) }}
          >
            {formatScaleHour(h, "extend")}
          </span>
        ))}
      {overnight &&
        TAIL_SCALE_HOURS.map((h) => (
          <span
            key={`t-${h}`}
            className={hourTickClass(h, "tail", windows)}
            style={{ left: tailScaleLeft(h, windows) }}
          >
            {formatScaleHour(h, "tail")}
          </span>
        ))}
    </div>
  );
}

function TimelineGrid({ windows }) {
  const overnight = usesOvernightScale(windows);
  return (
    <div className="sym-timeline-grid" aria-hidden="true">
      {gridDisplayMinutes(windows).map((m) => (
        <div key={m} className="sym-timeline-grid-line" style={{ left: displayPct(m, windows) }} />
      ))}
      {overnight && (
        <div className="sym-timeline-zone-divider" style={{ left: displayPct(PRIMARY_MIN, windows) }} />
      )}
      {overnight && (
        <div className="sym-timeline-zone-divider" style={{ left: displayPct(DISPLAY_MIN, windows) }} />
      )}
    </div>
  );
}

function TimelineRow({ label, windows, onChange, readOnly }) {
  const dragRef = useRef({ active: false, index: 0, edge: "open" });
  const createRef = useRef(null);
  const trackRef = useRef(null);
  const winsRef = useRef(windows);
  const onChangeRef = useRef(onChange);
  const [createPreview, setCreatePreview] = useState(null);
  const wins = windows.length ? windows : DEFAULT_DAY_WINDOW;
  const classified = classifyWindows(wins);
  const markers = collectMarkers(wins);
  const overnight = usesOvernightScale(wins);

  winsRef.current = wins;
  onChangeRef.current = onChange;

  function findWindowIndex(storageMin, edge, zone) {
    return classified.findIndex(
      (w) => w.zone === zone && (edge === "open" ? w.open === storageMin : w.close === storageMin),
    );
  }

  function applyMarkerMove(clientX, trackEl, shiftKey) {
    const clampTail = dragRef.current.edge === "close";
    const track = minuteFromTrack(clientX, trackEl, shiftKey, clampTail, winsRef.current);
    if (!track) return;
    const idx = dragRef.current.index;
    const edge = dragRef.current.edge;
    const current = winsRef.current;
    const setWins = onChangeRef.current;
    if (edge === "close") {
      setWins(applyCloseDrag(current, idx, track));
      return;
    }
    setWins(applyOpenDrag(current, idx, track));
  }

  function nudgeMarker(idx, edge, delta) {
    onChange(nudgeSessionMarker(wins, idx, edge, delta));
  }

  function finishCreate(clientX, shiftKey) {
    const ref = createRef.current;
    if (!ref) return;
    createRef.current = null;
    setCreatePreview(null);
    const end = minuteFromTrack(clientX, ref.track, shiftKey, false, winsRef.current);
    if (!end) return;
    const start = ref.start;
    if (start.zone !== end.zone) return;
    const lo = Math.min(start.minute, end.minute);
    const hi = Math.max(start.minute, end.minute);
    if (hi - lo < 1) return;
    if (end.zone === "extension" && !hasMidnightPair(wins)) return;
    onChange(insertSessionWindow(wins, lo, hi, end.zone));
  }

  function hasMidnightPair(w) {
    return w.some((win) => win.close === PRIMARY_MIN);
  }

  useEffect(() => {
    const onUp = (e) => {
      if (createRef.current) {
        finishCreate(e.clientX, e.shiftKey);
      } else if (dragRef.current.active) {
        onChangeRef.current(mergeSessionWindows(winsRef.current));
      }
      dragRef.current.active = false;
    };
    function onMove(e) {
      if (createRef.current) {
        const end = minuteFromTrack(e.clientX, createRef.current.track, e.shiftKey, false, winsRef.current);
        const start = createRef.current.start;
        if (!end || start.zone !== end.zone) {
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
    const start = minuteFromTrack(e.clientX, track, e.shiftKey, false, wins);
    if (!start) return;
    if (start.zone === "extension" && !hasMidnightPair(wins)) {
      return;
    }
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
      <div className="sym-timeline-frame">
        <div className={`sym-timeline-track-wrap${overnight ? " overnight" : ""}`}>
          <TimelineGrid windows={wins} />
          <TimelineScale windows={wins} />
          <div
            ref={trackRef}
            className={`sym-timeline-track${readOnly ? " readonly" : " editable"}${overnight ? " overnight" : " primary-only"}`}
            onMouseDown={onTrackMouseDown}
          >
            <div className="sym-timeline-track-zones" aria-hidden="true">
              {overnight ? (
                <>
                  <div className="sym-timeline-track-zone sym-timeline-track-primary" />
                  <div className="sym-timeline-track-zone sym-timeline-track-extend sym-timeline-track-extend-active" />
                  <div className="sym-timeline-track-zone sym-timeline-track-tail" />
                </>
              ) : (
                <div className="sym-timeline-track-zone sym-timeline-track-primary sym-timeline-track-full" />
              )}
            </div>
            {classified.map((w, i) => (
              <div
                key={`${w.zone}-${i}-${w.open}`}
                className="sym-timeline-seg"
                style={{
                  left: displayPct(toDisplayMinute(w.open, w.zone), wins),
                  width: displayPct(
                    toDisplayMinute(w.close, w.zone) - toDisplayMinute(w.open, w.zone),
                    wins,
                  ),
                }}
              />
            ))}
            {createPreview && createPreview.hiDisplay > createPreview.loDisplay && (
              <div
                className="sym-timeline-create-preview"
                style={{
                  left: displayPct(createPreview.loDisplay, wins),
                  width: displayPct(createPreview.hiDisplay - createPreview.loDisplay, wins),
                }}
              />
            )}
            {markers.map((m) => {
              const idx = findWindowIndex(m.storageMin, m.edge, m.zone);
              const draggable = !readOnly && idx >= 0;
              return (
                <div
                  key={`${m.displayMin}-${m.edge}-${m.zone}`}
                  className={`sym-timeline-marker${draggable ? " sym-timeline-marker-draggable" : ""}`}
                  style={{ left: displayPct(m.displayMin, wins) }}
                >
                  <span
                    className={`sym-timeline-marker-label ${m.placement}${m.stack ? ` stack-${m.stack}` : ""}${m.end ? " end" : ""}`}
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
                      aria-valuemax={m.zone === "extension" ? EXTEND_MIN : PRIMARY_MIN}
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

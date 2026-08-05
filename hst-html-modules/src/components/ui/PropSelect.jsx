import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

/**
 * MT5-style select: white square field + arrow; list opens downward as overlay (no layout shift).
 */
export function PropSelect({ value, options, onChange, wide, narrow }) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);

  useEffect(() => {
    if (!open) return;
    function close(e) {
      if (rootRef.current?.contains(e.target)) return;
      if (e.target.closest?.(".prop-select-menu")) return;
      setOpen(false);
    }
    function onKey(e) {
      if (e.key === "Escape") setOpen(false);
    }
    const t = setTimeout(() => {
      document.addEventListener("mousedown", close);
    }, 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  function openMenu(e) {
    e.stopPropagation();
    const el = rootRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: r.width });
    setOpen(true);
  }

  function pick(opt) {
    onChange(opt);
    setOpen(false);
  }

  const cls = [
    "prop-select",
    wide && "prop-select-wide",
    narrow && "prop-select-narrow",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div ref={rootRef} className={cls} onClick={(e) => e.stopPropagation()}>
        <div className="prop-select-box">
          <span className="prop-select-value">{value}</span>
          <button
            type="button"
            className="prop-select-btn"
            aria-label="Open list"
            onClick={openMenu}
          />
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <ul
            className="prop-select-menu"
            style={{
              position: "fixed",
              top: pos.top,
              left: pos.left,
              minWidth: pos.width,
            }}
          >
            {options.map((opt) => (
              <li
                key={opt}
                className={opt === value ? "sel" : undefined}
                onMouseDown={(e) => {
                  e.preventDefault();
                  pick(opt);
                }}
              >
                {opt}
              </li>
            ))}
          </ul>,
          document.body
        )}
    </>
  );
}

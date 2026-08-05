import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

function normalizeOptions(options) {
  return (options || []).map((o) =>
    typeof o === "string" ? { value: o, label: o } : o
  );
}

/**
 * MT5-style select: white field + arrow; list opens downward as overlay.
 * Options: string[] or { value, label, description? }[]
 */
export function PropSelect({
  value,
  options,
  onChange,
  wide,
  narrow,
  fill,
  disabled,
  initialOpen,
  className = "",
}) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);
  const items = normalizeOptions(options);
  const current =
    items.find((o) => String(o.value) === String(value)) ?? items[0];
  const display = current?.label ?? value ?? "";

  useEffect(() => {
    if (!initialOpen || disabled) return;
    const el = rootRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: r.width });
    setOpen(true);
  }, [initialOpen, disabled]);

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
    if (disabled) return;
    e.stopPropagation();
    const el = rootRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: r.width });
    setOpen(true);
  }

  function pick(opt) {
    if (disabled) return;
    onChange?.(opt.value);
    setOpen(false);
  }

  const cls = [
    "prop-select",
    wide && "prop-select-wide",
    narrow && "prop-select-narrow",
    fill && "prop-select-fill",
    disabled && "prop-select-disabled",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div ref={rootRef} className={cls} onClick={(e) => e.stopPropagation()}>
        <div className="prop-select-box">
          <span className="prop-select-value" title={display}>
            {display}
          </span>
          <button
            type="button"
            className="prop-select-btn"
            aria-label="Open list"
            disabled={disabled}
            onClick={openMenu}
          />
        </div>
      </div>
      {open &&
        !disabled &&
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
            {items.map((opt) => (
              <li
                key={String(opt.value)}
                className={
                  String(opt.value) === String(value) ? "sel" : undefined
                }
                title={opt.description || undefined}
                onMouseDown={(e) => {
                  e.preventDefault();
                  pick(opt);
                }}
              >
                {opt.label}
              </li>
            ))}
          </ul>,
          document.body
        )}
    </>
  );
}

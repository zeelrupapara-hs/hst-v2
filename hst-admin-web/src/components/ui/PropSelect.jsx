import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

const normalize = (options) =>
  (options || []).map((o) => (typeof o === "object" ? o : { value: o, label: String(o) }));

/**
 * Platform select: white field + arrow button; list opens as a fixed overlay.
 * @param {{value: any, options: Array<string|number|{value:any,label:string}>, onChange?: Function, fill?: boolean, disabled?: boolean, className?: string}} props
 */
export function PropSelect({ value, options, onChange, fill, disabled, className = "" }) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);
  const items = normalize(options);
  const current = items.find((o) => String(o.value) === String(value)) ?? items[0];

  useEffect(() => {
    if (!open) return;
    function close(e) {
      if (rootRef.current?.contains(e.target)) return;
      if (e.target.closest?.(".prop-select-menu")) return;
      setOpen(false);
    }
    const onKey = (e) => e.key === "Escape" && setOpen(false);
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
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
    const r = rootRef.current.getBoundingClientRect();
    setPos({ top: r.bottom, left: r.left, width: r.width });
    setOpen(true);
  }

  const cls = ["prop-select", fill && "prop-select-fill", disabled && "prop-select-disabled", className]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div ref={rootRef} className={cls} onClick={(e) => e.stopPropagation()}>
        <div className="prop-select-box" onClick={openMenu}>
          <span className="prop-select-value" title={current?.label}>
            {current?.label ?? ""}
          </span>
          <button type="button" className="prop-select-btn" aria-label="Open list" disabled={disabled} onClick={openMenu} />
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <ul
            className="prop-select-menu"
            style={{ position: "fixed", top: pos.top, left: pos.left, minWidth: pos.width }}
          >
            {items.map((opt) => (
              <li
                key={String(opt.value)}
                className={String(opt.value) === String(value) ? "sel" : undefined}
                onMouseDown={(e) => {
                  e.preventDefault();
                  onChange?.(opt.value);
                  setOpen(false);
                }}
              >
                {opt.label}
              </li>
            ))}
          </ul>,
          document.body,
        )}
    </>
  );
}

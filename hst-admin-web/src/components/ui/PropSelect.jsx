import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useExclusiveDropdown } from "@/hooks/useExclusiveDropdown.js";

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
  const menuRef = useRef(null);
  const openRef = useRef(false);
  const items = normalize(options);
  const current = items.find((o) => String(o.value) === String(value)) ?? items[0];
  const announceOpen = useExclusiveDropdown(open, setOpen);

  useEffect(() => {
    openRef.current = open;
  }, [open]);

  useEffect(() => {
    if (!open) return;
    function close(e) {
      if (rootRef.current?.contains(e.target)) return;
      if (menuRef.current?.contains(e.target)) return;
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
    e.preventDefault();
    e.stopPropagation();
    if (openRef.current) {
      setOpen(false);
      return;
    }
    announceOpen();
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
        <div className="prop-select-box" onMouseDown={openMenu}>
          <span className="prop-select-value" title={current?.label}>
            {current?.label ?? ""}
          </span>
          <span className="prop-select-btn-wrap">
            <button
              type="button"
              className="prop-select-btn"
              aria-label="Open list"
              disabled={disabled}
              tabIndex={-1}
            />
          </span>
        </div>
      </div>
      {open &&
        pos &&
        createPortal(
          <ul
            ref={menuRef}
            className="prop-select-menu prop-select-menu-front"
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

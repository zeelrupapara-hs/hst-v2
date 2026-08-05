import { useEffect } from "react";
import { createPortal } from "react-dom";

/**
 * The list context menu: Add/Edit/Delete and friends, opened at the cursor.
 * @param {{x: number, y: number, items: Array<{label: string, onClick?: Function, disabled?: boolean}|"sep">, onClose: Function}} props
 */
export function ContextMenu({ x, y, items, onClose }) {
  useEffect(() => {
    const close = () => onClose();
    const onKey = (e) => e.key === "Escape" && onClose();
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [onClose]);

  return createPortal(
    <ul className="ctx-menu" style={{ position: "fixed", top: y, left: x, zIndex: 1000 }}>
      {items.map((item, i) =>
        item === "sep" ? (
          <li key={i} className="ctx-sep" />
        ) : (
          <li
            key={item.label}
            className={`ctx-item${item.disabled ? " disabled" : ""}`}
            onMouseDown={(e) => {
              e.preventDefault();
              if (item.disabled) return;
              onClose();
              item.onClick?.();
            }}
          >
            <span className="ctx-label">{item.label}</span>
          </li>
        ),
      )}
    </ul>,
    document.body,
  );
}

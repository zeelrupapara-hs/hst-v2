import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { Icon } from "@/components/ui/Icon.jsx";

/**
 * The list context menu. An item is {label, icon?, shortcut?, checked?, disabled?, items?, onClick}
 * or the string "sep". Unavailable commands are greyed, not hidden, as the reference does.
 */
export function ContextMenu({ x, y, items, onClose }) {
  const [openSub, setOpenSub] = useState(null);

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

  const pick = (item) => {
    if (item.disabled || item.items) return;
    onClose();
    item.onClick?.();
  };

  return createPortal(
    <ul className="ctx-menu ctx-menu-front" style={{ position: "fixed", top: y, left: x }}>
      {items.map((item, i) =>
        item === "sep" ? (
          <li key={`sep-${i}`} className="ctx-sep" />
        ) : (
          <li
            key={item.label}
            className={`ctx-item${item.disabled ? " disabled" : ""}`}
            onMouseEnter={() => setOpenSub(item.items ? item.label : null)}
            onMouseDown={(e) => {
              e.preventDefault();
              pick(item);
            }}
          >
            <span className="ctx-icon">
              {item.checked ? "✔" : item.icon ? <Icon id={item.icon} /> : null}
            </span>
            <span className="ctx-label">{item.label}</span>
            {item.shortcut && <span className="ctx-shortcut">{item.shortcut}</span>}
            {item.items && <span className="ctx-arrow">▸</span>}
            {item.items && openSub === item.label && (
              <ul className="ctx-menu ctx-submenu">
                {item.items.map((sub) => (
                  <li
                    key={sub.label}
                    className={`ctx-item${sub.disabled ? " disabled" : ""}`}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      if (sub.disabled) return;
                      onClose();
                      sub.onClick?.();
                    }}
                  >
                    <span className="ctx-icon">{sub.checked ? "✔" : null}</span>
                    <span className="ctx-label">{sub.label}</span>
                  </li>
                ))}
              </ul>
            )}
          </li>
        ),
      )}
    </ul>,
    document.body,
  );
}

/**
 * The tail every list menu shares, in reference order. `on` supplies the handlers a module
 * actually has; anything missing renders greyed rather than disappearing.
 */
export function listMenuTail({ on = {}, view = {}, extras = [] } = {}) {
  return [
    "sep",
    { label: "Move Up", icon: "", disabled: !on.moveUp, onClick: on.moveUp },
    { label: "Move Down", disabled: !on.moveDown, onClick: on.moveDown },
    { label: on.sortLabel || "Sort Alphabetically", disabled: !on.sort, onClick: on.sort },
    ...extras,
    "sep",
    { label: "Export to File", disabled: !on.exportFile, onClick: on.exportFile },
    { label: "Import from File", disabled: !on.importFile, onClick: on.importFile },
    "sep",
    { label: "Journal", disabled: !on.journal, onClick: on.journal },
    { label: "Find", shortcut: "Ctrl+F", disabled: !on.find, onClick: on.find },
    "sep",
    { label: "Auto Arrange", checked: view.autoArrange !== false, onClick: on.toggleAutoArrange },
    { label: "Grid", checked: view.grid !== false, onClick: on.toggleGrid },
  ];
}

/** Add / Edit / Delete, with the reference shortcuts, as the head of every list menu. */
export function listMenuHead({ addLabel = "Add", onAdd, onEdit, onDelete, hasSelection }) {
  return [
    { label: addLabel, icon: "add", shortcut: "Ctrl+N", onClick: onAdd },
    { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: !hasSelection, onClick: onEdit },
    { label: "Delete", icon: "delete", shortcut: "Ctrl+D", disabled: !hasSelection, onClick: onDelete },
  ];
}

/** Grid/table context menu head — MT5 uses Insert / Enter / Delete on editable tables. */
export function gridMenuHead({ onAdd, onEdit, onDelete, hasSelection }) {
  return [
    { label: "Add", icon: "add", shortcut: "Insert", onClick: onAdd },
    { label: "Edit", icon: "edit", shortcut: "Enter", disabled: !hasSelection, onClick: onEdit },
    { label: "Delete", icon: "delete", shortcut: "Delete", disabled: !hasSelection, onClick: onDelete },
  ];
}

/** Grid/table context menu tail — Select All, Copy, Find. Handlers optional; missing ones render greyed. */
export function gridMenuTail({ onSelectAll, onCopy, onFind, onFindNext, onFindPrev } = {}) {
  return [
    "sep",
    { label: "Select All", shortcut: "Ctrl+A", disabled: !onSelectAll, onClick: onSelectAll },
    { label: "Copy", shortcut: "Ctrl+C", disabled: !onCopy, onClick: onCopy },
    "sep",
    { label: "Find", shortcut: "Ctrl+F", disabled: !onFind, onClick: onFind },
    { label: "Find Next", shortcut: "F3", disabled: !onFindNext, onClick: onFindNext },
    { label: "Find Previous", shortcut: "Shift+F3", disabled: !onFindPrev, onClick: onFindPrev },
  ];
}

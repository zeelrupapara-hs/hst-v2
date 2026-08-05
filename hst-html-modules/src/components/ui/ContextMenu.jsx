import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Icon } from "./Icon.jsx";

/**
 * @typedef {object} ContextMenuItem
 * @property {string} [id]
 * @property {"separator"} [type]
 * @property {string} [label]
 * @property {string} [iconId]
 * @property {string} [icon]
 * @property {string} [shortcut]
 * @property {boolean} [disabled]
 * @property {boolean} [check]
 * @property {boolean} [checked]
 * @property {ContextMenuItem[]|(() => ContextMenuItem[])} [children]
 * @property {() => void} [onClick]
 */

function resolveItems(items) {
  return typeof items === "function" ? items() : items || [];
}

function MenuRow({ item, onClose, onSubmenu }) {
  if (item.type === "separator") {
    return <div className="ctx-sep" role="separator" />;
  }

  const disabled = !!item.disabled;
  const hasChildren = item.children?.length > 0;

  return (
    <div
      className={`ctx-item${disabled ? " disabled" : ""}${hasChildren ? " has-children" : ""}`}
      onClick={(e) => {
        e.stopPropagation();
        if (disabled || hasChildren) return;
        item.onClick?.();
        onClose();
      }}
      onMouseEnter={(e) => {
        if (!disabled && hasChildren) onSubmenu(item, e.currentTarget);
      }}
    >
      <span className="ctx-icon">
        {item.iconId ? (
          <Icon name={item.iconId} size={16} />
        ) : (
          item.icon || ""
        )}
      </span>
      <span className="ctx-label">{item.label}</span>
      {item.check ? (
        <span className="ctx-check">{item.checked ? "✓" : ""}</span>
      ) : null}
      {item.shortcut ? (
        <span className="ctx-shortcut">{item.shortcut}</span>
      ) : null}
      {hasChildren ? <span className="ctx-arrow">▶</span> : null}
    </div>
  );
}

function SubmenuPanel({ items, anchor, onClose }) {
  const ref = useRef(null);
  const [pos, setPos] = useState({ left: 0, top: 0 });

  useEffect(() => {
    const el = ref.current;
    const rect = anchor.getBoundingClientRect();
    let left = rect.right - 2;
    let top = rect.top;
    if (el) {
      const mb = el.getBoundingClientRect();
      if (left + mb.width > window.innerWidth) left = rect.left - mb.width + 2;
      if (top + mb.height > window.innerHeight) {
        top = window.innerHeight - mb.height - 4;
      }
    }
    setPos({ left, top });
  }, [anchor]);

  return createPortal(
    <div
      ref={ref}
      className="ctx-menu ctx-submenu"
      style={{ left: pos.left, top: pos.top }}
      onClick={(e) => e.stopPropagation()}
    >
      {resolveItems(items).map((item, i) => (
        <MenuRow
          key={item.id || item.label || i}
          item={item}
          onClose={onClose}
          onSubmenu={() => {}}
        />
      ))}
    </div>,
    document.body
  );
}

export function ContextMenu({ x, y, items, onClose }) {
  const ref = useRef(null);
  const [submenu, setSubmenu] = useState(null);

  useEffect(() => {
    function onDocDown(e) {
      if (ref.current?.contains(e.target)) return;
      if (document.querySelector(".ctx-submenu")?.contains(e.target)) return;
      onClose();
    }
    function onKeyDown(e) {
      if (e.key === "Escape") onClose();
    }
    document.addEventListener("mousedown", onDocDown, true);
    document.addEventListener("keydown", onKeyDown, true);
    document.addEventListener("scroll", onClose, true);
    return () => {
      document.removeEventListener("mousedown", onDocDown, true);
      document.removeEventListener("keydown", onKeyDown, true);
      document.removeEventListener("scroll", onClose, true);
    };
  }, [onClose]);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    let left = x;
    let top = y;
    const rect = el.getBoundingClientRect();
    if (left + rect.width > window.innerWidth) left = x - rect.width;
    if (top + rect.height > window.innerHeight) top = y - rect.height;
    el.style.left = `${Math.max(4, left)}px`;
    el.style.top = `${Math.max(4, top)}px`;
  }, [x, y, items]);

  function openSubmenu(item, anchor) {
    const kids = resolveItems(item.children);
    if (!kids.length) {
      setSubmenu(null);
      return;
    }
    setSubmenu({ items: kids, anchor });
  }

  return createPortal(
    <>
      <div
        ref={ref}
        className="ctx-menu"
        style={{ left: x, top: y }}
        onClick={(e) => e.stopPropagation()}
      >
        {resolveItems(items).map((item, i) => (
          <MenuRow
            key={item.id || item.label || i}
            item={item}
            onClose={onClose}
            onSubmenu={openSubmenu}
          />
        ))}
      </div>
      {submenu && (
        <SubmenuPanel
          items={submenu.items}
          anchor={submenu.anchor}
          onClose={onClose}
        />
      )}
    </>,
    document.body
  );
}

/** @returns {{ menu: {x:number,y:number,items:ContextMenuItem[]}|null, show:(e:MouseEvent, items:ContextMenuItem[])=>void, close:()=>void }} */
export function useContextMenu() {
  const [menu, setMenu] = useState(null);

  function show(e, items) {
    e.preventDefault();
    e.stopPropagation();
    setMenu({ x: e.clientX, y: e.clientY, items });
  }

  function close() {
    setMenu(null);
  }

  return { menu, show, close };
}

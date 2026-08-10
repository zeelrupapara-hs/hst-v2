import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Icon } from "@/components/ui/Icon.jsx";
import { useGroups } from "@/hooks/useGroups.js";

// The group tree, built from the backslash paths the groups already carry.
function buildTree(groups) {
  const root = { name: "Groups", children: new Map(), leaves: [] };
  for (const g of groups) {
    const parts = String(g.group).split("\\").filter(Boolean);
    if (!parts.length) continue;
    let node = root;
    for (const part of parts.slice(0, -1)) {
      if (!node.children.has(part)) node.children.set(part, { name: part, children: new Map(), leaves: [] });
      node = node.children.get(part);
    }
    node.leaves.push({ name: parts[parts.length - 1], path: g.group });
  }
  return root;
}

function Branch({ node, path, depth, open, toggle, onPick, maskable }) {
  const key = path.join("\\");
  const expanded = open.has(key);
  return (
    <>
      <div className="grptree-row" style={{ paddingLeft: depth * 14 }}>
        <button
          type="button"
          className="grptree-arrow"
          aria-label={expanded ? "collapse" : "expand"}
          onClick={(e) => {
            e.stopPropagation();
            toggle(key);
          }}
        >
          {expanded ? "▾" : "▸"}
        </button>
        <button
          type="button"
          className="grptree-item grptree-folder"
          onClick={(e) => {
            e.stopPropagation();
            if (maskable) onPick(depth === 0 ? "*" : `${key}\\*`);
            else toggle(key);
          }}
        >
          <Icon id="folder" size={12} /> {node.name}
        </button>
      </div>
      {expanded && (
        <>
          {[...node.children.values()]
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((child) => (
              <Branch
                key={child.name}
                node={child}
                path={[...path, child.name]}
                depth={depth + 1}
                open={open}
                toggle={toggle}
                onPick={onPick}
                maskable={maskable}
              />
            ))}
          {node.leaves
            .slice()
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((g) => (
              <div key={g.path} className="grptree-row" style={{ paddingLeft: (depth + 1) * 14 + 16 }}>
                <button type="button" className="grptree-item" onClick={(e) => { e.stopPropagation(); onPick(g.path); }}>
                  <Icon id="groups" size={12} /> {g.name}
                </button>
              </div>
            ))}
        </>
      )}
    </>
  );
}

/**
 * The group picker as a tree, everywhere a group is chosen. A leaf picks that group; with
 * maskable set, a folder picks its whole branch as a mask, and the field can be typed into.
 */
export function GroupTreeSelect({ value, onChange, maskable = false }) {
  const { groups } = useGroups();
  const [openList, setOpenList] = useState(false);
  const [open, setOpen] = useState(() => new Set([""]));
  const [pos, setPos] = useState(null);
  const rootRef = useRef(null);

  useEffect(() => {
    if (!openList) return;
    const close = (e) => {
      if (rootRef.current?.contains(e.target)) return;
      if (e.target.closest?.(".grptree-pop")) return;
      setOpenList(false);
    };
    const onKey = (e) => e.key === "Escape" && setOpenList(false);
    const t = setTimeout(() => document.addEventListener("mousedown", close), 0);
    document.addEventListener("keydown", onKey);
    return () => {
      clearTimeout(t);
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", onKey);
    };
  }, [openList]);

  const toggle = (key) =>
    setOpen((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });

  function toggleList(e) {
    e?.stopPropagation();
    if (!openList && rootRef.current) {
      const r = rootRef.current.getBoundingClientRect();
      setPos({ top: r.bottom + 1, left: r.left, minWidth: r.width });
    }
    setOpenList(!openList);
  }

  const pick = (v) => {
    onChange(v);
    setOpenList(false);
  };

  return (
    <span ref={rootRef} className="grptree-select" onClick={(e) => e.stopPropagation()}>
      {maskable ? (
        <input type="text" value={value ?? ""} onChange={(e) => onChange(e.target.value)} />
      ) : (
        <input type="text" readOnly value={value ?? ""} onClick={toggleList} />
      )}
      <button type="button" className="grptree-drop" aria-label="browse groups" onClick={toggleList}>
        ▾
      </button>
      {openList &&
        pos &&
        createPortal(
          <div className="grptree-pop" style={{ position: "fixed", ...pos }}>
            <Branch
              node={buildTree(groups)}
              path={[]}
              depth={0}
              open={open}
              toggle={toggle}
              onPick={pick}
              maskable={maskable}
            />
          </div>,
          document.body,
        )}
    </span>
  );
}

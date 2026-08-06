import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Icon } from "@/components/ui/Icon.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { navIcon } from "@/lib/icons.js";
import {
  isSymbolFolderNode,
  isSymbolsNavNode,
  symbolsNavFolderPath,
  symbolsNavMenuItems,
} from "@/lib/navContextMenus.js";
import { createSymbolFolder, deleteSymbolFolder } from "@/api/endpoints/symbols.js";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useSymbolFolders } from "@/hooks/useSymbolFolders.js";
import { useDatafeeds } from "@/hooks/useDatafeeds.js";
import { useGroups } from "@/hooks/useGroups.js";
import { useSession } from "@/hooks/useSession.js";
import { buildGroupTree, sortedGroupKeys } from "@/lib/groupTree.js";
import {
  annotateTreeCounts,
  buildFolderTree,
  emptyFolderPathsToMap,
  sortedTreeChildKeys,
} from "@/lib/symbolTree.js";

function groupFolderChildren(folderNode) {
  return sortedGroupKeys(folderNode).map((key) => {
    const child = folderNode.children[key];
    const node = {
      key: `groupfolder:${child.path}`,
      label: child.name,
      route: `/groups?folder=${encodeURIComponent(child.path)}`,
      count: child.count || undefined,
    };
    if (sortedGroupKeys(child).length) node.children = groupFolderChildren(child);
    return node;
  });
}

function symbolFolderChildren(folderNode, emptyFolders) {
  const parentPath = folderNode.isRoot ? "" : folderNode.path;
  const seen = new Set(sortedTreeChildKeys(folderNode));

  const nodes = sortedTreeChildKeys(folderNode).map((key) => {
    const child = folderNode.children[key];
    const node = {
      key: `symfolder:${child.path}`,
      label: child.name,
      route: `/symbols?folder=${encodeURIComponent(child.path)}`,
      count: child.count || undefined,
    };
    if (sortedTreeChildKeys(child).length || emptyFolders[child.path]?.length) {
      node.children = symbolFolderChildren(child, emptyFolders);
    }
    return node;
  });

  for (const name of emptyFolders[parentPath] || []) {
    if (seen.has(name)) continue;
    const path = parentPath ? `${parentPath}\\${name}` : name;
    nodes.push({
      key: `symfolder:${path}`,
      label: name,
      route: `/symbols?folder=${encodeURIComponent(path)}`,
      count: 0,
      children: symbolFolderChildren({ path, children: {} }, emptyFolders),
    });
  }

  return nodes.sort((a, b) => a.label.localeCompare(b.label, undefined, { sensitivity: "base" }));
}

function shapeTree(nav, symbols = [], datafeeds = [], groups = [], emptyFolders = {}) {
  const nodes = (nav?.nodes ?? []).map((n) => {
    if (n.key === "groups" && groups.length) {
      const root = buildGroupTree(groups);
      return { ...n, route: "/groups", count: root.count, children: groupFolderChildren(root) };
    }
    if (n.key === "symbols") {
      const root = annotateTreeCounts(buildFolderTree(symbols), symbols);
      return {
        ...n,
        count: root.count,
        children: symbolFolderChildren(root, emptyFolders),
      };
    }
    if (n.key === "datafeeds" && datafeeds?.length) {
      return {
        ...n,
        count: datafeeds.length,
        children: datafeeds.map((f) => ({
          key: `feed:${f.datafeed_id}`,
          label: f.name,
          route: `/datafeeds?feed=${f.datafeed_id}`,
          offline: f.enable !== 1 || f.sys_connection !== 1,
        })),
      };
    }
    return n;
  });
  const feeds = nodes.filter((n) => n.section === "feeds");
  const rest = nodes.filter((n) => n.section !== "feeds");

  const children = [...rest];
  if (feeds.length) {
    children.splice(rest.length, 0, {
      key: "integrations",
      label: "Integrations",
      route: "",
      children: feeds,
    });
  }

  return [
    {
      key: "servers",
      label: "Servers",
      route: "",
      children: [{ key: "server", label: "Trade Server", route: "", children }],
    },
  ];
}

function NavNode({ node, panel, depth, pendingFolder, onCommitFolder, onCancelFolder, onContextMenu }) {
  const navigate = useNavigate();
  const location = useLocation();
  const [open, setOpen] = useState(depth < 3);

  const folderPath = symbolsNavFolderPath(node);
  const showPendingAdd = pendingFolder && pendingFolder.parentPath === folderPath;
  const hasChildren = node.children?.length > 0 || showPendingAdd;

  useEffect(() => {
    if (showPendingAdd) setOpen(true);
  }, [showPendingAdd]);

  const to = node.route ? `/${panel}${node.route}` : null;
  const isActive = to !== null && location.pathname + location.search === to;

  function onRowClick() {
    if (to) navigate(to);
    else if (hasChildren) setOpen(!open);
  }

  function onRowContextMenu(e) {
    if (!isSymbolsNavNode(node)) return;
    e.preventDefault();
    e.stopPropagation();
    onContextMenu(node, e);
  }

  return (
    <li className="nav-branch">
      <div
        className={`nav-item nav-indent-${depth}${isActive ? " active" : ""}`}
        role={to ? "link" : "button"}
        onClick={onRowClick}
        onContextMenu={onRowContextMenu}
      >
        <span
          className={`chevron${hasChildren ? "" : " empty"}`}
          onClick={(e) => {
            e.stopPropagation();
            if (hasChildren) setOpen(!open);
          }}
        >
          {hasChildren ? (open ? "▼" : "▶") : "▶"}
        </span>
        <Icon id={navIcon(node.key)} title={node.label} />
        <span className={`label${node.offline ? " nav-feed-disabled" : ""}`}>{node.label}</span>
        {node.count != null && <span className="count">({node.count})</span>}
      </div>
      {hasChildren && open && (
        <ul>
          {node.children?.map((child) => (
            <NavNode
              key={child.key}
              node={child}
              panel={panel}
              depth={depth + 1}
              pendingFolder={pendingFolder}
              onCommitFolder={onCommitFolder}
              onCancelFolder={onCancelFolder}
              onContextMenu={onContextMenu}
            />
          ))}
          {showPendingAdd && (
            <li className="nav-branch">
              <div className={`nav-item nav-indent-${depth + 1} nav-item-editing`}>
                <span className="chevron empty">▶</span>
                <Icon id="folder" title="New group" />
                <input
                  className="nav-folder-input"
                  value={pendingFolder.draft}
                  placeholder="New group"
                  autoFocus
                  onFocus={(e) => e.target.select()}
                  onChange={(e) => onCommitFolder({ ...pendingFolder, draft: e.target.value }, false)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      onCommitFolder({ ...pendingFolder, draft: e.target.value }, true);
                    }
                    if (e.key === "Escape") onCancelFolder();
                  }}
                  onBlur={(e) => {
                    const draft = e.target.value.trim();
                    if (!draft) {
                      onCancelFolder();
                      return;
                    }
                    onCommitFolder({ ...pendingFolder, draft }, true);
                  }}
                />
              </div>
            </li>
          )}
        </ul>
      )}
    </li>
  );
}

function DeleteFolderConfirm({ confirmDelete, onConfirm, onClose }) {
  const close = useDialogStack(onClose);
  return (
    <div className="config-window sym-confirm">
      <div className="config-title">
        <span className="config-title-text">Delete</span>
      </div>
      <div className="config-body">
        <p>
          {confirmDelete.cascade
            ? `Delete folder '${confirmDelete.path}' and all ${confirmDelete.count} symbol(s) beneath it?`
            : `Delete empty folder '${confirmDelete.path}'?`}
        </p>
      </div>
      <div className="config-actions">
        <button type="button" className="config-ok" onClick={() => { onConfirm(confirmDelete); close(); }}>
          OK
        </button>
        <button type="button" onClick={close}>
          Cancel
        </button>
      </div>
    </div>
  );
}

export function NavTree({ nav, panel }) {
  const session = useSession();
  const navigate = useNavigate();
  const location = useLocation();
  const { symbols, reload: reloadSymbols } = useSymbols();
  const { folders, reload: reloadFolders } = useSymbolFolders();
  const { datafeeds } = useDatafeeds();
  const { groups } = useGroups();
  const [menu, setMenu] = useState(null);
  const [pendingFolder, setPendingFolder] = useState(null);
  const [confirmDelete, setConfirmDelete] = useState(null);

  const canEditSymbols = session.can?.right_cfg_symbols !== false;
  const emptyFolders = useMemo(() => emptyFolderPathsToMap(folders), [folders]);

  const refreshTree = useCallback(async () => {
    await Promise.all([reloadSymbols(), reloadFolders(), session.reload?.()]);
  }, [reloadSymbols, reloadFolders, session]);

  const openFolder = useCallback(
    (node) => {
      const path = symbolsNavFolderPath(node);
      if (path === null) return;
      setMenu(null);
      navigate(`/${panel}/symbols?folder=${encodeURIComponent(path)}`);
    },
    [navigate, panel],
  );

  const startAddFolder = useCallback((node) => {
    const parentPath = symbolsNavFolderPath(node);
    if (parentPath === null) return;
    setPendingFolder({ parentPath, draft: "" });
    setMenu(null);
  }, []);

  const commitFolder = useCallback(
    async (pending, finalize) => {
      if (!finalize) {
        setPendingFolder(pending);
        return;
      }

      const name = pending.draft.trim();
      setPendingFolder(null);
      if (!name) return;

      const res = await createSymbolFolder(pending.parentPath, name);
      if (!res.ok) {
        window.alert(res.message || "create folder failed");
        return;
      }
      await refreshTree();
    },
    [refreshTree],
  );

  const requestDeleteFolder = useCallback((node) => {
    const path = symbolsNavFolderPath(node);
    if (!path || !isSymbolFolderNode(node)) return;
    setMenu(null);
    const count = node.count ?? 0;
    setConfirmDelete({ path, count, cascade: count > 0 });
  }, []);

  const doDeleteFolder = useCallback(async ({ path, cascade }) => {
    const res = await deleteSymbolFolder(path, cascade);
    if (!res.ok) {
      window.alert(res.message || "delete folder failed");
      return;
    }

    const prefix = `folder=${encodeURIComponent(path)}`;
    if (location.search.includes(prefix)) {
      navigate(`/${panel}/symbols`, { replace: true });
    }
    await refreshTree();
  }, [location.search, navigate, panel, refreshTree]);

  const cancelFolder = useCallback(() => setPendingFolder(null), []);

  const onContextMenu = useCallback(
    (node, e) => {
      const isFolder = isSymbolFolderNode(node);
      setMenu({
        x: e.clientX,
        y: e.clientY,
        node,
        items: symbolsNavMenuItems({
          canEdit: canEditSymbols,
          isFolder,
          onAdd: () => startAddFolder(node),
          onEdit: isFolder ? () => openFolder(node) : undefined,
          onDelete: isFolder ? () => requestDeleteFolder(node) : undefined,
          onRefresh: refreshTree,
        }),
      });
    },
    [canEditSymbols, openFolder, refreshTree, startAddFolder, requestDeleteFolder],
  );

  return (
    <aside className="navigator">
      <div className="nav-header">Navigator</div>
      <ul className="nav-tree">
        {shapeTree(nav, symbols, datafeeds, groups, emptyFolders).map((node) => (
          <NavNode
            key={node.key}
            node={node}
            panel={panel}
            depth={0}
            pendingFolder={pendingFolder}
            onCommitFolder={commitFolder}
            onCancelFolder={cancelFolder}
            onContextMenu={onContextMenu}
          />
        ))}
      </ul>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menu.items} onClose={() => setMenu(null)} />
      )}
      {confirmDelete && (
        <DialogOverlay>
          <DeleteFolderConfirm
            confirmDelete={confirmDelete}
            onConfirm={doDeleteFolder}
            onClose={() => setConfirmDelete(null)}
          />
        </DialogOverlay>
      )}
    </aside>
  );
}

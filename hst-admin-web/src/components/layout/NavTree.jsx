import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Icon } from "@/components/ui/Icon.jsx";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { navIcon } from "@/lib/icons.js";
import {
  isGroupFolderNode,
  isGroupsNavNode,
  datafeedsNavMenuItems,
  isDatafeedsNavNode,
  isSymbolFolderNode,
  isSymbolsNavNode,
  groupsNavFolderPath,
  groupsNavMenuItems,
  symbolsNavFolderPath,
  symbolsNavMenuItems,
} from "@/lib/navContextMenus.js";
import { createSymbolFolder, deleteSymbolFolder, renameSymbolFolder, updateSymbol } from "@/api/endpoints/symbols.js";
import { deleteGroup } from "@/api/endpoints/groups.js";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useSymbolFolders } from "@/hooks/useSymbolFolders.js";
import { useDatafeeds } from "@/hooks/useDatafeeds.js";
import { useGroups } from "@/hooks/useGroups.js";
import { useSession } from "@/hooks/useSession.js";
import { buildGroupTree, groupsUnderFolder, sortedGroupKeys } from "@/lib/groupTree.js";
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
    if (n.key === "groups") {
      const root = buildGroupTree(groups);
      return { ...n, route: "/groups", count: root.count || undefined, children: groupFolderChildren(root) };
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
          editRoute: `/datafeeds?feed=${f.datafeed_id}&edit=1`,
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

const SYMBOLS_DRAG = "application/x-hst-symbols";

function NavNode({ node, panel, depth, pendingFolder, onCommitFolder, onCancelFolder, onContextMenu, onDropSymbols }) {
  const navigate = useNavigate();
  const location = useLocation();
  const [open, setOpen] = useState(depth < 3);
  const [dragOver, setDragOver] = useState(false);

  const folderPath = symbolsNavFolderPath(node);
  const dropTarget = onDropSymbols && isSymbolFolderNode(node);
  const showPendingAdd = pendingFolder?.mode === "add" && pendingFolder.parentPath === folderPath;
  const showPendingRename = pendingFolder?.mode === "rename" && pendingFolder.fromPath === folderPath;
  const hasChildren = node.children?.length > 0 || showPendingAdd;

  useEffect(() => {
    if (showPendingAdd || showPendingRename) setOpen(true);
  }, [showPendingAdd, showPendingRename]);

  const to = node.route ? `/${panel}${node.route}` : null;
  const isActive = to !== null && location.pathname + location.search === to;

  function onRowClick() {
    if (to) navigate(to);
    else if (hasChildren) setOpen(!open);
  }

  function onRowContextMenu(e) {
    if (!isSymbolsNavNode(node) && !isGroupsNavNode(node) && !isDatafeedsNavNode(node)) return;
    e.preventDefault();
    e.stopPropagation();
    onContextMenu(node, e);
  }

  function onDragOver(e) {
    if (!dropTarget || !e.dataTransfer.types.includes(SYMBOLS_DRAG)) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    setDragOver(true);
  }

  function onDrop(e) {
    if (!dropTarget) return;
    e.preventDefault();
    setDragOver(false);
    const raw = e.dataTransfer.getData(SYMBOLS_DRAG);
    if (raw) onDropSymbols(JSON.parse(raw), folderPath);
  }

  return (
    <li className="nav-branch">
      <div
        className={`nav-item nav-indent-${depth}${isActive ? " active" : ""}${showPendingRename ? " nav-item-editing" : ""}${dragOver ? " nav-item-dragover" : ""}`}
        role={to ? "link" : "button"}
        onClick={showPendingRename ? undefined : onRowClick}
        onDoubleClick={node.editRoute ? () => navigate(`/${panel}${node.editRoute}`) : undefined}
        onContextMenu={onRowContextMenu}
        onDragOver={onDragOver}
        onDragLeave={() => setDragOver(false)}
        onDrop={onDrop}
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
        {showPendingRename ? (
          <>
            <Icon id={navIcon(node.key)} title={node.label} />
            <input
              className="nav-folder-input"
              value={pendingFolder.draft}
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
                if (!draft || draft === node.label) {
                  onCancelFolder();
                  return;
                }
                onCommitFolder({ ...pendingFolder, draft }, true);
              }}
            />
          </>
        ) : (
          <>
            <Icon id={navIcon(node.key)} title={node.label} />
            <span className={`label${node.offline ? " nav-feed-disabled" : ""}`}>{node.label}</span>
            {node.count != null && <span className="count">({node.count})</span>}
          </>
        )}
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
              onDropSymbols={onDropSymbols}
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

function DeleteGroupFolderConfirm({ confirmDelete, onConfirm, onClose }) {
  const close = useDialogStack(onClose);
  return (
    <div className="config-window sym-confirm">
      <div className="config-title">
        <span className="config-title-text">Delete</span>
      </div>
      <div className="config-body">
        <p>
          {confirmDelete.count === 1
            ? `Delete group '${confirmDelete.path}'?`
            : `Delete section '${confirmDelete.path}' and all ${confirmDelete.count} group(s) beneath it?`}
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
  const { datafeeds, reload: reloadDatafeeds } = useDatafeeds();
  const { groups, reload: reloadGroups } = useGroups();
  const [menu, setMenu] = useState(null);
  const [pendingFolder, setPendingFolder] = useState(null);
  const [confirmDeleteSymbol, setConfirmDeleteSymbol] = useState(null);
  const [confirmDeleteGroup, setConfirmDeleteGroup] = useState(null);

  const canEditSymbols = session.can?.right_cfg_symbols !== false;
  const canEditGroups = session.can?.right_cfg_groups !== false;
  const canEditDatafeeds = session.can?.right_cfg_datafeeds !== false;
  const emptyFolders = useMemo(() => emptyFolderPathsToMap(folders), [folders]);

  const refreshTree = useCallback(async () => {
    await Promise.all([reloadSymbols(), reloadFolders(), session.reload?.()]);
  }, [reloadSymbols, reloadFolders, session]);

  const refreshGroupsTree = useCallback(async () => {
    await Promise.all([reloadGroups(), session.reload?.()]);
  }, [reloadGroups, session]);

  const openGroupFolder = useCallback(
    (node) => {
      const path = groupsNavFolderPath(node);
      if (path === null) return;
      setMenu(null);
      navigate(`/${panel}/groups?folder=${encodeURIComponent(path)}`);
    },
    [navigate, panel],
  );

  const startAddGroup = useCallback(
    (node) => {
      const path = groupsNavFolderPath(node);
      if (path === null) return;
      setMenu(null);
      const qs = new URLSearchParams();
      if (path) qs.set("folder", path);
      qs.set("create", "1");
      navigate(`/${panel}/groups?${qs.toString()}`);
    },
    [navigate, panel],
  );

  const requestDeleteGroupFolder = useCallback(
    (node) => {
      const path = groupsNavFolderPath(node);
      if (!path || !isGroupFolderNode(node)) return;
      setMenu(null);
      const targets = groupsUnderFolder(groups, path);
      if (!targets.length) return;
      setConfirmDeleteGroup({ path, count: targets.length, targets });
    },
    [groups],
  );

  const doDeleteGroupFolder = useCallback(
    async ({ targets, path }) => {
      for (const row of targets) {
        const res = await deleteGroup(row.group_id);
        if (!res.ok) {
          window.alert(res.message || "delete failed");
          return;
        }
      }

      const prefix = `folder=${encodeURIComponent(path)}`;
      if (location.search.includes(prefix)) {
        navigate(`/${panel}/groups`, { replace: true });
      }
      await refreshGroupsTree();
    },
    [location.search, navigate, panel, refreshGroupsTree],
  );

  const startEditFolder = useCallback((node) => {
    const path = symbolsNavFolderPath(node);
    if (!path || !isSymbolFolderNode(node)) return;
    setPendingFolder({ mode: "rename", fromPath: path, draft: node.label });
    setMenu(null);
  }, []);

  const startAddFolder = useCallback((node) => {
    const parentPath = symbolsNavFolderPath(node);
    if (parentPath === null) return;
    setPendingFolder({ mode: "add", parentPath, draft: "" });
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

      if (pending.mode === "rename") {
        const from = pending.fromPath;
        const parts = from.split("\\");
        parts[parts.length - 1] = name;
        const to = parts.join("\\");
        if (from === to) return;

        const res = await renameSymbolFolder(from, to);
        if (!res.ok) {
          window.alert(res.message || "rename folder failed");
          return;
        }

        const prefix = `folder=${encodeURIComponent(from)}`;
        if (location.search.includes(prefix)) {
          navigate(`/${panel}/symbols?folder=${encodeURIComponent(to)}`, { replace: true });
        }
      } else {
        const res = await createSymbolFolder(pending.parentPath, name);
        if (!res.ok) {
          window.alert(res.message || "create folder failed");
          return;
        }
      }

      await refreshTree();
    },
    [location.search, navigate, panel, refreshTree],
  );

  const moveSymbols = useCallback(async (ids, folder) => {
    for (const id of ids) {
      const res = await updateSymbol(id, { path: folder });
      if (!res.ok) window.alert(res.message || "move symbol failed");
    }
    refreshTree();
  }, [refreshTree]);

  const requestDeleteSymbolFolder = useCallback((node) => {
    const path = symbolsNavFolderPath(node);
    if (!path || !isSymbolFolderNode(node)) return;
    setMenu(null);
    const count = node.count ?? 0;
    setConfirmDeleteSymbol({ path, count, cascade: count > 0 });
  }, []);

  const doDeleteSymbolFolder = useCallback(async ({ path, cascade }) => {
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
      if (isDatafeedsNavNode(node)) {
        const isFeed = node.key !== "datafeeds";
        setMenu({
          x: e.clientX,
          y: e.clientY,
          node,
          items: datafeedsNavMenuItems({
            canEdit: canEditDatafeeds,
            isFeed,
            onAdd: () => navigate(`/${panel}/datafeeds?add=1`),
            onEdit: isFeed ? () => navigate(`/${panel}${node.editRoute}`) : undefined,
            onRefresh: () => Promise.all([reloadDatafeeds(), session.reload?.()]),
          }),
        });
        return;
      }

      if (isGroupsNavNode(node)) {
        const isFolder = isGroupFolderNode(node);
        setMenu({
          x: e.clientX,
          y: e.clientY,
          node,
          items: groupsNavMenuItems({
            canEdit: canEditGroups,
            isFolder,
            onAdd: () => startAddGroup(node),
            onEdit: isFolder ? () => openGroupFolder(node) : undefined,
            onDelete: isFolder ? () => requestDeleteGroupFolder(node) : undefined,
            onRefresh: refreshGroupsTree,
          }),
        });
        return;
      }

      const isFolder = isSymbolFolderNode(node);
      setMenu({
        x: e.clientX,
        y: e.clientY,
        node,
        items: symbolsNavMenuItems({
          canEdit: canEditSymbols,
          isFolder,
          onAdd: () => startAddFolder(node),
          onEdit: isFolder ? () => startEditFolder(node) : undefined,
          onDelete: isFolder ? () => requestDeleteSymbolFolder(node) : undefined,
          onRefresh: refreshTree,
        }),
      });
    },
    [
      canEditDatafeeds,
      canEditGroups,
      canEditSymbols,
      navigate,
      openGroupFolder,
      panel,
      reloadDatafeeds,
      session,
      refreshGroupsTree,
      refreshTree,
      requestDeleteGroupFolder,
      requestDeleteSymbolFolder,
      startAddFolder,
      startAddGroup,
      startEditFolder,
    ],
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
            onDropSymbols={canEditSymbols ? moveSymbols : undefined}
          />
        ))}
      </ul>
      {menu && (
        <ContextMenu x={menu.x} y={menu.y} items={menu.items} onClose={() => setMenu(null)} />
      )}
      {confirmDeleteSymbol && (
        <DialogOverlay>
          <DeleteFolderConfirm
            confirmDelete={confirmDeleteSymbol}
            onConfirm={doDeleteSymbolFolder}
            onClose={() => setConfirmDeleteSymbol(null)}
          />
        </DialogOverlay>
      )}
      {confirmDeleteGroup && (
        <DialogOverlay>
          <DeleteGroupFolderConfirm
            confirmDelete={confirmDeleteGroup}
            onConfirm={doDeleteGroupFolder}
            onClose={() => setConfirmDeleteGroup(null)}
          />
        </DialogOverlay>
      )}
    </aside>
  );
}

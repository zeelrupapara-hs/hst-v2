import { useState } from "react";
import { NavLink, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { NavIcon } from "../ui/Icon.jsx";
import { ContextMenu, useContextMenu } from "../ui/ContextMenu.jsx";
import { Mt5InputDialog } from "../ui/Mt5InputDialog.jsx";
import { resolveNavIcon } from "../../lib/icons.js";
import { symbolNavLink } from "../../lib/symbolNavTree.js";
import { groupNavLink } from "../../lib/groupNavTree.js";
import { datafeedNavLink } from "../../lib/datafeedNavTree.js";
import { buildSymbolFolderMenu } from "../../lib/symbolContextMenu.js";
import { createSymbolGroup } from "../../lib/data.js";

function isNodeActive(node, activeModule, currentFolder, activeRecordId) {
  if (node.moduleId !== activeModule) return false;
  if (node.datafeedId !== undefined) {
    return String(node.datafeedId) === String(activeRecordId);
  }
  if (node.folderPath !== undefined) {
    return currentFolder === node.folderPath;
  }
  if (node.groupPath !== undefined) {
    return currentFolder === node.groupPath;
  }
  if (activeRecordId) return false;
  return true;
}

function nodeLink(panel, node) {
  if (node.datafeedId !== undefined) {
    return datafeedNavLink(panel, node.datafeedId);
  }
  if (node.folderPath !== undefined) {
    return symbolNavLink(panel, node.folderPath);
  }
  if (node.groupPath !== undefined) {
    return groupNavLink(panel, node.groupPath);
  }
  if (node.moduleId) {
    return `/${panel}/${node.moduleId}`;
  }
  return null;
}

function navClass(node, isActive) {
  const disabledFeed = node.datafeedId !== undefined && !node.enable;
  return `nav-item nav-indent-${node.indent}${isActive ? " active" : ""}${
    disabledFeed ? " nav-feed-disabled" : ""
  }`;
}

function NavItemContent({ node, open, onToggleOpen }) {
  const hasChildren = node.children?.length > 0;
  return (
    <>
      <span
        className={`chevron${hasChildren ? "" : " empty"}`}
        onClick={hasChildren ? onToggleOpen : undefined}
      >
        {hasChildren ? (open ? "▼" : "▶") : "▶"}
      </span>
      <NavIcon name={resolveNavIcon(node)} title={node.label} />
      <span className="label">{node.label}</span>
      {node.count && <span className="count">{node.count}</span>}
    </>
  );
}

function NavItem({ panel, node, onSymbolFolderContextMenu }) {
  const navigate = useNavigate();
  const { moduleId: activeModule, recordId: activeRecordId } = useParams();
  const [searchParams] = useSearchParams();
  const currentFolder = searchParams.get("folder") || "";
  const [open, setOpen] = useState(node.open !== false);
  const hasChildren = node.children?.length > 0;
  const isNavigable = !!node.moduleId;
  const isActive =
    isNavigable &&
    isNodeActive(node, activeModule, currentFolder, activeRecordId);
  const linkTo = nodeLink(panel, node);
  const isSymbolFolder = node.folderPath !== undefined;
  const isGroupFolder = node.groupPath !== undefined;
  // both are addressed by a query string, which a NavLink does not look at when it decides
  // whether it is active: left to it, selecting one folder would select the whole tree
  const isFolderRow = isSymbolFolder || isGroupFolder;
  const linkEnd = hasChildren && node.moduleId === "datafeeds";

  function toggleOpen(e) {
    e.preventDefault();
    e.stopPropagation();
    setOpen(!open);
  }

  function goTo(e) {
    if (isFolderRow && linkTo) {
      e.preventDefault();
      navigate(linkTo);
    }
  }

  function onFolderContextMenu(e) {
    if (!isSymbolFolder || !onSymbolFolderContextMenu) return;
    e.preventDefault();
    e.stopPropagation();
    onSymbolFolderContextMenu(e, node);
  }

  const folderCtx = isSymbolFolder
    ? { onContextMenu: onFolderContextMenu }
    : {};

  if (hasChildren && !isNavigable) {
    return (
      <li className="nav-branch">
        <div className={navClass(node, false)} onClick={() => setOpen(!open)}>
          <NavItemContent node={node} open={open} />
        </div>
        {open && (
          <ul>
            {node.children.map((child, i) => (
              <NavItem
                key={
                  child.datafeedId ??
                  child.groupPath ??
                  child.folderPath ??
                  child.moduleId ??
                  child.label ??
                  i
                }
                panel={panel}
                node={child}
                onSymbolFolderContextMenu={onSymbolFolderContextMenu}
              />
            ))}
          </ul>
        )}
      </li>
    );
  }

  if (hasChildren && isNavigable) {
    const RowTag = isFolderRow ? "div" : NavLink;
    const rowProps = isFolderRow
      ? {
          className: navClass(node, isActive),
          onClick: goTo,
          role: "button",
          ...folderCtx,
        }
      : {
          to: linkTo,
          end: linkEnd,
          className: ({ isActive: on }) => navClass(node, on),
        };

    return (
      <li className="nav-branch">
        <RowTag {...rowProps}>
          <NavItemContent
            node={node}
            open={open}
            onToggleOpen={toggleOpen}
          />
        </RowTag>
        {open && (
          <ul>
            {node.children.map((child, i) => (
              <NavItem
                key={
                  child.datafeedId ??
                  child.groupPath ??
                  child.folderPath ??
                  child.moduleId ??
                  child.label ??
                  i
                }
                panel={panel}
                node={child}
                onSymbolFolderContextMenu={onSymbolFolderContextMenu}
              />
            ))}
          </ul>
        )}
      </li>
    );
  }

  if (isNavigable) {
    if (isFolderRow) {
      return (
        <li>
          <div
            className={navClass(node, isActive)}
            onClick={goTo}
            role="button"
            {...folderCtx}
          >
            <NavItemContent node={node} open={open} />
          </div>
        </li>
      );
    }

    return (
      <li>
        <NavLink
          to={linkTo}
          end={linkEnd}
          className={({ isActive: on }) => navClass(node, on)}
          onClick={(e) => e.stopPropagation()}
        >
          <NavItemContent node={node} open={open} />
        </NavLink>
      </li>
    );
  }

  return (
    <li>
      <div
        className={`${navClass(node, false)}${node.disabled ? " disabled" : ""}`}
      >
        <NavItemContent node={node} open={open} />
      </div>
    </li>
  );
}

export function NavTree({ panel, tree, onNavRefresh }) {
  const navigate = useNavigate();
  const { menu, show, close } = useContextMenu();
  const [toast, setToast] = useState(null);
  const [groupDialog, setGroupDialog] = useState(null);
  const [groupName, setGroupName] = useState("");
  const [groupError, setGroupError] = useState("");

  function showToast(msg) {
    setToast(msg);
    setTimeout(() => setToast(null), 2800);
  }

  function openGroupDialog(node) {
    const folderPath = node.folderPath ?? "";
    setGroupName("");
    setGroupError("");
    setGroupDialog({
      parentPath: folderPath,
      parentLabel: node.label,
      title: folderPath ? `Add group under ${node.label}` : "Add symbol group",
      label: folderPath ? "Subfolder name:" : "Group name:",
    });
  }

  function closeGroupDialog() {
    setGroupDialog(null);
    setGroupName("");
    setGroupError("");
  }

  async function submitGroupDialog() {
    if (!groupDialog) return;
    const name = groupName.trim();
    if (!name) {
      setGroupError("Group name is required");
      return;
    }
    if (/[\\\/]/.test(name)) {
      setGroupError("Name cannot contain \\ or /");
      return;
    }
    const res = await createSymbolGroup(groupDialog.parentPath, name);
    if (!res.data) {
      setGroupError("A group with this name already exists here");
      return;
    }
    closeGroupDialog();
    onNavRefresh?.();
    navigate(symbolNavLink(panel, res.data));
    showToast(`Created group "${name}"`);
  }

  function onSymbolFolderContextMenu(e, node) {
    const folderPath = node.folderPath ?? "";
    const items = buildSymbolFolderMenu({
      folderPath,
      folderLabel: node.label,
      toast: showToast,
      handlers: {
        add: () => {
          close();
          openGroupDialog(node);
        },
        edit: () => navigate(symbolNavLink(panel, folderPath)),
        delete: () =>
          showToast("Delete folder — remove all symbols in this group first"),
        sort: () =>
          showToast("Sort Alphabetically — server-side sort not implemented"),
      },
    });
    show(e, items);
  }

  return (
    <aside className="navigator" onClick={close}>
      <div className="nav-header">Navigator</div>
      <ul className="nav-tree">
        {tree.map((node, i) => (
          <NavItem
            key={node.label || node.moduleId || i}
            panel={panel}
            node={node}
            onSymbolFolderContextMenu={onSymbolFolderContextMenu}
          />
        ))}
      </ul>
      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          items={menu.items}
          onClose={close}
        />
      )}
      {groupDialog && (
        <Mt5InputDialog
          title={groupDialog.title}
          label={groupDialog.label}
          value={groupName}
          onChange={(v) => {
            setGroupName(v);
            if (groupError) setGroupError("");
          }}
          onOk={submitGroupDialog}
          onCancel={closeGroupDialog}
          error={groupError}
        />
      )}
      {toast && <div className="ctx-toast show">{toast}</div>}
    </aside>
  );
}

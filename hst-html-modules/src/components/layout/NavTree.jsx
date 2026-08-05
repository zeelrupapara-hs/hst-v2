import { useState } from "react";
import { NavLink, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { NavIcon } from "../ui/Icon.jsx";
import { resolveNavIcon } from "../../lib/icons.js";
import { symbolNavLink } from "../../lib/symbolNavTree.js";

function isNodeActive(node, activeModule, currentFolder) {
  if (node.moduleId !== activeModule) return false;
  if (node.folderPath !== undefined) {
    return currentFolder === node.folderPath;
  }
  return true;
}

function nodeLink(panel, node) {
  if (node.folderPath !== undefined) {
    return symbolNavLink(panel, node.folderPath);
  }
  if (node.moduleId) {
    return `/${panel}/${node.moduleId}`;
  }
  return null;
}

function navClass(node, isActive) {
  return `nav-item nav-indent-${node.indent}${isActive ? " active" : ""}`;
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

function NavItem({ panel, node }) {
  const navigate = useNavigate();
  const { moduleId: activeModule } = useParams();
  const [searchParams] = useSearchParams();
  const currentFolder = searchParams.get("folder") || "";
  const [open, setOpen] = useState(node.open !== false);
  const hasChildren = node.children?.length > 0;
  const isNavigable = !!node.moduleId;
  const isActive = isNavigable && isNodeActive(node, activeModule, currentFolder);
  const linkTo = nodeLink(panel, node);
  const isSymbolFolder = node.folderPath !== undefined;

  function toggleOpen(e) {
    e.preventDefault();
    e.stopPropagation();
    setOpen(!open);
  }

  function goTo(e) {
    if (isSymbolFolder && linkTo) {
      e.preventDefault();
      navigate(linkTo);
    }
  }

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
                key={child.moduleId || child.folderPath || child.label || i}
                panel={panel}
                node={child}
              />
            ))}
          </ul>
        )}
      </li>
    );
  }

  if (hasChildren && isNavigable) {
    const RowTag = isSymbolFolder ? "div" : NavLink;
    const rowProps = isSymbolFolder
      ? { className: navClass(node, isActive), onClick: goTo, role: "button" }
      : {
          to: linkTo,
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
                key={child.moduleId || child.folderPath || child.label || i}
                panel={panel}
                node={child}
              />
            ))}
          </ul>
        )}
      </li>
    );
  }

  if (isNavigable) {
    if (isSymbolFolder) {
      return (
        <li>
          <div
            className={navClass(node, isActive)}
            onClick={goTo}
            role="button"
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

export function NavTree({ panel, tree }) {
  return (
    <aside className="navigator">
      <div className="nav-header">Navigator</div>
      <ul className="nav-tree">
        {tree.map((node, i) => (
          <NavItem key={node.label || node.moduleId || i} panel={panel} node={node} />
        ))}
      </ul>
    </aside>
  );
}

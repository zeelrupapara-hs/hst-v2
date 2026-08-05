import { useState } from "react";
import { NavLink, useParams } from "react-router-dom";
import { Icon } from "../ui/Icon.jsx";

function NavItem({ panel, node }) {
  const { moduleId: activeModule } = useParams();
  const [open, setOpen] = useState(node.open !== false);
  const hasChildren = node.children?.length;
  const isBranch = hasChildren && !node.moduleId;

  if (isBranch) {
    return (
      <li className="nav-branch">
        <div
          className={`nav-item nav-indent-${node.indent}`}
          onClick={() => setOpen(!open)}
        >
          <span className="chevron">{open ? "▼" : "▶"}</span>
          <span className="ico">
            <Icon name={node.icon} />
          </span>
          <span className="label">{node.label}</span>
          {node.count && <span className="count">{node.count}</span>}
        </div>
        {open && (
          <ul>
            {node.children.map((child, i) => (
              <NavItem key={child.moduleId || child.label || i} panel={panel} node={child} />
            ))}
          </ul>
        )}
      </li>
    );
  }

  if (node.moduleId) {
    const active = activeModule === node.moduleId;
    return (
      <li>
        <NavLink
          to={`/${panel}/${node.moduleId}`}
          className={`nav-item nav-indent-${node.indent}${active ? " active" : ""}`}
          onClick={(e) => e.stopPropagation()}
        >
          <span className="chevron empty">▶</span>
          <span className="ico">
            <Icon name={node.icon} />
          </span>
          <span className="label">{node.label}</span>
          {node.count && <span className="count">{node.count}</span>}
        </NavLink>
      </li>
    );
  }

  return (
    <li>
      <div className={`nav-item nav-indent-${node.indent}${node.disabled ? " disabled" : ""}`}>
        <span className="chevron empty">▶</span>
        <span className="ico">
          <Icon name={node.icon} />
        </span>
        <span className="label">{node.label}</span>
        {node.count && <span className="count">{node.count}</span>}
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
          <NavItem key={node.label || i} panel={panel} node={node} />
        ))}
      </ul>
    </aside>
  );
}

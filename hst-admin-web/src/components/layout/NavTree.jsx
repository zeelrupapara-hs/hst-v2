import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Icon } from "@/components/ui/Icon.jsx";
import { navIcon } from "@/lib/icons.js";
import { useSymbols } from "@/hooks/useSymbols.js";
import { useDatafeeds } from "@/hooks/useDatafeeds.js";
import { useGroups } from "@/hooks/useGroups.js";
import { buildGroupTree, sortedGroupKeys } from "@/lib/groupTree.js";

/** Group folders from the paths; counts are groups beneath, the reference tree rule. */
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
import { annotateTreeCounts, buildFolderTree, sortedTreeChildKeys } from "@/lib/symbolTree.js";

/** Symbol folder tree hangs under the Symbols node; a folder link narrows the list. */
function symbolFolderChildren(folderNode) {
  return sortedTreeChildKeys(folderNode).map((key) => {
    const child = folderNode.children[key];
    const node = {
      key: `symfolder:${child.path}`,
      label: child.name,
      route: `/symbols?folder=${encodeURIComponent(child.path)}`,
      count: child.count || undefined,
    };
    if (sortedTreeChildKeys(child).length) node.children = symbolFolderChildren(child);
    return node;
  });
}

/**
 * The navigator, built from GET /api/v1/navigation. The reference nests everything under
 * Servers > Trade Server and groups the feed modules under Integrations; the backend does
 * not know that spine yet, so it is applied here (prompt §13-A).
 */
function shapeTree(nav, symbols = [], datafeeds = [], groups = []) {
  const nodes = (nav?.nodes ?? []).map((n) => {
    if (n.key === "groups" && groups.length) {
      const root = buildGroupTree(groups);
      return { ...n, route: "/groups", count: root.count, children: groupFolderChildren(root) };
    }
    if (n.key === "symbols") {
      const root = annotateTreeCounts(buildFolderTree(symbols), symbols);
      return { ...n, count: root.count, children: symbolFolderChildren(root) };
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

function NavNode({ node, panel, depth }) {
  const navigate = useNavigate();
  const location = useLocation();
  const [open, setOpen] = useState(depth < 3);

  const hasChildren = node.children?.length > 0;
  const to = node.route ? `/${panel}${node.route}` : null;
  const isActive = to !== null && location.pathname + location.search === to;

  function onRowClick() {
    if (to) navigate(to);
    else if (hasChildren) setOpen(!open);
  }

  return (
    <li className="nav-branch">
      <div
        className={`nav-item nav-indent-${depth}${isActive ? " active" : ""}`}
        role={to ? "link" : "button"}
        onClick={onRowClick}
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
          {node.children.map((child) => (
            <NavNode key={child.key} node={child} panel={panel} depth={depth + 1} />
          ))}
        </ul>
      )}
    </li>
  );
}

export function NavTree({ nav, panel }) {
  const { symbols } = useSymbols();
  const { datafeeds } = useDatafeeds();
  const { groups } = useGroups();

  return (
    <aside className="navigator">
      <div className="nav-header">Navigator</div>
      <ul className="nav-tree">
        {shapeTree(nav, symbols, datafeeds, groups).map((node) => (
          <NavNode key={node.key} node={node} panel={panel} depth={0} />
        ))}
      </ul>
    </aside>
  );
}

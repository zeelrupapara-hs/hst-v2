import { Link, Outlet } from "react-router-dom";
import { ToolbarButton } from "../ui/ToolbarButton.jsx";
import { FullscreenButton } from "./FullscreenButton.jsx";
import { NavTree } from "./NavTree.jsx";
import { StatusBar } from "./StatusBar.jsx";
import { Toolbox } from "./Toolbox.jsx";

export function TerminalShell({
  panel,
  title,
  nav,
  refreshNav,
  onNavRefresh,
  moduleLabel,
  demo,
}) {
  return (
    <>
      <div className="terminal-title">
        <span>{title}</span>
        <div className="title-actions">
          <Link className="panel-switch" to="/login">
            Connect API
          </Link>
          <Link className="panel-switch" to="/">
            Switch panel
          </Link>
          <FullscreenButton />
        </div>
      </div>

      <nav className="menu-bar" aria-label="Main menu">
        <span>File</span>
        <span>Edit</span>
        <span>View</span>
        <span>Services</span>
        <span>Help</span>
      </nav>

      <div className="toolbar" aria-label="Standard toolbar">
        <ToolbarButton icon="add" title="Add">
          Add
        </ToolbarButton>
        <ToolbarButton icon="edit" title="Edit">
          Edit
        </ToolbarButton>
        <ToolbarButton icon="delete" title="Delete">
          Delete
        </ToolbarButton>
        {panel === "admin" && (
          <>
            <span className="toolbar-sep" />
            <button type="button" className="tb-btn" title="Move Up">
              ↑
            </button>
            <button type="button" className="tb-btn" title="Move Down">
              ↓
            </button>
            <span className="toolbar-sep" />
            <button type="button" className="tb-btn" title="Apply Changes">
              Apply
            </button>
          </>
        )}
        <ToolbarButton icon="refresh" title="Refresh">
          Refresh
        </ToolbarButton>
        <div className="toolbar-search">
          <label>Search</label>
          <input
            type="text"
            placeholder={
              panel === "admin" ? "Find configuration…" : "Find account, order…"
            }
            disabled
          />
        </div>
      </div>

      <div className="terminal-body">
        <div className="terminal-main">
          <NavTree panel={panel} tree={nav} onNavRefresh={onNavRefresh} />
          <main className="workspace module-workspace">
            <Outlet context={{ refreshNav: refreshNav || onNavRefresh }} />
          </main>
        </div>
        <Toolbox panel={panel} />
        <StatusBar demo={demo} moduleLabel={moduleLabel} />
      </div>
    </>
  );
}

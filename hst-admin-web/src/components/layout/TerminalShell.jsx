import { Outlet } from "react-router-dom";
import { NavTree } from "@/components/layout/NavTree.jsx";
import { StatusBar } from "@/components/layout/StatusBar.jsx";
import { Toolbox } from "@/components/layout/Toolbox.jsx";
import { ToolbarButton } from "@/components/ui/ToolbarButton.jsx";
import { useFullscreen } from "@/hooks/useFullscreen.js";
import { useSession } from "@/app/session.jsx";

const MENUS = ["File", "Edit", "View", "Services", "Help"];

export function TerminalShell({ panel }) {
  const session = useSession();
  const { fullscreen, toggle } = useFullscreen();

  const title =
    panel === "admin"
      ? "HST Administrator — Trade Server (Live)"
      : "HST Manager — Trade Server (Live)";

  return (
    <>
      <div className="terminal-title">
        <span>{title}</span>
        <div className="title-actions">
          <button type="button" className="panel-switch" onClick={session.logout}>
            Disconnect
          </button>
          <button
            type="button"
            className="title-btn"
            title={fullscreen ? "Exit Full Screen" : "Full Screen"}
            onClick={toggle}
          >
            {fullscreen ? "❐" : "□"}
          </button>
        </div>
      </div>

      <nav className="menu-bar" aria-label="Main menu">
        {MENUS.map((m) => (
          <span key={m}>{m}</span>
        ))}
      </nav>

      <div className="toolbar" aria-label="Standard toolbar">
        <ToolbarButton icon="add" title="Add">Add</ToolbarButton>
        <ToolbarButton icon="edit" title="Edit">Edit</ToolbarButton>
        <ToolbarButton icon="delete" title="Delete">Delete</ToolbarButton>
        <span className="toolbar-sep" />
        <button type="button" className="tb-btn" title="Move Up">↑</button>
        <button type="button" className="tb-btn" title="Move Down">↓</button>
        <span className="toolbar-sep" />
        <button type="button" className="tb-btn" title="Apply Changes">Apply</button>
        <ToolbarButton icon="refresh" title="Refresh" onClick={session.reload}>
          Refresh
        </ToolbarButton>
        <div className="toolbar-search">
          <label>Search</label>
          <input type="text" placeholder="Find configuration…" disabled />
        </div>
      </div>

      <div className="terminal-body">
        <div className="terminal-main">
          <NavTree nav={session.nav} panel={panel} />
          <main className="workspace module-workspace">
            <Outlet />
          </main>
        </div>
        <Toolbox />
        <StatusBar connected={session.status === "ready"} moduleLabel="" />
      </div>
    </>
  );
}

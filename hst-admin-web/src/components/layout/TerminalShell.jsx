import { Outlet } from "react-router-dom";
import { NavTree } from "@/components/layout/NavTree.jsx";
import { StatusBar } from "@/components/layout/StatusBar.jsx";
import { Toolbox } from "@/components/layout/Toolbox.jsx";
import { ToolbarButton } from "@/components/ui/ToolbarButton.jsx";
import { ToolboxProvider } from "@/hooks/useToolbox.jsx";
import { ToolbarActionsProvider } from "@/hooks/useToolbarActions.jsx";
import { useFullscreen } from "@/hooks/useFullscreen.js";
import { useSession } from "@/hooks/useSession.js";

function StandardToolbar({ session }) {
  return (
    <div className="toolbar" aria-label="Standard toolbar">
      <button type="button" className="tb-btn tb-disconnect" onClick={session.logout}>
        <span className="tb-x">✕</span> Disconnect
      </button>
      <ToolbarButton icon="refresh" title="Refresh" onClick={session.reload}>
        Refresh
      </ToolbarButton>
      <div className="toolbar-search">
        <label>Search</label>
        <input type="text" placeholder="Find configuration…" disabled />
      </div>
    </div>
  );
}

/** The window chrome: light title bar, menu bar, the standard toolbar, then the split body. */
export function TerminalShell({ panel }) {
  const session = useSession();
  const { fullscreen, toggle } = useFullscreen();

  const title =
    panel === "admin"
      ? "HST Administrator — Trade Server (Live)"
      : "HST Manager — Trade Server (Live)";

  return (
    <ToolboxProvider>
      <ToolbarActionsProvider>
        <div className="terminal-title">
          <span className="terminal-title-text">{title}</span>
          <div className="title-actions">
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

        <StandardToolbar session={session} />

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
      </ToolbarActionsProvider>
    </ToolboxProvider>
  );
}

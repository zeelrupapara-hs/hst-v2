import { Outlet } from "react-router-dom";
import { NavTree } from "@/components/layout/NavTree.jsx";
import { StatusBar } from "@/components/layout/StatusBar.jsx";
import { Toolbox } from "@/components/layout/Toolbox.jsx";
import { ToolbarButton } from "@/components/ui/ToolbarButton.jsx";
import { ToolboxProvider } from "@/hooks/useToolbox.jsx";
import { ToolbarActionsProvider } from "@/hooks/useToolbarActions.jsx";
import { useFullscreen } from "@/hooks/useFullscreen.js";
import { useSession } from "@/hooks/useSession.js";
import { switchTerminal } from "@/api/endpoints/auth.js";
import { startSocket, stopSocket } from "@/api/socket.js";

function StandardToolbar({ session, panel }) {
  // the other panel's button shows only when this login holds the right to connect there
  const other =
    panel === "admin"
      ? { type: 33, label: "Manager", icon: "manager", allowed: Boolean(session.can?.right_manager) }
      : { type: 32, label: "Administrator", icon: "administrator", allowed: Boolean(session.can?.right_admin) };

  // the new session supersedes this one, so the socket goes quiet before the swap: a
  // session.revoked for the old token must not read as a sign-out
  async function switchPanel() {
    stopSocket();
    const res = await switchTerminal(other.type);
    if (!res.ok) {
      startSocket();
      window.alert(res.message || "panel switch failed");
      return;
    }
    window.location.assign(panel === "admin" ? "/manager" : "/admin");
  }

  return (
    <div className="toolbar" aria-label="Standard toolbar">
      <button type="button" className="tb-btn tb-disconnect" onClick={session.logout}>
        <span className="tb-x">✕</span> Logout
      </button>
      {other.allowed && (
        <ToolbarButton icon={other.icon} title={`Open the ${other.label} panel`} onClick={switchPanel}>
          Switch to {other.label}
        </ToolbarButton>
      )}
      <div className="toolbar-search">
        <label>Search</label>
        <input type="text" placeholder="Find configuration…" disabled />
      </div>
    </div>
  );
}

/** Corner arrows, outward to go full screen and inward to come back. */
function FullscreenGlyph({ exit }) {
  const d = exit
    ? "M9 3v4a2 2 0 0 1-2 2H3 M15 3v4a2 2 0 0 0 2 2h4 M9 21v-4a2 2 0 0 0-2-2H3 M15 21v-4a2 2 0 0 1 2-2h4"
    : "M3 9V5a2 2 0 0 1 2-2h4 M21 9V5a2 2 0 0 0-2-2h-4 M3 15v4a2 2 0 0 0 2 2h4 M21 15v4a2 2 0 0 1-2 2h-4";
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={d} />
    </svg>
  );
}

/** The window chrome: light title bar, menu bar, the standard toolbar, then the split body. */
export function TerminalShell({ panel }) {
  const session = useSession();
  const { fullscreen, toggle } = useFullscreen();

  const who = session.login
    ? ` — ${session.name ? `${session.name} (${session.login})` : session.login}`
    : "";
  const title = (panel === "admin" ? "HST Administrator" : "HST Manager") + who;

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
              <FullscreenGlyph exit={fullscreen} />
            </button>
          </div>
        </div>

        <StandardToolbar session={session} panel={panel} />

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

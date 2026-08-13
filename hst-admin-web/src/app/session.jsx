import { useCallback, useEffect, useRef, useState } from "react";
import { getToken } from "@/api/client.js";
import { signOut } from "@/api/endpoints/auth.js";
import { fetchNavigation } from "@/api/endpoints/navigation.js";
import { SessionContext } from "@/hooks/useSession.js";
import { onEvent, startSocket, stopSocket } from "@/api/socket.js";
import { applySymbolLiveness, reloadSymbols } from "@/hooks/useSymbols.js";
import { reloadSymbolFolders } from "@/hooks/useSymbolFolders.js";
import { reloadGroups } from "@/hooks/useGroups.js";
import { reloadDatafeeds, applyDatafeedRuntime } from "@/hooks/useDatafeeds.js";

/**
 * Session = the token plus what the navigator answered: who this is, which panel,
 * the nodes, and the can{} rights map every button gates on.
 */
export function SessionProvider({ children }) {
  const [state, setState] = useState({ status: "loading", nav: null });
  const debounce = useRef(0);

  const loadNav = useCallback(async () => {
    if (!getToken()) {
      setState({ status: "signed_out", nav: null });
      return;
    }

    const res = await fetchNavigation();
    setState(
      res.ok
        ? { status: "ready", nav: res.data }
        : { status: "signed_out", nav: null }
    );
  }, []);

  // config events arrive in bursts; one coalesced refetch recomputes every count
  const refreshNav = useCallback(() => {
    clearTimeout(debounce.current);
    debounce.current = setTimeout(loadNav, 2000);
  }, [loadNav]);

  const logout = useCallback(async () => {
    stopSocket();
    await signOut();
    setState({ status: "signed_out", nav: null });
  }, []);

  useEffect(() => {
    loadNav();
    return () => clearTimeout(debounce.current);
  }, [loadNav]);

  // every admin action lands in the journal, so journal_create is the config-change signal
  useEffect(() => {
    if (state.status !== "ready") {
      stopSocket();
      return;
    }
    startSocket();
    const offConfig = onEvent("*", (event) => {
      if (event.type === "datafeed_status_updated" || event.type === "datafeed_stats_updated") return;
      if (event.type === "symbol_liveness_updated") return;
      if (!/_created$|_updated$|_deleted$/.test(event.type)) return;
      if (event.type.startsWith("symbol")) {
        reloadSymbols();
        reloadSymbolFolders();
      } else if (event.type.startsWith("group")) reloadGroups();
      else if (event.type.startsWith("datafeed")) reloadDatafeeds();
      refreshNav();
    });
    const offStatus = onEvent("datafeed_status_updated", (event) => {
      applyDatafeedRuntime(event.payload);
    });
    const offStats = onEvent("datafeed_stats_updated", (event) => {
      applyDatafeedRuntime(event.payload);
    });
    const offLiveness = onEvent("symbol_liveness_updated", (event) => {
      applySymbolLiveness(event.payload?.live);
    });
    const offBalance = onEvent("balance_create", refreshNav);
    const offRevoked = onEvent("session.revoked", logout);
    // rights were rewritten in place: reload so every panel respects the new access
    const offRefreshed = onEvent("session_refreshed", () => window.location.reload());
    return () => {
      offConfig();
      offStatus();
      offStats();
      offLiveness();
      offBalance();
      offRevoked();
      offRefreshed();
    };
  }, [state.status, refreshNav, logout]);

  const nav = state.nav;

  return (
    <SessionContext.Provider
      value={{
        status: state.status,
        nav,
        can: nav?.can ?? {},
        login: nav?.login,
        terminal: nav?.terminal,
        reload: loadNav,
        refreshNav,
        logout,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
}

import { useCallback, useEffect, useRef, useState } from "react";
import { getToken } from "@/api/client.js";
import { signOut } from "@/api/endpoints/auth.js";
import { fetchNavigation } from "@/api/endpoints/navigation.js";
import { SessionContext } from "@/hooks/useSession.js";

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
    await signOut();
    setState({ status: "signed_out", nav: null });
  }, []);

  useEffect(() => {
    loadNav();
    return () => clearTimeout(debounce.current);
  }, [loadNav]);

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

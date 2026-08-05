import { useEffect, useState } from "react";
import { adminNav } from "../features/navigation/adminNav.js";
import { managerNav } from "../features/navigation/managerNav.js";
import { fetchList } from "../lib/data.js";
import { cloneNav, injectSymbolFolders } from "../lib/symbolNavTree.js";

/**
 * Admin navigator with symbol folders merged under Symbols (MT5 style).
 * @param {"admin"|"manager"} panel
 * @param {string} activeFolder selected folder path when on symbols module
 */
export function useAdminNavTree(panel, activeFolder = "") {
  const base = panel === "admin" ? adminNav : managerNav;
  const [nav, setNav] = useState(() => cloneNav(base));

  useEffect(() => {
    if (panel !== "admin") {
      setNav(cloneNav(managerNav));
      return;
    }

    let cancelled = false;

    (async () => {
      const res = await fetchList({
        endpoint: "/api/v1/symbols",
        query: { limit: "500" },
      });
      if (cancelled) return;
      setNav(
        injectSymbolFolders(cloneNav(adminNav), res.data || [], activeFolder)
      );
    })();

    return () => {
      cancelled = true;
    };
  }, [panel, activeFolder]);

  return nav;
}

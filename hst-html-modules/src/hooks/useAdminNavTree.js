import { useCallback, useEffect, useState } from "react";
import { adminNav } from "../features/navigation/adminNav.js";
import { managerNav } from "../features/navigation/managerNav.js";
import { fetchList, fetchSymbolGroups } from "../lib/data.js";
import { cloneNav, injectSymbolFolders } from "../lib/symbolNavTree.js";
import { injectDatafeedItems } from "../lib/datafeedNavTree.js";
import { injectGroupFolders } from "../lib/groupNavTree.js";

/**
 * Admin navigator with symbol folders and data feed rows (style).
 * @param {"admin"|"manager"} panel
 * @param {string} activeFolder selected folder path when on symbols module
 * @param {string|number} activeDatafeedId selected feed when on datafeeds module
 */
export function useAdminNavTree(panel, activeFolder = "", activeDatafeedId = "") {
  const base = panel === "admin" ? adminNav : managerNav;
  const [nav, setNav] = useState(() => cloneNav(base));
  const [refreshKey, setRefreshKey] = useState(0);

  const refreshNav = useCallback(() => {
    setRefreshKey((k) => k + 1);
  }, []);

  useEffect(() => {
    if (panel !== "admin") {
      setNav(cloneNav(managerNav));
      return;
    }

    let cancelled = false;

    (async () => {
      const [symRes, groupRes, dfRes, grpRes] = await Promise.all([
        fetchList({
          endpoint: "/api/v1/symbols",
          query: { limit: "500" },
        }),
        fetchSymbolGroups(),
        fetchList({
          endpoint: "/api/v1/datafeeds",
          query: { limit: "500" },
        }),
        fetchList({
          endpoint: "/api/v1/groups",
          query: { flat: "1", limit: "500" },
        }),
      ]);
      if (cancelled) return;
      let tree = injectSymbolFolders(
        cloneNav(adminNav),
        symRes.data || [],
        activeFolder,
        groupRes.data || []
      );
      tree = injectDatafeedItems(tree, dfRes.data || [], activeDatafeedId);
      tree = injectGroupFolders(tree, grpRes.data || [], activeFolder);
      setNav(tree);
    })();

    return () => {
      cancelled = true;
    };
  }, [panel, activeFolder, activeDatafeedId, refreshKey]);

  return { nav, refreshNav };
}

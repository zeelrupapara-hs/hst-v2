/** Build a datafeeds module link for a specific feed record. */
export function datafeedNavLink(panel, datafeedId) {
  return `/${panel}/datafeeds/${encodeURIComponent(datafeedId)}`;
}

/**
 * Inject individual data feed rows under the Data Feeds nav node (Administrator).
 * @param {Array} nav
 * @param {Array<{ datafeed_id?: number, name?: string, enable?: number }>} datafeeds
 * @param {string|number} [activeDatafeedId]
 */
export function injectDatafeedItems(nav, datafeeds, activeDatafeedId = "") {
  const list = datafeeds || [];
  const active = activeDatafeedId ? String(activeDatafeedId) : "";

  function walk(nodes) {
    return nodes.map((node) => {
      if (
        node.moduleId === "datafeeds" &&
        node.datafeedId === undefined &&
        node.folderPath === undefined
      ) {
        const indent = (node.indent ?? 0) + 1;
        const children = list.map((df) => ({
          moduleId: "datafeeds",
          datafeedId: df.datafeed_id,
          label: df.name || `Feed #${df.datafeed_id}`,
          icon: "datafeeds",
          indent,
          enable: df.enable,
        }));
        const onFeed =
          active && children.some((c) => String(c.datafeedId) === active);

        return {
          ...node,
          count: list.length ? `(${list.length})` : undefined,
          open: node.open !== false || onFeed,
          children,
        };
      }
      if (node.children) {
        return { ...node, children: walk(node.children) };
      }
      return node;
    });
  }

  return walk(nav);
}

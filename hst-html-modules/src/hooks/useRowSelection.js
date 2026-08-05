import { useCallback, useMemo, useState } from "react";

export function useRowSelection(rows) {
  const [selected, setSelected] = useState(() => new Set(rows.length ? [0] : []));

  const resetSelection = useCallback((count) => {
    setSelected(count ? new Set([0]) : new Set());
  }, []);

  const toggleSelect = useCallback((index, multi, range) => {
    setSelected((prev) => {
      if (range && prev.size) {
        const anchor = Math.min(...prev);
        const next = new Set();
        const from = Math.min(anchor, index);
        const to = Math.max(anchor, index);
        for (let i = from; i <= to; i++) next.add(i);
        return next;
      }
      if (multi) {
        const next = new Set(prev);
        if (next.has(index)) next.delete(index);
        else next.add(index);
        return next;
      }
      return new Set([index]);
    });
  }, []);

  const selectedRows = useMemo(
    () =>
      [...selected]
        .sort((a, b) => a - b)
        .map((i) => rows[i])
        .filter(Boolean),
    [selected, rows]
  );

  return { selected, toggleSelect, resetSelection, selectedRows };
}

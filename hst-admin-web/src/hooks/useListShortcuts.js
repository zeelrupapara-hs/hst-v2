import { useEffect } from "react";

/**
 * Wire standard list shortcuts when the module root is focused.
 * @param {React.RefObject<HTMLElement>} ref
 * @param {{onAdd?: Function, onEdit?: Function, onDelete?: Function, onFind?: Function}} handlers
 */
export function useListShortcuts(ref, { onAdd, onEdit, onDelete, onFind } = {}) {
  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    function onKeyDown(e) {
      if (!(e.ctrlKey || e.metaKey) && e.key !== "Delete") return;
      const key = e.key.toLowerCase();
      if (key === "n" && onAdd) {
        e.preventDefault();
        onAdd();
      } else if (key === "u" && onEdit) {
        e.preventDefault();
        onEdit();
      } else if (key === "d" && onDelete) {
        e.preventDefault();
        onDelete();
      } else if (key === "f" && onFind) {
        e.preventDefault();
        onFind();
      } else if (e.key === "Delete" && onDelete) {
        e.preventDefault();
        onDelete();
      }
    }

    el.addEventListener("keydown", onKeyDown);
    return () => el.removeEventListener("keydown", onKeyDown);
  }, [ref, onAdd, onEdit, onDelete, onFind]);
}

import { useCallback, useEffect, useId } from "react";
import { useDismissLayer } from "@/hooks/useDialogStack.jsx";

const OPEN_EVENT = "hst:exclusive-dropdown";

/** Tell every other registered dropdown to close. Call when opening this one. */
export function announceExclusiveDropdown(id) {
  document.dispatchEvent(new CustomEvent(OPEN_EVENT, { detail: id }));
}

/**
 * Only one dropdown/list may be open at a time within the admin UI.
 * @param {boolean} open
 * @param {(open: boolean) => void} setOpen
 */
export function useExclusiveDropdown(open, setOpen) {
  const id = useId();
  const close = useCallback(() => setOpen(false), [setOpen]);
  useDismissLayer(open, close);

  useEffect(() => {
    function onOtherOpen(e) {
      if (e.detail !== id) setOpen(false);
    }
    document.addEventListener(OPEN_EVENT, onOtherOpen);
    return () => document.removeEventListener(OPEN_EVENT, onOtherOpen);
  }, [id, setOpen]);

  const announceOpen = useCallback(() => {
    announceExclusiveDropdown(id);
  }, [id]);

  useEffect(() => {
    if (open) announceExclusiveDropdown(id);
  }, [open, id]);

  return announceOpen;
}

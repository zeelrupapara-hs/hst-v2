import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
} from "react";

const DialogStackContext = createContext(null);

/** Tracks open modals/overlays so Escape dismisses the top layer first (LIFO). */
export function DialogStackProvider({ children }) {
  const stackRef = useRef([]);

  const register = useCallback((id, handlers = {}) => {
    stackRef.current = [...stackRef.current, { id, ...handlers }];
    return () => {
      stackRef.current = stackRef.current.filter((entry) => entry.id !== id);
    };
  }, []);

  const requestClose = useCallback((id, fn) => {
    const stack = stackRef.current;
    if (stack.length === 0 || stack[stack.length - 1].id === id) fn();
  }, []);

  const dismissTop = useCallback(() => {
    const stack = stackRef.current;
    if (stack.length === 0) return false;
    const top = stack[stack.length - 1];
    if (top.onEscape?.()) return true;
    top.close?.();
    return true;
  }, []);

  useEffect(() => {
    const onKey = (e) => {
      if (e.key !== "Escape") return;
      if (dismissTop()) {
        e.preventDefault();
        e.stopPropagation();
      }
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [dismissTop]);

  const value = useMemo(() => ({ register, requestClose }), [register, requestClose]);

  return <DialogStackContext.Provider value={value}>{children}</DialogStackContext.Provider>;
}

function useDialogStackContext() {
  return useContext(DialogStackContext);
}

/** Move to the previous tab on Escape; return false when already on the first tab. */
export function prevTabEscape(tabs, activeTab, setActiveTab) {
  const ids = tabs.map((t) => (typeof t === "string" ? t : t.id));
  const idx = ids.indexOf(activeTab);
  if (idx <= 0) return false;
  setActiveTab(ids[idx - 1]);
  return true;
}

/** Register a transient layer (dropdown, context menu) above modals. */
export function useDismissLayer(active, onDismiss) {
  const ctx = useDialogStackContext();
  const id = useId();
  const dismissRef = useRef(onDismiss);
  dismissRef.current = onDismiss;

  useLayoutEffect(() => {
    if (!active || !ctx) return undefined;
    const dismiss = () => dismissRef.current?.();
    return ctx.register(id, {
      close: dismiss,
      onEscape: () => {
        dismiss();
        return true;
      },
    });
  }, [active, ctx, id]);
}

/**
 * Register this modal layer and return a close handler that respects stack order.
 * @param {Function} onClose
 * @param {{ onEscape?: () => boolean }} [options] onEscape returns true when handled (e.g. previous tab)
 */
export function useDialogStack(onClose, { onEscape } = {}) {
  const ctx = useDialogStackContext();
  const id = useId();
  const onCloseRef = useRef(onClose);
  const onEscapeRef = useRef(onEscape);
  onCloseRef.current = onClose;
  onEscapeRef.current = onEscape;

  const close = useCallback(() => {
    if (!onCloseRef.current) return;
    if (!ctx) {
      onCloseRef.current();
      return;
    }
    ctx.requestClose(id, () => onCloseRef.current?.());
  }, [ctx, id]);

  useLayoutEffect(() => {
    if (!ctx) return undefined;
    return ctx.register(id, {
      close: () => {
        if (!onCloseRef.current) return;
        ctx.requestClose(id, () => onCloseRef.current?.());
      },
      onEscape: () => onEscapeRef.current?.() ?? false,
    });
  }, [ctx, id]);

  return close;
}

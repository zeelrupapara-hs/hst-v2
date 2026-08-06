import { createContext, useCallback, useContext, useId, useLayoutEffect, useMemo, useRef } from "react";

const DialogStackContext = createContext(null);

/** Tracks open modals so only the topmost one may close (LIFO). */
export function DialogStackProvider({ children }) {
  const stackRef = useRef([]);

  const register = useCallback((id) => {
    stackRef.current = [...stackRef.current, id];
    return () => {
      stackRef.current = stackRef.current.filter((entry) => entry !== id);
    };
  }, []);

  const requestClose = useCallback((id, fn) => {
    const stack = stackRef.current;
    if (stack.length === 0 || stack[stack.length - 1] === id) fn();
  }, []);

  const value = useMemo(() => ({ register, requestClose }), [register, requestClose]);

  return <DialogStackContext.Provider value={value}>{children}</DialogStackContext.Provider>;
}

function useDialogStackContext() {
  return useContext(DialogStackContext);
}

/** Register this modal layer and return a close handler that respects stack order. */
export function useDialogStack(onClose) {
  const ctx = useDialogStackContext();
  const id = useId();

  useLayoutEffect(() => {
    if (!ctx) return undefined;
    return ctx.register(id);
  }, [ctx, id]);

  const close = useCallback(() => {
    if (!onClose) return;
    if (!ctx) {
      onClose();
      return;
    }
    ctx.requestClose(id, onClose);
  }, [ctx, id, onClose]);

  return close;
}

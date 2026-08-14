import { useCallback, useRef, useState } from "react";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog.jsx";

/**
 * Promise-form ConfirmDialog, so a handler can ask inline the way window.confirm allowed:
 *
 *   const { confirm, confirmElement } = useConfirm();
 *   if (!(await confirm({ title: "Groups", message: "Delete selected configuration record?" }))) return;
 *
 * Render {confirmElement} once next to the module's other dialogs.
 * @returns {{ confirm: (opts: object) => Promise<boolean>, confirmElement: object|null }}
 */
export function useConfirm() {
  const [opts, setOpts] = useState(null);
  const resolver = useRef(null);

  const confirm = useCallback(
    (next) =>
      new Promise((resolve) => {
        resolver.current = resolve;
        setOpts(next);
      }),
    [],
  );

  // the Yes button fires onYes then onClose, so the first answer wins and the close resolves false only when Yes never ran
  const settle = (answer) => {
    resolver.current?.(answer);
    resolver.current = null;
    setOpts(null);
  };

  const confirmElement = opts ? (
    <ConfirmDialog
      prompt=""
      overlayClassName="admin-confirm-overlay"
      {...opts}
      onYes={() => settle(true)}
      onClose={() => settle(false)}
    />
  ) : null;

  return { confirm, confirmElement };
}

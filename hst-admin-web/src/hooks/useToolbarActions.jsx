import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

const empty = {
  onAdd: undefined,
  onEdit: undefined,
  onDelete: undefined,
  canAdd: false,
  canEdit: false,
  canDelete: false,
};

export const ToolbarActionsContext = createContext(null);

/** Provider lives in TerminalShell; modules register handlers while mounted. */
export function ToolbarActionsProvider({ children }) {
  const [actions, setActions] = useState(empty);

  const register = useCallback((next) => {
    setActions({
      onAdd: next.onAdd,
      onEdit: next.onEdit,
      onDelete: next.onDelete,
      canAdd: Boolean(next.canAdd),
      canEdit: Boolean(next.canEdit),
      canDelete: Boolean(next.canDelete),
    });
  }, []);

  const clear = useCallback(() => setActions(empty), []);

  const value = useMemo(() => ({ actions, register, clear }), [actions, register, clear]);

  return (
    <ToolbarActionsContext.Provider value={value}>
      {children}
    </ToolbarActionsContext.Provider>
  );
}

export function useToolbarActions() {
  return useContext(ToolbarActionsContext);
}

/**
 * Register standard toolbar handlers for the active module.
 * @param {{onAdd?: Function, onEdit?: Function, onDelete?: Function, canAdd?: boolean, canEdit?: boolean, canDelete?: boolean}} handlers
 */
export function useRegisterToolbarActions({
  onAdd,
  onEdit,
  onDelete,
  canAdd = false,
  canEdit = false,
  canDelete = false,
}) {
  const ctx = useToolbarActions();

  useEffect(() => {
    if (!ctx?.register || !ctx?.clear) return undefined;
    ctx.register({ onAdd, onEdit, onDelete, canAdd, canEdit, canDelete });
    return () => ctx.clear();
  }, [ctx?.register, ctx?.clear, onAdd, onEdit, onDelete, canAdd, canEdit, canDelete]);
}

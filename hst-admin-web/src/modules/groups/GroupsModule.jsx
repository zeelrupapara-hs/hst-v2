import { useMemo, useRef, useState, useEffect, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import { useGroups } from "@/hooks/useGroups.js";
import { useSession } from "@/hooks/useSession.js";
import { useListShortcuts } from "@/hooks/useListShortcuts.js";
import { useRegisterToolbarActions } from "@/hooks/useToolbarActions.jsx";
import { ContextMenu, listMenuHead, listMenuTail } from "@/components/ui/ContextMenu.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { AuthMode_name, MarginMode_short } from "@/constants/groups.js";
import { filterGroupsByFolder } from "@/lib/groupTree.js";
import { deleteGroup } from "@/api/endpoints/groups.js";
import { GroupDialog } from "./GroupDialog.jsx";

const ENABLE_CONNECTION = 2;

/** Groups list: real groups of the selected folder, the folder tree lives in the navigator. */
export function GroupsModule() {
  const { groups, loading, reload } = useGroups();
  const session = useSession();
  const rootRef = useRef(null);
  const [params, setSearchParams] = useSearchParams();
  const folder = params.get("folder") || "";
  const [selected, setSelected] = useState([]);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const [sorted, setSorted] = useState(false);
  const [grid, setGrid] = useState(true);
  const canEdit = session.can?.right_cfg_groups !== false;

  const rows = useMemo(() => {
    const list = filterGroupsByFolder(groups, folder);
    return sorted ? [...list].sort((a, b) => a.group.localeCompare(b.group)) : list;
  }, [groups, folder, sorted]);

  const openAdd = useCallback(() => setDialog({ id: "new" }), []);

  const openEdit = useCallback(() => {
    const row = rows[selected[0]];
    if (row) setDialog({ id: row.group_id });
  }, [rows, selected]);

  useEffect(() => {
    if (params.get("create") !== "1" || !canEdit) return;
    setDialog({ id: "new" });
    const next = new URLSearchParams(params);
    next.delete("create");
    setSearchParams(next, { replace: true });
  }, [params, canEdit, setSearchParams]);

  const pick = (i, e) => {
    if (e.shiftKey && selected.length) {
      const from = selected[selected.length - 1];
      const [lo, hi] = from < i ? [from, i] : [i, from];
      setSelected(Array.from({ length: hi - lo + 1 }, (_, k) => lo + k));
    } else if (e.ctrlKey || e.metaKey) {
      setSelected(selected.includes(i) ? selected.filter((s) => s !== i) : [...selected, i]);
    } else {
      setSelected([i]);
    }
  };

  const onDelete = useCallback(async () => {
    const targets = selected.map((i) => rows[i]).filter(Boolean);
    if (!targets.length) return;
    const what = targets.length === 1 ? `group '${targets[0].group}'` : `${targets.length} groups`;
    if (!window.confirm(`Delete ${what}?`)) return;
    for (const row of targets) {
      const res = await deleteGroup(row.group_id);
      if (!res.ok) window.alert(res.message || "delete failed");
    }
    saved();
  }, [selected, rows]);

  function saved() {
    reload();
    session.refreshNav?.();
  }

  const hasSelection = selected.length > 0;

  useListShortcuts(rootRef, {
    onAdd: canEdit ? openAdd : undefined,
    onEdit: canEdit && hasSelection ? openEdit : undefined,
    onDelete: canEdit && hasSelection ? onDelete : undefined,
  });

  useRegisterToolbarActions({
    onAdd: canEdit ? openAdd : undefined,
    onEdit: canEdit && hasSelection ? openEdit : undefined,
    onDelete: canEdit && hasSelection ? onDelete : undefined,
    canAdd: canEdit,
    canEdit: canEdit && hasSelection,
    canDelete: canEdit && hasSelection,
  });

  const items = [
    { label: "Servers", disabled: true },
    "sep",
    ...listMenuHead({
      onAdd: openAdd,
      onEdit: openEdit,
      onDelete,
      hasSelection,
    }),
    ...listMenuTail({
      on: {
        sort: () => setSorted(true),
        toggleGrid: () => setGrid((v) => !v),
      },
      view: { grid },
      extras: [
        { label: "Automation triggers", disabled: true },
        { label: "Automation actions", disabled: true },
        { label: "Import from Server", disabled: true },
      ],
    }),
  ];

  return (
    <div className="module-root" ref={rootRef} tabIndex={-1}>
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className={`data-table data-table-auto${grid ? " data-table-grid" : ""}`}>
          <thead>
            <tr>
              <th>Group</th>
              <th>Server</th>
              <th>Company</th>
              <th>Type</th>
              <th>Authentication</th>
              <th>Margin</th>
              <th>Currency</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => {
              const off = ((row.permission_flags ?? 0) & ENABLE_CONNECTION) === 0;
              return (
                <tr
                  key={row.group_id}
                  className={`${selected.includes(i) ? "selected" : ""}${off ? " grp-row-disabled" : ""}`.trim()}
                  onClick={(e) => pick(i, e)}
                  onContextMenu={(e) => !selected.includes(i) && pick(i, e)}
                  onDoubleClick={() => canEdit && setDialog({ id: row.group_id })}
                >
                  <td>
                    <span className="sym-symbol-cell">
                      <Icon id="groups" />
                      {row.group}
                    </span>
                  </td>
                  <td>Main Server</td>
                  <td>{row.company || "—"}</td>
                  <td>{MarginMode_short[row.margin_mode] ?? row.margin_mode}</td>
                  <td>{AuthMode_name[row.auth_mode] ?? row.auth_mode}</td>
                  <td>{`${row.margin_call ?? 0} / ${row.margin_stop_out ?? 0} ${row.margin_so_mode === 1 ? row.currency : "%"}`}</td>
                  <td>{row.currency}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      {menu && canEdit && (
        <ContextMenu x={menu.x} y={menu.y} onClose={() => setMenu(null)} items={items} />
      )}
      {dialog && (
        <GroupDialog
          groupId={dialog.id}
          folderPath={folder}
          onClose={() => setDialog(null)}
          onSaved={saved}
        />
      )}
    </div>
  );
}

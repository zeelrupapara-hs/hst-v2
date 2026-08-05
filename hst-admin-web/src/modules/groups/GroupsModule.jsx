import { useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useGroups } from "@/hooks/useGroups.js";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { AuthMode_name, MarginMode_short } from "@/constants/groups.js";
import { filterGroupsByFolder } from "@/lib/groupTree.js";
import { deleteGroup } from "@/api/endpoints/groups.js";
import { GroupDialog } from "./GroupDialog.jsx";

/** Groups list: real groups of the selected folder, the folder tree lives in the navigator. */
export function GroupsModule() {
  const { groups, loading, reload } = useGroups();
  const session = useSession();
  const [params] = useSearchParams();
  const folder = params.get("folder") || "";
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const canEdit = session.can?.right_cfg_groups !== false;

  const rows = useMemo(
    () => filterGroupsByFolder(groups, folder).sort((a, b) => a.group.localeCompare(b.group)),
    [groups, folder],
  );

  async function onDelete(row) {
    if (!window.confirm(`Delete group '${row.group}'?`)) return;
    const res = await deleteGroup(row.group_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  function saved() {
    reload();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className="data-table data-table-grid data-table-auto">
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
            {rows.map((row, i) => (
              <tr
                key={row.group_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: row.group_id })}
              >
                <td>
                  <span className="sym-symbol-cell">
                    <Icon id="groups" />
                    {row.group}
                  </span>
                </td>
                <td>Trade Server</td>
                <td>{row.company || "—"}</td>
                <td>{MarginMode_short[row.margin_mode] ?? row.margin_mode}</td>
                <td>{AuthMode_name[row.auth_mode] ?? row.auth_mode}</td>
                <td>{`${row.margin_call ?? 0} / ${row.margin_stop_out ?? 0} %`}</td>
                <td>{row.currency}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {menu && canEdit && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          onClose={() => setMenu(null)}
          items={[
            { label: "Add", onClick: () => setDialog({ id: "new" }) },
            { label: "Edit", disabled: selected == null, onClick: () => setDialog({ id: rows[selected]?.group_id }) },
            "sep",
            { label: "Delete", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
          ]}
        />
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

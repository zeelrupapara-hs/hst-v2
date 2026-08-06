import { useEffect, useState } from "react";
import { useSession } from "@/hooks/useSession.js";
import { ContextMenu } from "@/components/ui/ContextMenu.jsx";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createClient,
  deleteClient,
  fetchClient,
  fetchClients,
  updateClient,
} from "@/api/endpoints/users.js";
import { ClientStatus_name, ClientType_name, KycStatus_name } from "@/constants/users.js";
import { formatNs } from "@/lib/time.js";

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

function Field({ label, value, onChange, wide }) {
  return (
    <>
      <label>{label}</label>
      <input
        type="text"
        value={value ?? ""}
        className={wide ? "wide" : ""}
        onChange={(e) => onChange(e.target.value)}
      />
    </>
  );
}

const PATCH_FIELDS = [
  "client_type", "client_status", "kyc_status", "comment", "person_name",
  "person_last_name", "person_middle_name", "company_name", "contact_email",
  "contact_phone", "address_country", "address_city", "lead_campaign", "lead_source",
];

function ClientDialog({ clientId, onClose, onSaved }) {
  const isNew = clientId === "new";
  const [tab, setTab] = useState("General");
  const [draft, setDraft] = useState(isNew ? { client_type: 1, client_status: 0, kyc_status: 0 } : null);
  const [original, setOriginal] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(clientId);
  const close = useDialogStack(onClose);

  useEffect(() => {
    if (isNew) return;
    fetchClient(clientId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load client");
        return;
      }
      setDraft(res.data);
      setOriginal(res.data);
    });
  }, [clientId, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    if (isNew) {
      if (!draft.person_name?.trim() || !draft.contact_email?.trim()) {
        setError("Name and email are required");
        return;
      }
      const res = await createClient(draft);
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
    } else {
      const patch = {};
      for (const key of PATCH_FIELDS) {
        if (draft[key] !== undefined && draft[key] !== original[key]) patch[key] = draft[key];
      }
      if (Object.keys(patch).length) {
        const res = await updateClient(clientId, patch);
        if (!res.ok) {
          setError(res.message || "save failed");
          return;
        }
      }
    }
    onSaved();
    close();
  }

  const TABS = ["General", "Personal", "Address"];

  function panel() {
    if (!draft) return null;
    if (tab === "General") {
      return (
        <div className="form-grid">
          <label>Type</label>
          <PropSelect fill value={draft.client_type} options={enumOptions(ClientType_name)} onChange={(v) => set("client_type", v)} />
          <label>Status</label>
          <PropSelect fill value={draft.client_status} options={enumOptions(ClientStatus_name)} onChange={(v) => set("client_status", v)} />
          <label>KYC status</label>
          <PropSelect fill value={draft.kyc_status} options={enumOptions(KycStatus_name)} onChange={(v) => set("kyc_status", v)} />
          <Field label="Lead source" value={draft.lead_source} onChange={(v) => set("lead_source", v)} />
          <Field label="Lead campaign" value={draft.lead_campaign} onChange={(v) => set("lead_campaign", v)} />
          <Field label="Comment" value={draft.comment} onChange={(v) => set("comment", v)} wide />
        </div>
      );
    }
    if (tab === "Personal") {
      return (
        <div className="form-grid">
          <Field label="First name" value={draft.person_name} onChange={(v) => set("person_name", v)} />
          <Field label="Last name" value={draft.person_last_name} onChange={(v) => set("person_last_name", v)} />
          <Field label="Middle name" value={draft.person_middle_name} onChange={(v) => set("person_middle_name", v)} />
          <Field label="Company" value={draft.company_name} onChange={(v) => set("company_name", v)} />
          <Field label="Email" value={draft.contact_email} onChange={(v) => set("contact_email", v)} wide />
          <Field label="Phone" value={draft.contact_phone} onChange={(v) => set("contact_phone", v)} />
        </div>
      );
    }
    return (
      <div className="form-grid">
        <Field label="Country" value={draft.address_country} onChange={(v) => set("address_country", v)} />
        <Field label="City" value={draft.address_city} onChange={(v) => set("address_city", v)} />
      </div>
    );
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          height={470}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "Client: New" : `Client: ${draft?.person_name ?? clientId}`}
          tabs={
            <div className="config-tabs">
              {TABS.map((t) => (
                <button key={t} type="button" className={tab === t ? "active" : ""} onClick={() => setTab(t)}>
                  {t}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={close}>Cancel</button>
              <button type="button" className="config-help" disabled>Help</button>
            </div>
          }
        >
          <div className="config-panel active">
            <div className="sym-sessions-intro">
              <span className="sym-tab-intro-icon" aria-hidden="true">
                <Icon id="clients-tree" size={48} />
              </span>
              <p>The client record: identity, pipeline status and contact details.</p>
            </div>
            {panel()}
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

/** Clients list with the pipeline status and KYC columns. */
export function ClientsModule() {
  const session = useSession();
  const [rows, setRows] = useState(null);
  const [selected, setSelected] = useState(null);
  const [dialog, setDialog] = useState(null);
  const [menu, setMenu] = useState(null);
  const [view, setView] = useState({ grid: true, autoArrange: true });
  const canEdit = session.can?.right_clients_edit !== false;

  const load = () => fetchClients().then((res) => res.ok && setRows(res.data || []));

  useEffect(() => {
    load();
  }, []);

  async function onDelete(row) {
    if (!window.confirm(`Delete client '${row.person_name}'?`)) return;
    const res = await deleteClient(row.client_id);
    if (!res.ok) window.alert(res.message || "delete failed");
    saved();
  }

  function saved() {
    load();
    session.refreshNav?.();
  }

  return (
    <div className="module-root">
      <div className="table-wrap" onContextMenu={(e) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY }); }}>
        <table className={`data-table${view.grid ? " data-table-grid" : ""}${view.autoArrange ? " data-table-auto" : ""}`}>
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Email</th>
              <th>Phone</th>
              <th>Status</th>
              <th>KYC</th>
              <th>Modified</th>
            </tr>
          </thead>
          <tbody>
            {(rows || []).map((row, i) => (
              <tr
                key={row.client_id}
                className={selected === i ? "selected" : ""}
                onClick={() => setSelected(i)}
                onContextMenu={() => setSelected(i)}
                onDoubleClick={() => canEdit && setDialog({ id: row.client_id })}
              >
                <td>{row.client_id}</td>
                <td>{[row.person_name, row.person_last_name].filter(Boolean).join(" ")}</td>
                <td>{row.contact_email}</td>
                <td>{row.contact_phone}</td>
                <td>{ClientStatus_name[row.client_status] ?? row.client_status}</td>
                <td>{KycStatus_name[row.kyc_status] ?? row.kyc_status}</td>
                <td>{row.date_modified ? formatNs(row.date_modified) : "—"}</td>
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
            { label: "New Client", icon: "add", shortcut: "Ctrl+N", onClick: () => setDialog({ id: "new" }) },
            { label: "Edit", icon: "edit", shortcut: "Ctrl+U", disabled: selected == null, onClick: () => setDialog({ id: rows[selected]?.client_id }) },
            { label: "Delete", icon: "delete", shortcut: "Ctrl+D", disabled: selected == null, onClick: () => onDelete(rows[selected]) },
            "sep",
            { label: "Export", disabled: true },
            { label: "Find", shortcut: "Ctrl+F", disabled: true },
            "sep",
            { label: "Auto Arrange", checked: view.autoArrange, onClick: () => setView((v) => ({ ...v, autoArrange: !v.autoArrange })) },
            { label: "Grid", checked: view.grid, onClick: () => setView((v) => ({ ...v, grid: !v.grid })) },
          ]}
        />
      )}
      {dialog && <ClientDialog clientId={dialog.id} onClose={() => setDialog(null)} onSaved={saved} />}
    </div>
  );
}

import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { useGroups } from "@/hooks/useGroups.js";
import { createUser, fetchUser, updateUser } from "@/api/endpoints/users.js";
import { AccountRight_checks, LimitRight_checks } from "@/constants/users.js";

const TABS = ["Personal", "Account", "Limits", "Security"];

function Intro({ children }) {
  return (
    <div className="sym-sessions-intro">
      <span className="sym-tab-intro-icon" aria-hidden="true">
        <Icon id="accounts-tree" size={48} />
      </span>
      <p>{children}</p>
    </div>
  );
}

function Field({ label, value, onChange, type = "text", wide, readOnly }) {
  return (
    <>
      <label>{label}</label>
      <input
        type={type}
        readOnly={readOnly || !onChange}
        value={value ?? ""}
        className={wide ? "wide" : ""}
        onChange={onChange ? (e) => onChange(e.target.value) : undefined}
      />
    </>
  );
}

function RightCheck({ rights, def, onChange }) {
  const on = ((rights ?? 0) & def.bit) !== 0;
  return (
    <label className="sym-check">
      <input
        type="checkbox"
        checked={def.inverted ? !on : on}
        onChange={() => onChange((rights ?? 0) ^ def.bit)}
      />{" "}
      {def.label}
    </label>
  );
}

const newDraft = (group) => ({
  login: null,
  group: group || "",
  name: "",
  email: "",
  phone: "",
  country: "",
  city: "",
  comment: "",
  leverage: 100,
  rights: 0x0001 | 0x0002 | 0x0020 | 0x0040 | 0x0100,
  password_main: "",
  password_investor: "",
});

/** The trading account dialog; "new" creates via CrtUser, edits diff through UptUser. */
export function AccountDialog({ login, onClose, onSaved }) {
  const isNew = login === "new";
  const { groups } = useGroups();
  const [activeTab, setActiveTab] = useState(isNew ? "Personal" : "Personal");
  const [draft, setDraft] = useState(isNew ? newDraft(groups[0]?.group) : null);
  const [original, setOriginal] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(login);

  useEffect(() => {
    if (isNew) return;
    fetchUser(login).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load account");
        return;
      }
      setDraft(res.data);
      setOriginal(res.data);
    });
  }, [login, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    if (!draft) return;
    if (isNew) {
      if (!draft.name.trim() || !draft.email.trim() || !draft.group) {
        setError("Name, email and group are required");
        return;
      }
      if (draft.password_main.length < 8 || draft.password_investor.length < 8) {
        setError("Passwords need at least 8 characters");
        return;
      }
      const res = await createUser({
        group: draft.group,
        name: draft.name.trim(),
        email: draft.email.trim(),
        phone: draft.phone,
        country: draft.country,
        city: draft.city,
        comment: draft.comment,
        rights: draft.rights,
        password_main: draft.password_main,
        password_investor: draft.password_investor,
      });
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
    } else {
      const patch = {};
      for (const key of ["group", "rights", "name", "email", "phone", "country", "city", "comment", "leverage"]) {
        if (draft[key] !== undefined && draft[key] !== original[key]) patch[key] = draft[key];
      }
      if (Object.keys(patch).length) {
        const res = await updateUser(login, patch);
        if (!res.ok) {
          setError(res.message || "save failed");
          return;
        }
      }
    }
    onSaved();
    onClose();
  }

  const groupOptions = groups.map((g) => ({ value: g.group, label: g.group }));

  function panel(tab) {
    if (!draft) return null;
    switch (tab) {
      case "Personal":
        return (
          <>
            <Intro>
              The account holder's personal details. The name and email identify the account in
              reports and mailings.
            </Intro>
            <div className="form-grid">
              <Field label="Name" value={draft.name} onChange={(v) => set("name", v)} wide />
              <Field label="Email" value={draft.email} onChange={(v) => set("email", v)} wide />
              <Field label="Phone" value={draft.phone} onChange={(v) => set("phone", v)} />
              <Field label="Country" value={draft.country} onChange={(v) => set("country", v)} />
              <Field label="City" value={draft.city} onChange={(v) => set("city", v)} />
              <Field label="Comment" value={draft.comment} onChange={(v) => set("comment", v)} wide />
            </div>
          </>
        );
      case "Account":
        return (
          <>
            <Intro>
              The group defines the account's trading conditions; the leverage and the connection
              rights are set here.
            </Intro>
            <div className="form-grid">
              <label>Group</label>
              <PropSelect fill value={draft.group} options={groupOptions} onChange={(v) => set("group", v)} />
              <Field
                label="Leverage"
                value={`1 : ${draft.leverage ?? 100}`}
                onChange={(v) => set("leverage", Number(v.replace(/[^0-9]/g, "")) || 100)}
              />
            </div>
            <div className="form-grid grp-check-stack">
              {AccountRight_checks.map((def) => (
                <RightCheck key={def.bit} rights={draft.rights} def={def} onChange={(v) => set("rights", v)} />
              ))}
            </div>
          </>
        );
      case "Limits":
        return (
          <>
            <Intro>Trading permissions of the account; the stricter of account and group applies.</Intro>
            <div className="form-grid grp-check-stack">
              {LimitRight_checks.map((def) => (
                <RightCheck key={def.bit} rights={draft.rights} def={def} onChange={(v) => set("rights", v)} />
              ))}
            </div>
          </>
        );
      case "Security":
        return isNew ? (
          <>
            <Intro>Initial passwords for the account. Both need at least 8 characters.</Intro>
            <div className="form-grid">
              <Field label="Master password" type="password" value={draft.password_main} onChange={(v) => set("password_main", v)} />
              <Field label="Investor password" type="password" value={draft.password_investor} onChange={(v) => set("password_investor", v)} />
            </div>
          </>
        ) : (
          <Intro>
            Passwords are changed by the account owner through the trading terminal; a staff-side
            reset endpoint is not available yet.
          </Intro>
        );
    }
  }

  return (
    <div className="dialog-overlay" onClick={onClose} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          draggable
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "Account: New" : `Account: ${login} — ${draft?.name ?? "…"}`}
          tabs={
            <div className="config-tabs">
              {TABS.map((tab) => (
                <button
                  key={tab}
                  type="button"
                  className={activeTab === tab ? "active" : ""}
                  onClick={() => setActiveTab(tab)}
                >
                  {tab}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>OK</button>
              <button type="button" onClick={onClose}>Cancel</button>
            </div>
          }
        >
          {TABS.map((tab) => (
            <div key={tab} className={`config-panel${activeTab === tab ? " active" : ""}`}>
              {activeTab === tab && panel(tab)}
            </div>
          ))}
        </SettingsDialog>
      </div>
    </div>
  );
}

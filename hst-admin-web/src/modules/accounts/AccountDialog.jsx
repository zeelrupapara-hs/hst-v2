import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { Icon } from "@/components/ui/Icon.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { useGroups } from "@/hooks/useGroups.js";
import { createUser, fetchUser, resetUserPassword, updateUser } from "@/api/endpoints/users.js";
import { AccountRight_checks, LimitRight_checks } from "@/constants/users.js";
import { AccountOverviewTab } from "./AccountOverviewTab.jsx";
import { generatePassword } from "@/lib/passwords.js";

const EDIT_TABS = ["Overview", "Personal", "Account", "Limits", "Security"];
const NEW_TABS = ["Personal", "Account", "Limits", "Security"];

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

const PASSWORD_KINDS = [
  { kind: "main", label: "Master password", note: "used for full access to the trading account" },
  { kind: "investor", label: "Investor password", note: "used for read-only access to the trading account" },
  { kind: "api", label: "API password", note: "used for access to the server through the web API" },
];

function PasswordRow({ login, kind }) {
  const [value, setValue] = useState("");
  const [status, setStatus] = useState("");

  async function onChangePassword() {
    if (value.length < 8) {
      setStatus("minimum 8 characters");
      return;
    }
    const res = await resetUserPassword(login, kind.kind, value);
    setStatus(res.ok ? "password changed" : res.message || "change failed");
  }

  return (
    <div className="acc-pass-row">
      <div className="acc-pass-note">
        {kind.label} {kind.note}
      </div>
      <div className="acc-pass-line">
        <input value={value} onChange={(e) => { setValue(e.target.value); setStatus(""); }} />
        <button type="button" onClick={onChangePassword}>Change</button>
        <button type="button" onClick={() => { setValue(generatePassword()); setStatus(""); }}>
          Generate
        </button>
      </div>
      <div className="acc-pass-status">{status || "minimum 8 characters"}</div>
    </div>
  );
}

// the reference opens the dialog with the two passwords already generated
function NewPassword({ label, value, onChange }) {
  return (
    <>
      <label>{label}</label>
      <span className="acc-new-pass">
        <input value={value} onChange={(e) => onChange(e.target.value)} />
        <button type="button" title="Generate" onClick={() => onChange(generatePassword())}>
          ↻
        </button>
      </span>
    </>
  );
}

const newDraft = (group) => ({
  login: "next",
  group: group || "",
  name: "",
  last_name: "",
  middle_name: "",
  company: "",
  email: "",
  phone: "",
  country: "",
  zip_code: "",
  state: "",
  city: "",
  address: "",
  rights: 0x0001 | 0x0002 | 0x0020 | 0x0040 | 0x0100,
  password_main: generatePassword(),
  password_investor: generatePassword(),
  password_phone: "",
});

/** The trading account dialog; "new" creates via CrtUser, edits diff through UptUser. */
export function AccountDialog({ login, onClose, onSaved }) {
  const isNew = login === "new";
  const { groups } = useGroups();
  const [activeTab, setActiveTab] = useState(isNew ? "Personal" : "Overview");
  const TABS = isNew ? NEW_TABS : EDIT_TABS;
  const [draft, setDraft] = useState(isNew ? newDraft(groups[0]?.group) : null);
  const [original, setOriginal] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(login);
  const close = useDialogStack(onClose);

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
      if (!draft.group || !draft.name.trim()) {
        setError("Group and first name are required");
        return;
      }
      if (draft.password_main.length < 8 || draft.password_investor.length < 8) {
        setError("Passwords need at least 8 characters");
        return;
      }
      const body = {
        group: draft.group,
        name: draft.name.trim(),
        last_name: draft.last_name,
        middle_name: draft.middle_name,
        company: draft.company,
        email: draft.email.trim(),
        phone: draft.phone,
        country: draft.country,
        state: draft.state,
        zip_code: draft.zip_code,
        city: draft.city,
        address: draft.address,
        rights: draft.rights,
        password_main: draft.password_main,
        password_investor: draft.password_investor,
      };
      // the literal next means the closest free number
      const preferred = Number(draft.login);
      if (Number.isInteger(preferred) && preferred > 0) body.login = preferred;
      if (draft.password_phone) body.password_phone = draft.password_phone;
      const res = await createUser(body);
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
    close();
  }

  // sorted so a folder's groups sit together, labelled with the full path as the reference shows
  const groupOptions = [...groups]
    .sort((a, b) => a.group.localeCompare(b.group))
    .map((g) => ({ value: g.group, label: g.group }));

  // the reference's New Account window: one page, the password block beside the name rows
  function newAccountPage() {
    return (
      <div className="config-panel active">
        <div className="form-grid acc-new-grid">
          <label>Preferred login</label>
          <span className="acc-new-cell">
            <input value={draft.login ?? ""} onChange={(e) => set("login", e.target.value)} />
          </span>
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <label>Group</label>
          <PropSelect fill value={draft.group} options={groupOptions} onChange={(v) => set("group", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <Field label="First name" value={draft.name} onChange={(v) => set("name", v)} />
          <span className="acc-new-pass-head no-colon">Passwords</span>

          <Field label="Last name" value={draft.last_name} onChange={(v) => set("last_name", v)} />
          <NewPassword label="Master" value={draft.password_main} onChange={(v) => set("password_main", v)} />

          <Field label="Middle name" value={draft.middle_name} onChange={(v) => set("middle_name", v)} />
          <NewPassword label="Investor" value={draft.password_investor} onChange={(v) => set("password_investor", v)} />

          <Field label="Company" value={draft.company} onChange={(v) => set("company", v)} />
          <NewPassword label="Phone" value={draft.password_phone} onChange={(v) => set("password_phone", v)} />

          <Field label="Email" value={draft.email} onChange={(v) => set("email", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <Field label="Phone" value={draft.phone} onChange={(v) => set("phone", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <Field label="Country" value={draft.country} onChange={(v) => set("country", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <Field label="Zip code" value={draft.zip_code} onChange={(v) => set("zip_code", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <Field label="State" value={draft.state} onChange={(v) => set("state", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <Field label="City" value={draft.city} onChange={(v) => set("city", v)} />
          <span className="acc-new-gap" />
          <span className="acc-new-gap" />

          <label>Address</label>
          <span className="acc-new-address">
            <input value={draft.address ?? ""} onChange={(e) => set("address", e.target.value)} />
          </span>
        </div>
      </div>
    );
  }

  function panel(tab) {
    if (!draft) return null;
    switch (tab) {
      case "Overview":
        return <AccountOverviewTab login={login} user={draft} />;
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
          <div className="acc-pass-tab">
            {PASSWORD_KINDS.map((kind) => (
              <PasswordRow key={kind.kind} login={login} kind={kind} />
            ))}
            <p className="acc-pass-warn">
              Changing the master password ends the account's open sessions.
            </p>
          </div>
        );
    }
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          width={isNew ? 700 : 613}
          height={isNew ? 560 : 560}
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          title={isNew ? "New Account" : `Account: ${login} — ${draft?.name ?? "…"}`}
          tabs={
            isNew ? null : (
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
            )
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
          {isNew
            ? draft && newAccountPage()
            : TABS.map((tab) => (
                <div key={tab} className={`config-panel${activeTab === tab ? " active" : ""}`}>
                  {activeTab === tab && panel(tab)}
                </div>
              ))}
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

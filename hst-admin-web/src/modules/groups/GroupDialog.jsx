import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { createGroup, fetchGroup, updateGroup } from "@/api/endpoints/groups.js";
import {
  GroupCommonTab,
  GroupCompanyTab,
  GroupMarginTab,
  GroupNewsMailTab,
  GroupPermissionsTab,
  GroupReportsTab,
} from "./GroupTabs.jsx";
import { GroupCommissionsTab, GroupSymbolsTab } from "./GroupSubTabs.jsx";

const TABS = [
  "Common",
  "Company",
  "News & Mail",
  "Permissions",
  "Margin",
  "Symbols",
  "Commissions",
  "Reports",
];

const newDraft = (folderPath) => ({
  group: folderPath ? `${folderPath}\\` : "",
  currency: "USD",
  currency_digits: 2,
  auth_mode: 0,
  auth_password_min: 8,
  permission_flags: 2 | 16,
  company: "",
  news_mode: 2,
  mail_mode: 1,
  trade_flags: 1 | 2 | 4,
  trade_transfer_mode: 0,
  trade_interest_rate: 0,
  trade_virtual_credit: 0,
  margin_mode: 0,
  margin_call: 50,
  margin_stop_out: 30,
  margin_so_mode: 0,
  margin_free_mode: 1,
  margin_free_profit_mode: 0,
  margin_leverage_id: 0,
  limit_history: 0,
  limit_orders: 0,
  limit_symbols: 0,
  limit_positions: 0,
  demo_deposit: 0,
  demo_leverage: 0,
  reports_mode: 0,
  reports_flags: 0,
});

/** Fields UptGroup can write; the diff never sends anything else. */
const PATCH_FIELDS = [
  "group",
  "permission_flags", "auth_mode", "auth_password_min", "company", "company_page",
  "company_email", "company_deposit", "company_withdrawal",
  "company_support_page", "company_support_email", "company_catalog",
  "currency", "currency_digits", "reports_mode", "reports_flags", "reports_email",
  "reports_smtp", "reports_smtp_login", "news_mode", "news_category", "mail_mode",
  "trade_flags", "trade_interest_rate", "trade_virtual_credit", "trade_transfer_mode",
  "margin_free_mode", "margin_so_mode", "margin_call", "margin_stop_out",
  "margin_free_profit_mode", "margin_mode", "margin_leverage_id", "limit_history", "limit_orders",
  "limit_symbols", "limit_positions", "limit_positions_volume", "demo_leverage", "demo_deposit",
];

/** The 8-tab group dialog. groupId "new" creates; an existing group renames in place. */
export function GroupDialog({ groupId, folderPath = "", onClose, onSaved }) {
  const isNew = groupId === "new";
  const [activeTab, setActiveTab] = useState("Common");
  const [draft, setDraft] = useState(isNew ? newDraft(folderPath) : null);
  const [original, setOriginal] = useState(null);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(groupId);

  useEffect(() => {
    if (isNew) return;
    fetchGroup(groupId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load group");
        return;
      }
      setDraft(res.data);
      setOriginal(res.data);
    });
  }, [groupId, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    if (!draft) return;
    if (!draft.group?.trim() || draft.group.endsWith("\\")) {
      setError("Group name is required");
      return;
    }

    if (isNew) {
      const body = Object.fromEntries(
        PATCH_FIELDS.filter((k) => draft[k] !== undefined && draft[k] !== null).map((k) => [k, draft[k]]),
      );
      const res = await createGroup({ ...body, group: draft.group.trim() });
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
    } else {
      const patch = {};
      for (const key of PATCH_FIELDS) {
        if (draft[key] !== undefined && draft[key] !== original[key]) patch[key] = draft[key];
      }
      if (patch.group) patch.group = patch.group.trim();
      if (Object.keys(patch).length) {
        const res = await updateGroup(groupId, patch);
        if (!res.ok) {
          setError(res.message || "save failed");
          return;
        }
      }
    }
    onSaved();
    onClose();
  }

  const title = `Group: ${isNew ? draft.group || "New" : draft?.group ?? "…"}`;

  function panel(tab) {
    if (!draft) return null;
    switch (tab) {
      case "Common":
        return <GroupCommonTab g={draft} set={set} />;
      case "Company":
        return <GroupCompanyTab g={draft} set={set} />;
      case "News & Mail":
        return <GroupNewsMailTab g={draft} set={set} />;
      case "Permissions":
        return <GroupPermissionsTab g={draft} set={set} />;
      case "Margin":
        return <GroupMarginTab g={draft} set={set} />;
      case "Symbols":
        return <GroupSymbolsTab groupId={groupId} />;
      case "Commissions":
        return <GroupCommissionsTab groupId={groupId} />;
      case "Reports":
        return <GroupReportsTab g={draft} set={set} />;
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
          width={613}
          height={620}
          draggable
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          title={title}
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
              <button type="button" className="config-help" disabled>Help</button>
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

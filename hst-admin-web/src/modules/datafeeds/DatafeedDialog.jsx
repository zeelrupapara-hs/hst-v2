import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import {
  createDatafeed,
  createDatafeedParam,
  createDatafeedSymbol,
  createDatafeedTranslate,
  deleteDatafeedParam,
  deleteDatafeedSymbol,
  deleteDatafeedTranslate,
  fetchDatafeed,
  fetchDatafeedModules,
  updateDatafeed,
  updateDatafeedParam,
  updateDatafeedTranslate,
} from "@/api/endpoints/datafeeds.js";
import {
  DatafeedCommonTab,
  DatafeedParametersTab,
  DatafeedSymbolsTab,
  DatafeedTimeoutsTab,
  DatafeedTranslationsTab,
} from "./DatafeedTabs.jsx";

const TABS = [
  { id: "common", label: "Common", Panel: DatafeedCommonTab },
  { id: "symbols", label: "Symbols", Panel: DatafeedSymbolsTab },
  { id: "translations", label: "Translations", Panel: DatafeedTranslationsTab },
  { id: "parameters", label: "Parameters", Panel: DatafeedParametersTab },
  { id: "timeouts", label: "Timeouts", Panel: DatafeedTimeoutsTab },
];

const MAIN_FIELDS = [
  "name", "module", "enable", "mode", "gateway_server", "feed_server", "feed_login",
  "feed_password", "gateway_login", "gateway_password", "timeout", "timeout_reconnect",
  "timeout_sleep", "attempts_sleep", "company", "issuer", "allow_import_symbols",
];

const newDraft = () => ({
  name: "",
  module: "",
  enable: 1,
  mode: 1,
  feed_server: "",
  feed_login: 0,
  gateway_server: "",
  gateway_login: 0,
  timeout: 0,
  timeout_reconnect: 5,
  timeout_sleep: 60,
  attempts_sleep: 10,
  feed_symbols: [],
  translates: [],
  params: [],
});

// A scope row edits as one text (mask or symbol); the wire splits it into symbol vs path.
const toScopeRow = (row) => ({ ...row, scope: row.path || row.symbol || "*" });
const fromScope = (scope) =>
  /[\\*!]/.test(scope) ? { symbol: "", path: scope } : { symbol: scope, path: "" };

/** Post-OK sync of one sub-list: new rows POST, missing DELETE, changed PATCH (or replace). */
async function syncRows(before, after, key, { create, update, remove }) {
  const beforeById = new Map(before.map((r) => [r[key], r]));
  const seen = new Set();
  for (const row of after) {
    const id = row[key];
    if (!id) {
      await create(row);
      continue;
    }
    seen.add(id);
    const prev = beforeById.get(id);
    if (prev && JSON.stringify(prev) !== JSON.stringify(row)) {
      if (update) await update(id, row);
      else {
        await remove(id);
        await create(row);
      }
    }
  }
  for (const id of beforeById.keys()) if (!seen.has(id)) await remove(id);
}

/** @param {{feedId: number|"new", onClose: Function, onSaved: Function}} props */
export function DatafeedDialog({ feedId, onClose, onSaved }) {
  const isNew = feedId === "new";
  const [activeTab, setActiveTab] = useState("common");
  const [draft, setDraft] = useState(isNew ? newDraft() : null);
  const [original, setOriginal] = useState(null);
  const [modules, setModules] = useState([]);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(feedId);

  useEffect(() => {
    fetchDatafeedModules(1).then((res) => res.ok && setModules(res.data || []));
    if (isNew) return;
    fetchDatafeed(feedId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load data feed");
        return;
      }
      const detail = { ...res.data, feed_symbols: (res.data.feed_symbols || []).map(toScopeRow) };
      setDraft(detail);
      setOriginal(detail);
    });
  }, [feedId, isNew]);

  const set = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  async function handleOk() {
    if (!draft.name?.trim() || !draft.module) {
      setError("Name and module are required");
      return;
    }

    let id = feedId;
    if (isNew) {
      const res = await createDatafeed(Object.fromEntries(MAIN_FIELDS.map((k) => [k, draft[k]]).filter(([, v]) => v !== undefined)));
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
      id = res.data?.datafeed_id;
    } else {
      const patch = {};
      for (const key of MAIN_FIELDS) {
        if (draft[key] !== undefined && draft[key] !== original[key]) patch[key] = draft[key];
      }
      if (Object.keys(patch).length) {
        const res = await updateDatafeed(id, patch);
        if (!res.ok) {
          setError(res.message || "save failed");
          return;
        }
      }
    }

    await syncRows(original?.feed_symbols || [], draft.feed_symbols || [], "feed_symbol_id", {
      create: (row) => createDatafeedSymbol(id, { ...fromScope(row.scope || "*"), exclude: !!row.exclude }),
      remove: (rowId) => deleteDatafeedSymbol(id, rowId),
    });
    await syncRows(original?.translates || [], draft.translates || [], "translate_id", {
      create: (row) =>
        createDatafeedTranslate(id, {
          symbol: row.symbol || "",
          source: row.source || "",
          bid_markup: row.bid_markup || 0,
          ask_markup: row.ask_markup || 0,
          digits: row.digits || 0,
        }),
      update: (rowId, row) =>
        updateDatafeedTranslate(id, rowId, {
          symbol: row.symbol,
          source: row.source,
          bid_markup: row.bid_markup,
          ask_markup: row.ask_markup,
        }),
      remove: (rowId) => deleteDatafeedTranslate(id, rowId),
    });
    await syncRows(original?.params || [], draft.params || [], "param_id", {
      create: (row) => createDatafeedParam(id, { param_key: row.param_key, value: String(row.value ?? "") }),
      update: (rowId, row) =>
        updateDatafeedParam(id, rowId, { param_key: row.param_key, value: String(row.value ?? "") }),
      remove: (rowId) => deleteDatafeedParam(id, rowId),
    });

    onSaved();
    onClose();
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
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          className="df-config-window"
          width={570}
          height={419}
          title="Data Feed"
          tabs={
            <div className="config-tabs df-config-tabs">
              {TABS.map((tab) => (
                <button
                  key={tab.id}
                  type="button"
                  className={activeTab === tab.id ? "active" : ""}
                  onClick={() => setActiveTab(tab.id)}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          }
          footer={
            <div className="config-actions df-config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>
                OK
              </button>
              <button type="button" onClick={onClose}>
                Cancel
              </button>
              <button type="button" className="config-help" disabled>
                Help
              </button>
            </div>
          }
        >
          {draft &&
            TABS.map(({ id: tabId, Panel }) => (
              <div key={tabId} className={`config-panel${activeTab === tabId ? " active" : ""}`}>
                <Panel d={draft} set={set} modules={modules} />
              </div>
            ))}
        </SettingsDialog>
      </div>
    </div>
  );
}

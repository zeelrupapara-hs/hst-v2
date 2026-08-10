import { useEffect, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack, prevTabEscape } from "@/hooks/useDialogStack.jsx";
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
import {
  datafeedCreateBody,
  datafeedPatchBody,
  newDatafeedDraft,
  normalizeDatafeedDraft,
} from "./datafeedPayload.js";

const TABS = [
  { id: "common", label: "Common", Panel: DatafeedCommonTab },
  { id: "symbols", label: "Symbols", Panel: DatafeedSymbolsTab },
  { id: "translations", label: "Translations", Panel: DatafeedTranslationsTab },
  { id: "parameters", label: "Parameters", Panel: DatafeedParametersTab },
  { id: "timeouts", label: "Timeouts", Panel: DatafeedTimeoutsTab },
];

const toScopeRow = (row) => ({ ...row, scope: row.path || row.symbol || "*" });
const fromScope = (scope) =>
  /[\\*!]/.test(scope) ? { symbol: "", path: scope } : { symbol: scope, path: "" };

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

async function syncParamPriorities(before, after, updateParam) {
  for (let i = 0; i < after.length; i++) {
    const row = after[i];
    if (!row.param_id) continue;
    const beforeIdx = before.findIndex((p) => p.param_id === row.param_id);
    if (beforeIdx !== i) await updateParam(row.param_id, { priority: i });
  }
}

/** @param {{feedId: number|"new", onClose: Function, onSaved: Function}} props */
export function DatafeedDialog({ feedId, onClose, onSaved }) {
  const isNew = feedId === "new";
  const [activeTab, setActiveTab] = useState("common");
  const [draft, setDraft] = useState(isNew ? newDatafeedDraft() : null);
  const [original, setOriginal] = useState(null);
  const [modules, setModules] = useState([]);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag(feedId);
  const close = useDialogStack(onClose, {
    onEscape: () => prevTabEscape(TABS, activeTab, setActiveTab),
  });

  useEffect(() => {
    if (isNew) return;
    fetchDatafeed(feedId).then((res) => {
      if (!res.ok) {
        setError(res.message || "failed to load data feed");
        return;
      }
      const detail = normalizeDatafeedDraft({
        ...res.data,
        feed_symbols: (res.data.feed_symbols || []).map(toScopeRow),
      });
      setDraft(detail);
      setOriginal(detail);
    });
  }, [feedId, isNew]);

  useEffect(() => {
    if (!draft) return;
    const mode = draft.mode === 2 ? 2 : 1;
    fetchDatafeedModules(mode).then((res) => {
      if (!res.ok) return;
      const next = res.data || [];
      setModules(next);
      setDraft((prev) => {
        if (!prev) return prev;
        const names = new Set(next.map((m) => m.module));
        if (prev.module && names.has(prev.module)) return prev;
        const fallback = next[0]?.module || "";
        if (!fallback) return prev.module ? { ...prev, module: "" } : prev;
        return { ...prev, module: fallback };
      });
    });
  }, [draft?.mode]);

  const set = (key, value) => {
    setDraft((prev) => ({ ...prev, [key]: value }));
    if (error && (key === "name" || key === "module")) setError("");
  };

  async function handleOk() {
    setError("");
    const module = draft.module || modules[0]?.module || "";
    if (!draft.name?.trim() || !module) {
      setError("Name and module are required");
      return;
    }
    const payload =
      module !== draft.module ? normalizeDatafeedDraft({ ...draft, module }) : normalizeDatafeedDraft(draft);

    let id = feedId;
    if (isNew) {
      const res = await createDatafeed(datafeedCreateBody(payload));
      if (!res.ok) {
        setError(res.message || "create failed");
        return;
      }
      id = res.data?.datafeed_id;
    } else {
      const patch = datafeedPatchBody(payload, original);
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
    await syncParamPriorities(original?.params || [], draft.params || [], (rowId, patch) =>
      updateDatafeedParam(id, rowId, patch)
    );

    onSaved();
    close();
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          onClose={close}
          onTitlePointerDown={onTitlePointerDown}
          className="df-config-window"
          width={570}
          height={520}
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
              <button type="button" onClick={close}>
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
                <Panel d={draft} set={set} modules={modules} feedId={feedId} isNew={isNew} />
              </div>
            ))}
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

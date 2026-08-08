import { useState } from "react";
import { loginField } from "./datafeedPayload.js";
import { resolveDatafeedSymbols } from "@/api/endpoints/datafeeds.js";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { DatafeedTabIntro, DfListPanel, DfRowIcon, ParamTypeIcon } from "./DfListPanel.jsx";

const MODE_OPTIONS = [
  { value: 1, label: "Quotes" },
  { value: 2, label: "News" },
];

const RECONNECT_OPTIONS = [1, 2, 3, 5, 10, 15, 30, 60];

function Field({ label, value, onChange, type = "text", width, suffix }) {
  return (
    <>
      <label>{label}</label>
      <input
        type={type}
        value={value ?? ""}
        className={width ? `df-input-${width}` : ""}
        onChange={(e) => onChange(e.target.value)}
      />
      {suffix && <span className="df-suffix">{suffix}</span>}
    </>
  );
}

export function DatafeedCommonTab({ d, set, modules }) {
  const [showNetwork, setShowNetwork] = useState(false);
  const moduleOptions = (modules || []).map((m) => ({ value: m.module, label: m.label || m.module }));
  if (d.module && !moduleOptions.some((o) => o.value === d.module)) {
    moduleOptions.push({ value: d.module, label: d.module });
  }
  const selected = (modules || []).find((m) => m.module === d.module);

  return (
    <>
      <DatafeedTabIntro>
        {selected?.description ||
          "The main information about the data feed server is specified on this tab."}
      </DatafeedTabIntro>
      <div className="form-grid sym-form-two-col df-common-form">
        <span className="df-form-pad" aria-hidden="true" />
        <label className="no-colon df-check-cell">
          <input
            type="checkbox"
            className="df-checkbox"
            checked={d.enable === 1}
            onChange={(e) => set("enable", e.target.checked ? 1 : 0)}
          />
          Enable
        </label>
        <Field label="Name" value={d.name} onChange={(v) => set("name", v)} width="wide" />
        <label>Module</label>
        <div className="df-module-pair">
          <PropSelect fill value={d.module ?? ""} options={moduleOptions} onChange={(v) => set("module", v)} />
          <PropSelect fill value={d.mode ?? 1} options={MODE_OPTIONS} onChange={(v) => set("mode", Number(v))} />
        </div>
        <Field label="Feed server" value={d.feed_server} onChange={(v) => set("feed_server", v)} width="wide" />
        <Field
          label="Feed login"
          value={loginField(d.feed_login)}
          onChange={(v) => set("feed_login", v)}
        />
        <Field
          label="Password"
          type="password"
          value={d.feed_password ?? ""}
          onChange={(v) => set("feed_password", v)}
        />
        <span className="df-form-pad" aria-hidden="true" />
        <button type="button" className="df-link" onClick={() => setShowNetwork((v) => !v)}>
          Show additional network settings
        </button>
        {showNetwork && (
          <>
            <Field
              label="Gateway server"
              value={d.gateway_server}
              onChange={(v) => set("gateway_server", v)}
              width="wide"
            />
            <Field
              label="Gateway login"
              value={loginField(d.gateway_login)}
              onChange={(v) => set("gateway_login", v)}
            />
            <Field
              label="Gateway password"
              type="password"
              value={d.gateway_password ?? ""}
              onChange={(v) => set("gateway_password", v)}
            />
          </>
        )}
        <Field label="Company" value={d.company} onChange={(v) => set("company", v)} width="wide" />
        <Field label="Issuer" value={d.issuer} onChange={(v) => set("issuer", v)} width="wide" />
        <Field
          label="Timeout"
          value={d.timeout ?? 0}
          onChange={(v) => set("timeout", Number(v) || 0)}
          suffix="seconds"
        />
      </div>
    </>
  );
}

export function DatafeedSymbolsTab({ d, set, feedId, isNew }) {
  const [resolveMsg, setResolveMsg] = useState("");

  async function onResolve() {
    if (isNew || !feedId) {
      setResolveMsg("Save the feed first to resolve scope.");
      return;
    }
    setResolveMsg("Resolving…");
    const res = await resolveDatafeedSymbols(feedId);
    if (!res.ok) {
      setResolveMsg(res.message || "resolve failed");
      return;
    }
    setResolveMsg(`${res.data?.count ?? 0} symbols in scope`);
  }

  return (
    <DfListPanel
      intro="Please specify the symbols for which the data feed will translate quotes."
      columns={[{ id: "scope", label: "", defaultValue: "*", editor: "symbols" }]}
      rows={d.feed_symbols || []}
      rowKey={(row, i) => row.feed_symbol_id ?? `new-${i}`}
      canEdit
      showMove={false}
      headless
      showAddRow
      beforeTable={
        <>
          <label className="no-colon df-check-cell df-symbols-import">
            <input
              type="checkbox"
              className="df-checkbox"
              checked={d.allow_import_symbols === 1}
              onChange={(e) => set("allow_import_symbols", e.target.checked ? 1 : 0)}
            />
            Allow importing symbol settings
          </label>
          {!isNew && (
            <div className="df-symbols-resolve">
              <button type="button" className="df-link" onClick={onResolve}>
                Resolve symbol scope
              </button>
              {resolveMsg && <span className="df-symbols-resolve-msg">{resolveMsg}</span>}
            </div>
          )}
        </>
      }
      onChangeRows={(next) => set("feed_symbols", next)}
      emptyLabel=""
      renderCellPrefix={() => <DfRowIcon />}
    />
  );
}

export function DatafeedTranslationsTab({ d, set }) {
  return (
    <DfListPanel
      intro={
        "If necessary, please specify parameters for converting data transmitted through the data feed: " +
        "name of the source symbol in the external system and value of correction of incoming prices. " +
        "If any of the parameters is not set, its source values will be used."
      }
      columns={[
        { id: "symbol", label: "Symbol", defaultValue: "", editor: "symbol" },
        { id: "source", label: "Source", defaultValue: "" },
        { id: "bid_markup", label: "Bid", defaultValue: 0, editor: "number", align: "num" },
        { id: "ask_markup", label: "Ask", defaultValue: 0, editor: "number", align: "num" },
      ]}
      rows={d.translates || []}
      rowKey={(row, i) => row.translate_id ?? `new-${i}`}
      canEdit
      onChangeRows={(next) => set("translates", next)}
      emptyLabel="No translations"
      renderCellPrefix={(row, col) => (col.id === "symbol" ? <DfRowIcon /> : null)}
    />
  );
}

export function DatafeedParametersTab({ d, set }) {
  return (
    <DfListPanel
      intro={
        "Please specify parameters of the data feed. These parameters are specific for each type of " +
        'data feed. They allow using additional settings that were not available in the "Server" tab.'
      }
      columns={[
        { id: "param_key", label: "Parameter", defaultValue: "New Parameter" },
        { id: "value", label: "Value", defaultValue: "" },
      ]}
      rows={d.params || []}
      rowKey={(row, i) => row.param_id ?? `new-${i}`}
      canEdit
      onChangeRows={(next) => set("params", next)}
      emptyLabel="No parameters"
      onDefault={() => set("params", [])}
      renderCellPrefix={(row, col) =>
        col.id === "param_key" ? <ParamTypeIcon type={row.type} /> : null
      }
    />
  );
}

export function DatafeedTimeoutsTab({ d, set }) {
  const num = (key) => (v) => set(key, Number(v) || 0);
  return (
    <>
      <DatafeedTabIntro>
        Please set up data feed timeout parameters for connection errors. These parameters allow quick
        restoring of a lost network connection and limiting the frequency of reconnections in case of
        continuous network problems.
      </DatafeedTabIntro>
      <div className="form-grid df-timeouts-form">
        <label>Interval between reconnections</label>
        <PropSelect
          fill
          value={d.timeout_reconnect ?? 1}
          options={RECONNECT_OPTIONS}
          onChange={(v) => set("timeout_reconnect", Number(v) || 0)}
        />
        <span className="df-suffix">seconds</span>
        <Field
          label="Number of reconnection attempts"
          value={d.attempts_sleep}
          onChange={num("attempts_sleep")}
        />
        <span className="df-suffix" aria-hidden="true" />
        <Field
          label="Interval between series of reconnections"
          value={d.timeout_sleep}
          onChange={num("timeout_sleep")}
          suffix="seconds"
        />
      </div>
    </>
  );
}

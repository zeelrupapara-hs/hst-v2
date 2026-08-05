import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { DatafeedTabIntro, DfListPanel, ParamTypeIcon } from "./DfListPanel.jsx";

const MODE_OPTIONS = [
  { value: 1, label: "Quotes" },
  { value: 2, label: "News" },
];

function Field({ label, value, onChange, type = "text", wide }) {
  return (
    <>
      <label>{label}</label>
      <input
        type={type}
        value={value ?? ""}
        className={wide ? "df-input-wide" : ""}
        onChange={(e) => onChange(e.target.value)}
      />
    </>
  );
}

export function DatafeedCommonTab({ d, set, modules }) {
  const moduleOptions = (modules || []).map((m) => ({ value: m.module, label: m.label || m.module }));
  if (d.module && !moduleOptions.some((o) => o.value === d.module)) {
    moduleOptions.push({ value: d.module, label: d.module });
  }

  return (
    <>
      <DatafeedTabIntro>
        The main information about the data feed server is specified on this tab.
      </DatafeedTabIntro>
      <div className="form-grid sym-form-two-col df-common-form">
        <label>Enable</label>
        <input
          type="checkbox"
          className="df-checkbox"
          checked={d.enable === 1}
          onChange={(e) => set("enable", e.target.checked ? 1 : 0)}
        />
        <span className="df-form-pad" aria-hidden="true" />
        <Field label="Name" value={d.name} onChange={(v) => set("name", v)} wide />
        <span className="df-form-pad" aria-hidden="true" />
        <label>Module</label>
        <div className="df-module-pair">
          <PropSelect fill value={d.module ?? ""} options={moduleOptions} onChange={(v) => set("module", v)} />
          <PropSelect fill value={d.mode ?? 1} options={MODE_OPTIONS} onChange={(v) => set("mode", Number(v))} />
        </div>
        <Field label="Feed server" value={d.feed_server} onChange={(v) => set("feed_server", v)} wide />
        <Field label="Feed login" value={d.feed_login} onChange={(v) => set("feed_login", Number(v) || 0)} />
        <Field
          label="Password"
          type="password"
          value={d.feed_password ?? ""}
          onChange={(v) => set("feed_password", v)}
        />
        <Field label="Gateway server" value={d.gateway_server} onChange={(v) => set("gateway_server", v)} wide />
        <Field label="Gateway login" value={d.gateway_login} onChange={(v) => set("gateway_login", Number(v) || 0)} />
        <Field
          label="Gateway password"
          type="password"
          value={d.gateway_password ?? ""}
          onChange={(v) => set("gateway_password", v)}
        />
        <p className="df-attention df-attention-full">
          When you change any of the parameters, the data feed automatically restarts to apply the
          changes.
        </p>
      </div>
    </>
  );
}

export function DatafeedSymbolsTab({ d, set }) {
  return (
    <DfListPanel
      intro="Specify symbols and groups available to this data feed. Use masks (*, !)."
      columns={[
        { id: "scope", label: "Symbol / Group", defaultValue: "*" },
        { id: "exclude", label: "Exclude", defaultValue: false, editor: "yesno" },
      ]}
      rows={d.feed_symbols || []}
      rowKey={(row, i) => row.feed_symbol_id ?? `new-${i}`}
      canEdit
      onChangeRows={(next) => set("feed_symbols", next)}
      emptyLabel="No symbols configured"
    />
  );
}

export function DatafeedTranslationsTab({ d, set }) {
  return (
    <DfListPanel
      intro="Map external source symbol names to platform symbols. The first matching rule wins."
      columns={[
        { id: "symbol", label: "Symbol", defaultValue: "" },
        { id: "source", label: "Source", defaultValue: "" },
        { id: "bid_markup", label: "Bid", defaultValue: 0, editor: "number" },
        { id: "ask_markup", label: "Ask", defaultValue: 0, editor: "number" },
      ]}
      rows={d.translates || []}
      rowKey={(row, i) => row.translate_id ?? `new-${i}`}
      canEdit
      onChangeRows={(next) => set("translates", next)}
      emptyLabel="No translations"
    />
  );
}

export function DatafeedParametersTab({ d, set }) {
  return (
    <DfListPanel
      intro="Please specify parameters of the data feed. These parameters are specific for each type of data feed."
      columns={[
        { id: "param_key", label: "Parameter", defaultValue: "New Parameter" },
        { id: "value", label: "Value", defaultValue: "" },
      ]}
      rows={d.params || []}
      rowKey={(row, i) => row.param_id ?? `new-${i}`}
      canEdit
      onChangeRows={(next) => set("params", next)}
      emptyLabel="No parameters"
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
      <DatafeedTabIntro>Reconnection timeout settings for the data feed.</DatafeedTabIntro>
      <div className="form-grid df-timeouts-form">
        <Field label="Interval between reconnections" value={d.timeout_reconnect} onChange={num("timeout_reconnect")} />
        <Field label="Number of reconnection attempts" value={d.attempts_sleep} onChange={num("attempts_sleep")} />
        <Field label="Interval between series of reconnections" value={d.timeout_sleep} onChange={num("timeout_sleep")} />
        <Field label="Connection timeout" value={d.timeout} onChange={num("timeout")} />
      </div>
    </>
  );
}

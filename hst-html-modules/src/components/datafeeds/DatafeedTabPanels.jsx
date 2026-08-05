import { DfListPanel, ParamTypeIcon, DatafeedTabIntro } from "./DfListPanel.jsx";
import { PropSelect } from "../ui/PropSelect.jsx";
import { FEEDER_CATEGORIES } from "../../lib/datafeedModules.js";
import { useFeedSymbolOptions } from "../../hooks/useFeedSymbolOptions.js";

function DfField({
  label,
  value,
  type = "text",
  readOnly,
  onChange,
  wide,
  checkbox,
  options,
  selectOptions,
}) {
  if (type === "select") {
    const opts = selectOptions?.length
      ? selectOptions
      : (options || []).map((o) => ({ value: o, label: o }));
    return (
      <>
        <label>{label}</label>
        <PropSelect
          fill
          disabled={readOnly}
          value={value ?? ""}
          options={opts.map((o) => ({
            value: o.value,
            label: o.label,
            description: o.description,
          }))}
          onChange={readOnly ? undefined : onChange}
        />
      </>
    );
  }

  if (checkbox) {
    return (
      <>
        <label>{label}</label>
        <input
          type="checkbox"
          className="df-checkbox"
          disabled={readOnly}
          checked={!!value}
          onChange={
            readOnly
              ? undefined
              : (e) => onChange?.(e.target.checked ? 1 : 0)
          }
        />
      </>
    );
  }

  return (
    <>
      <label>{label}</label>
      <input
        type={type}
        readOnly={readOnly}
        value={value ?? ""}
        className={wide ? "df-input-wide" : ""}
        onChange={
          readOnly ? undefined : (e) => onChange?.(e.target.value)
        }
      />
    </>
  );
}

export function DatafeedCommonTab({ d, modules, onFieldChange, onModeChange }) {
  if (!d) return null;
  const canEdit = !!onFieldChange;
  const ro = !canEdit;
  const set = (key) => (v) => onFieldChange(key, v);
  const moduleOptions = modules?.length
    ? modules.map((m) => ({
        value: m.module,
        label: m.label || m.module,
        description: m.description,
      }))
    : d.module
      ? [{ value: d.module, label: d.module }]
      : [];

  return (
    <>
      <DatafeedTabIntro>
        The main information about the data feed server is specified on this tab.
      </DatafeedTabIntro>
      <div className="form-grid sym-form-two-col df-common-form">
        <DfField
          label="Enable"
          checkbox
          value={d.enable}
          readOnly={ro}
          onChange={set("enable")}
        />
        <span className="df-form-pad" aria-hidden="true" />
        <span className="df-form-pad" aria-hidden="true" />
        <DfField
          label="Name"
          value={d.name}
          readOnly={ro}
          onChange={set("name")}
        />
        <span className="df-form-pad" aria-hidden="true" />
        <span className="df-form-pad" aria-hidden="true" />
        <label>Module</label>
        <div className="df-module-pair">
          <PropSelect
            fill
            disabled={ro || !moduleOptions.length}
            value={d.module ?? ""}
            options={moduleOptions.map((o) => ({
              value: o.value,
              label: o.label,
              description: o.description,
            }))}
            onChange={ro ? undefined : set("module")}
          />
          <PropSelect
            fill
            disabled={ro}
            value={String(d.mode ?? 1)}
            options={FEEDER_CATEGORIES.map((c) => ({
              value: String(c.mode),
              label: c.label,
            }))}
            onChange={ro ? undefined : (v) => onModeChange?.(Number(v))}
          />
        </div>
        <DfField
          label="Feed server"
          value={d.feed_server}
          wide
          readOnly={ro}
          onChange={set("feed_server")}
        />
        <DfField
          label="Feed login"
          value={d.feed_login}
          readOnly={ro}
          onChange={(v) => set("feed_login")(Number(v) || 0)}
        />
        <DfField
          label="Password"
          type="password"
          value={d.feed_password || ""}
          readOnly={ro}
          onChange={set("feed_password")}
        />

        <p className="df-attention df-attention-full">
          When you change any of the parameters, the data feed automatically
          restarts to apply the changes.
        </p>
      </div>
    </>
  );
}

export function DatafeedSymbolsTab({ d, onFieldChange }) {
  const symbolOptions = useFeedSymbolOptions();
  if (!d) return null;
  const rows = d.feed_symbols || [];
  const canEdit = !!onFieldChange;

  return (
    <DfListPanel
      intro="Specify symbols and groups available to this data feed. Use masks (*, !)."
      columns={[
        {
          id: "symbol",
          label: "Symbol / Group",
          defaultValue: "*",
          editor: "select",
          selectOptions: symbolOptions,
        },
        {
          id: "exclude",
          label: "Exclude",
          defaultValue: 0,
          editor: "yesno",
        },
      ]}
      rows={rows}
      rowKey={(row, i) => row.feed_symbol_id ?? i}
      canEdit={canEdit}
      onChangeRows={(next) => onFieldChange?.("feed_symbols", next)}
      emptyLabel="No symbols configured"
      beforeTable={
        <label className="df-inline-check">
          <input
            type="checkbox"
            checked={!!d.allow_import_symbols}
            disabled={!canEdit}
            onChange={
              canEdit
                ? (e) =>
                    onFieldChange(
                      "allow_import_symbols",
                      e.target.checked ? 1 : 0
                    )
                : undefined
            }
          />
          Allow importing symbol settings
        </label>
      }
    />
  );
}

export function DatafeedTranslationsTab({ d, onFieldChange }) {
  const symbolOptions = useFeedSymbolOptions();
  if (!d) return null;
  const rows = d.translates || [];
  const canEdit = !!onFieldChange;

  return (
    <DfListPanel
      intro="Map external source symbol names to platform symbols."
      columns={[
        {
          id: "symbol",
          label: "Symbol",
          defaultValue: "",
          editor: "select",
          selectOptions: symbolOptions.filter((o) => o.value !== "*"),
        },
        { id: "source", label: "Source", defaultValue: "", editor: "text" },
        { id: "bid_markup", label: "Bid", defaultValue: 0, editor: "number" },
        { id: "ask_markup", label: "Ask", defaultValue: 0, editor: "number" },
      ]}
      rows={rows}
      rowKey={(row, i) => row.translate_id ?? i}
      canEdit={canEdit}
      onChangeRows={(next) => onFieldChange?.("translates", next)}
      emptyLabel="No translations"
    />
  );
}

export function DatafeedParametersTab({ d, onFieldChange }) {
  if (!d) return null;
  const rows = d.params || [];
  const canEdit = !!onFieldChange;

  return (
    <DfListPanel
      intro={
        <>
          Please specify parameters of the data feed. These parameters are
          specific for each type of data feed. They allow using additional
          settings that were not available in the &apos;Server&apos; tab.
        </>
      }
      columns={[
        {
          id: "param_key",
          label: "Parameter",
          defaultValue: "New Parameter",
          editor: "text",
        },
        { id: "value", label: "Value", defaultValue: "", editor: "text" },
      ]}
      rows={rows}
      rowKey={(row, i) => row.param_id ?? i}
      canEdit={canEdit}
      showDefault
      onChangeRows={(next) => onFieldChange?.("params", next)}
      emptyLabel="No parameters"
      renderCellPrefix={(row, col) =>
        col.id === "param_key" ? <ParamTypeIcon type={row.type} /> : null
      }
    />
  );
}

export function DatafeedTimeoutsTab({ d, onFieldChange }) {
  if (!d) return null;
  const canEdit = !!onFieldChange;
  const ro = !canEdit;
  const set = (key) => (v) => onFieldChange(key, v);

  return (
    <>
      <DatafeedTabIntro>
        Reconnection timeout settings for the data feed.
      </DatafeedTabIntro>
      <div className="form-grid df-timeouts-form">
        <DfField
          label="Interval between reconnections"
          value={d.timeout_reconnect}
          readOnly={ro}
          onChange={(v) => set("timeout_reconnect")(Number(v) || 0)}
        />
        <DfField
          label="Number of reconnection attempts"
          value={d.attempts_sleep}
          readOnly={ro}
          onChange={(v) => set("attempts_sleep")(Number(v) || 0)}
        />
        <DfField
          label="Interval between series of reconnections"
          value={d.timeout_sleep}
          readOnly={ro}
          onChange={(v) => set("timeout_sleep")(Number(v) || 0)}
        />
        <DfField
          label="Connection timeout"
          value={d.timeout}
          readOnly={ro}
          onChange={(v) => set("timeout")(Number(v) || 0)}
        />
      </div>
    </>
  );
}

export const DATAFEED_TABS = [
  { id: "common", label: "Common", Panel: DatafeedCommonTab },
  { id: "symbols", label: "Symbols", Panel: DatafeedSymbolsTab },
  { id: "translations", label: "Translations", Panel: DatafeedTranslationsTab },
  { id: "parameters", label: "Parameters", Panel: DatafeedParametersTab },
  { id: "timeouts", label: "Timeouts", Panel: DatafeedTimeoutsTab },
];

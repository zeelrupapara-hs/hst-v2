import { useEffect, useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import { fetchLeverages } from "@/api/endpoints/leverages.js";
import { fetchMailServers } from "@/api/endpoints/mailServers.js";
import { newsLangSummary } from "@/constants/newsLanguages.js";
import { NewsLanguagesDialog } from "./NewsLanguagesDialog.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import { FlagCombo } from "@/modules/symbols/SymbolTabs.jsx";
import {
  AuthMode_name,
  Currency_options,
  FreeMarginMode_name,
  groupKind,
  HistoryLimit_name,
  limitOptions,
  demoLeverageOptions,
  MarginFlag_clearAcc,
  MarginFreeProfitMode_name,
  MarginMode_name,
  MarginMode_order,
  NewsMode_name,
  NotifyFlag_labels,
  PermissionFlag_forceOtp,
  PermissionFlag_labels,
  PermissionFlag_notifyMask,
  ReportsFlag_email,
  ReportsFlag_statements,
  ReportsFlag_support,
  ReportsMode_name,
  SignalsMode_options,
  StopOutMode_name,
  TradeFlag_labels,
  TradeFlag_signalsMask,
  TradeFlag_soCompensation,
  TradeFlag_soCompensationCredit,
  TradeFlag_soFullyHedged,
  TransferMode_name,
} from "@/constants/groups.js";

export function GroupTabIntro({ children }) {
  return (
    <div className="sym-sessions-intro">
      <span className="sym-tab-intro-icon" aria-hidden="true">
        <Icon id="groups" size={48} />
      </span>
      <p>{children}</p>
    </div>
  );
}

const enumOptions = (names, order) =>
  (order ?? Object.keys(names)).map((value) => ({ value: Number(value), label: names[value] }));

const decimals = (v, n) => (v == null || v === "" ? "" : Number(v).toFixed(n));

function Field({ label, value, onChange, onBlur, wide, readOnly, suffix, disabled }) {
  return (
    <>
      <label>{label}</label>
      <span className={suffix ? "grp-suffixed" : undefined}>
        <input
          type="text"
          readOnly={readOnly || !onChange}
          disabled={disabled}
          value={value ?? ""}
          className={wide ? "wide" : ""}
          onBlur={onBlur}
          onChange={onChange ? (e) => onChange(e.target.value) : undefined}
        />
        {suffix && <span className="grp-suffix">{suffix}</span>}
      </span>
    </>
  );
}

/** Fixed-decimal display without fighting the caret: the raw text wins until blur. */
function DecField({ label, value, digits, onChange, suffix }) {
  const [text, setText] = useState(null);
  return (
    <Field
      label={label}
      suffix={suffix}
      value={text ?? decimals(value ?? 0, digits)}
      onBlur={() => setText(null)}
      onChange={(v) => {
        setText(v);
        onChange(Number(v) || 0);
      }}
    />
  );
}

function NumField({ label, value, onChange, suffix, readOnly }) {
  return (
    <Field
      label={label}
      value={value}
      suffix={suffix}
      readOnly={readOnly}
      onChange={onChange ? (v) => onChange(v === "" ? 0 : Number(v) || 0) : undefined}
    />
  );
}

function SelectField({ label, value, names, order, options, onChange, disabled }) {
  return (
    <>
      <label>{label}</label>
      <PropSelect
        fill
        disabled={disabled}
        value={value ?? 0}
        options={options ?? enumOptions(names, order)}
        onChange={onChange}
      />
    </>
  );
}

function FlagCheck({ flags, bit, label, onChange, disabled }) {
  return (
    <label className={`sym-check${disabled ? " grp-check-disabled" : ""}`}>
      <input
        type="checkbox"
        disabled={disabled}
        checked={((flags ?? 0) & bit) !== 0}
        onChange={() => onChange((flags ?? 0) ^ bit)}
      />{" "}
      {label}
    </label>
  );
}

/** The reference right-aligns the caption and puts the box after it in the two-column blocks. */
function TrailingCheck({ flags, bit, label, onChange }) {
  return (
    <label className="sym-check grp-check-trailing">
      <span>{label}</span>
      <input
        type="checkbox"
        checked={((flags ?? 0) & bit) !== 0}
        onChange={() => onChange((flags ?? 0) ^ bit)}
      />
    </label>
  );
}

export function GroupCommonTab({ g, set }) {
  const notify = (g.permission_flags ?? 0) & PermissionFlag_notifyMask;
  const setNotify = (v) =>
    set("permission_flags", ((g.permission_flags ?? 0) & ~PermissionFlag_notifyMask) | v);

  return (
    <>
      <GroupTabIntro>
        Group is a set of users that have the same permission settings and service conditions.
        Please specify name of group, deposit currency, trade server, and authentication type.
      </GroupTabIntro>
      <div className="form-grid sym-form-two-col">
        <Field label="Name" value={g.group} onChange={(v) => set("group", v)} />
        <SelectField
          label="Currency"
          value={g.currency || "USD"}
          options={Currency_options}
          onChange={(v) => set("currency", v)}
        />
        <SelectField label="Trade server" value="main" options={[{ value: "main", label: "Trade Main, 1" }]} disabled />
        <SelectField
          label="Digits"
          value={g.currency_digits ?? 2}
          options={[0, 1, 2, 3, 4, 5, 6, 7, 8]}
          onChange={(v) => set("currency_digits", Number(v))}
        />
        <SelectField label="Authentication" value={g.auth_mode} names={AuthMode_name} onChange={(v) => set("auth_mode", v)} />
        <NumField label="Minimum password length" value={g.auth_password_min} onChange={(v) => set("auth_password_min", v)} />
      </div>
      <div className="form-grid">
        <label>Push notifications</label>
        <span className="grp-suffixed">
          <span className="grp-notify-select">
            <FlagCombo
              hideLabel
              emptyLabel=""
              labels={NotifyFlag_labels}
              value={notify}
              onChange={setNotify}
            />
          </span>
          <span className="grp-suffix">sent from the trade server</span>
        </span>
      </div>
      <div className="form-grid grp-check-stack">
        {PermissionFlag_labels.map(({ bit, label }) => (
          <FlagCheck
            key={bit}
            flags={g.permission_flags}
            bit={bit}
            label={label}
            onChange={(v) => set("permission_flags", v)}
          />
        ))}
        <FlagCheck
          flags={g.permission_flags}
          bit={PermissionFlag_forceOtp}
          label="Force one-time password usage"
          onChange={(v) => set("permission_flags", v)}
        />

      </div>
    </>
  );
}

export function GroupCompanyTab({ g, set }) {
  return (
    <>
      <GroupTabIntro>
        Please specify details of the group servicing company and the folder with mail and report
        templates.
      </GroupTabIntro>
      <div className="form-grid">
        <Field label="Company" value={g.company} onChange={(v) => set("company", v)} wide />
        <Field label="Company site" value={g.company_page} onChange={(v) => set("company_page", v)} wide />
        <Field label="Company email" value={g.company_email} onChange={(v) => set("company_email", v)} wide />
        <Field label="Deposit site" value={g.company_deposit} onChange={(v) => set("company_deposit", v)} wide />
        <Field label="Withdrawal site" value={g.company_withdrawal} onChange={(v) => set("company_withdrawal", v)} wide />
        <Field label="Support site" value={g.company_support_page} onChange={(v) => set("company_support_page", v)} wide />
        <Field label="Support email" value={g.company_support_email} onChange={(v) => set("company_support_email", v)} wide />
        <Field label="Templates folder" value={g.company_catalog} onChange={(v) => set("company_catalog", v)} wide />
      </div>
    </>
  );
}

export function GroupNewsMailTab({ g, set }) {
  const newsOff = (g.news_mode ?? 0) === 0;
  const [langsOpen, setLangsOpen] = useState(false);

  return (
    <>
      <GroupTabIntro>
        Please specify the settings of news received by the group and the possibility of using the
        mail system.
      </GroupTabIntro>
      <div className="form-grid">
        <SelectField
          label="News"
          value={g.news_mode}
          names={NewsMode_name}
          onChange={(v) => set("news_mode", v)}
        />
        <Field
          label="News categories"
          value={g.news_category ?? ""}
          disabled={newsOff}
          onChange={(v) => set("news_category", v)}
          wide
        />
        <label className={newsOff ? "grp-check-disabled" : undefined}>News languages</label>
        <span className="grp-suffixed grp-lang-picker">
          <input
            type="text"
            readOnly
            disabled={newsOff}
            value={newsLangSummary(g.news_langs)}
          />
          <button
            type="button"
            className="grp-change-btn"
            disabled={newsOff}
            onClick={() => setLangsOpen(true)}
          >
            Change
          </button>
        </span>
      </div>
      <div className="form-grid grp-check-stack">
        <label className="sym-check">
          <input
            type="checkbox"
            checked={(g.mail_mode ?? 0) === 1}
            onChange={(e) => set("mail_mode", e.target.checked ? 1 : 0)}
          />{" "}
          Enable internal mail system
        </label>
      </div>
      {langsOpen && (
        <NewsLanguagesDialog
          selected={g.news_langs ?? []}
          onClose={() => setLangsOpen(false)}
          onChange={(ids) => set("news_langs", ids)}
        />
      )}
    </>
  );
}

// Blank means "the trader chooses", which is absent on the wire — not zero.
const unset = (v) => (v.trim() === "" ? undefined : Number(v) || 0);

export function GroupPermissionsTab({ g, set }) {
  const isDemo = groupKind(g.group ?? "") === "demo";
  const signals = (g.trade_flags ?? 0) & TradeFlag_signalsMask;
  const setSignals = (v) => set("trade_flags", ((g.trade_flags ?? 0) & ~TradeFlag_signalsMask) | v);

  return (
    <>
      <GroupTabIntro>
        Please specify the group permissions for symbols, orders, use of Expert Advisors, etc.
      </GroupTabIntro>
      <div className="form-grid sym-form-two-col grp-permissions-grid">
        <SelectField
          label="Maximum symbols"
          value={g.limit_symbols ?? 0}
          options={limitOptions(g.limit_symbols)}
          onChange={(v) => set("limit_symbols", v)}
        />
        <SelectField
          label="Available history"
          value={g.limit_history}
          names={HistoryLimit_name}
          onChange={(v) => set("limit_history", v)}
        />
        <SelectField
          label="Maximum positions"
          value={g.limit_positions ?? 0}
          options={limitOptions(g.limit_positions)}
          onChange={(v) => set("limit_positions", v)}
        />
        <SelectField
          label="Maximum orders"
          value={g.limit_orders ?? 0}
          options={limitOptions(g.limit_orders)}
          onChange={(v) => set("limit_orders", v)}
        />
        <Field
          label="Deposit by default"
          value={g.demo_deposit ?? ""}
          readOnly={!isDemo}
          onChange={isDemo ? (v) => set("demo_deposit", unset(v)) : undefined}
        />
        <SelectField
          label="Leverage by default"
          value={g.demo_leverage ?? ""}
          disabled={!isDemo}
          options={demoLeverageOptions(g.demo_leverage)}
          onChange={isDemo ? (v) => set("demo_leverage", v === "" ? undefined : v) : undefined}
        />
        <DecField
          label="Annual interest rate"
          value={g.trade_interest_rate}
          digits={4}
          suffix="%"
          onChange={(v) => set("trade_interest_rate", v)}
        />
      </div>
      <div className="form-grid grp-permissions-wide">
        <SelectField label="Trading Signals" value={signals} options={SignalsMode_options} onChange={setSignals} />
        <SelectField label="Transfer of funds" value={g.trade_transfer_mode} names={TransferMode_name} onChange={(v) => set("trade_transfer_mode", v)} />
      </div>
      <div className="form-grid grp-check-cols grp-permissions-checks">
        {TradeFlag_labels.map(({ bit, label }) => (
          <TrailingCheck key={bit} flags={g.trade_flags} bit={bit} label={label} onChange={(v) => set("trade_flags", v)} />
        ))}
      </div>
    </>
  );
}

export function GroupMarginTab({ g, set }) {
  const compensates = ((g.trade_flags ?? 0) & TradeFlag_soCompensation) !== 0;
  const [profiles, setProfiles] = useState([]);

  useEffect(() => {
    fetchLeverages().then((res) => setProfiles(res.ok ? (res.data ?? []) : []));
  }, []);

  return (
    <>
      <GroupTabIntro>
        Please set up the mechanism of margin calculation for the group: choose the method of
        calculation and margin requirements.
      </GroupTabIntro>
      <div className="form-grid">
        <SelectField label="Risk management" value={g.margin_mode} names={MarginMode_name} order={MarginMode_order} onChange={(v) => set("margin_mode", v)} />
      </div>
      <div className="form-grid grp-margin-levels">
        <DecField label="Margin call level" value={g.margin_call} digits={2} onChange={(v) => set("margin_call", v)} />
        <DecField label="Stop out level" value={g.margin_stop_out} digits={2} onChange={(v) => set("margin_stop_out", v)} />
        <label>in</label>
        <PropSelect fill value={g.margin_so_mode ?? 0} options={enumOptions(StopOutMode_name)} onChange={(v) => set("margin_so_mode", v)} />
      </div>
      <div className="form-grid grp-check-stack">
        <FlagCheck
          flags={g.trade_flags}
          bit={TradeFlag_soFullyHedged}
          label="Stop out fully hedged accounts"
          onChange={(v) => set("trade_flags", v)}
        />
        <FlagCheck
          flags={g.trade_flags}
          bit={TradeFlag_soCompensation}
          label="Compensate negative balance after stop out"
          onChange={(v) => set("trade_flags", v)}
        />
        <FlagCheck
          flags={g.trade_flags}
          bit={TradeFlag_soCompensationCredit}
          label="Withdraw credit after negative balance compensation"
          disabled={!compensates}
          onChange={(v) => set("trade_flags", v)}
        />
        <FlagCheck
          flags={g.margin_flags}
          bit={MarginFlag_clearAcc}
          label="Clear accumulated values at the end of day"
          onChange={(v) => set("margin_flags", v)}
        />
      </div>
      <div className="form-grid">
        <SelectField
          label="Floating leverage profile"
          value={g.margin_leverage_id ?? ""}
          options={[
            { value: "", label: "None" },
            ...profiles.map((p) => ({ value: p.leverage_id, label: p.name })),
          ]}
          onChange={(v) => set("margin_leverage_id", v === "" ? undefined : Number(v))}
        />
      </div>
      <fieldset className="fieldset">
        <legend>Profit/loss in free margin</legend>
        <div className="form-grid">
          <SelectField label="Unrealized profit" value={g.margin_free_mode} names={FreeMarginMode_name} onChange={(v) => set("margin_free_mode", v)} />
          <SelectField label="Daily fixed profit" value={g.margin_free_profit_mode} names={MarginFreeProfitMode_name} onChange={(v) => set("margin_free_profit_mode", v)} />
        </div>
      </fieldset>
      <div className="form-grid">
        <NumField label="Virtual credit" value={g.trade_virtual_credit} suffix="(applies only to opening new positions)" onChange={(v) => set("trade_virtual_credit", v)} />
      </div>
    </>
  );
}

export function GroupReportsTab({ g, set }) {
  const disabled = (g.reports_mode ?? 0) === 0;
  const flags = g.reports_flags ?? 0;
  const [mailServers, setMailServers] = useState(null);

  useEffect(() => {
    fetchMailServers().then((res) => {
      if (res.ok) setMailServers(res.data || []);
    });
  }, []);

  const toggle = (bit) => set("reports_flags", flags ^ bit);
  const check = (bit, label) => (
    <label className={`sym-check${disabled ? " grp-check-disabled" : ""}`}>
      <input type="checkbox" disabled={disabled} checked={(flags & bit) !== 0} onChange={() => toggle(bit)} />{" "}
      {label}
    </label>
  );

  // MT5 stores the configured mail server name in ReportsEmail; ReportsSMTP is obsolete.
  const mailServer = g.reports_email || g.reports_smtp || "";
  const serverOptions = [{ value: "", label: "Default" }];
  for (const row of mailServers || []) {
    if (row.enabled !== false) serverOptions.push({ value: row.name, label: row.name });
  }
  if (mailServer && !serverOptions.some((o) => o.value === mailServer)) {
    serverOptions.push({ value: mailServer, label: mailServer });
  }

  function setReportsMode(mode) {
    set("reports_mode", mode);
    if (mode === 0) set("reports_flags", 0);
  }

  function setMailServer(name) {
    set("reports_email", name);
    if (g.reports_smtp) set("reports_smtp", "");
  }

  return (
    <>
      <GroupTabIntro>
        The platform can daily save the end-of-day state of accounts to a special database. That
        data is used for generating daily statements and various reports for managers.
      </GroupTabIntro>
      <div className="form-grid">
        <SelectField
          label="Generate report data"
          value={g.reports_mode}
          names={ReportsMode_name}
          onChange={setReportsMode}
        />
      </div>
      <div className="form-grid grp-check-stack">
        {check(ReportsFlag_statements, "Generate statements for clients")}
        {check(ReportsFlag_email, "Send statements by email")}
      </div>
      <div className="form-grid">
        <SelectField
          label="Mail server"
          value={mailServer}
          disabled={disabled || !(flags & ReportsFlag_email)}
          options={serverOptions}
          onChange={setMailServer}
        />
      </div>
      <div className="form-grid grp-check-stack">{check(ReportsFlag_support, "Send copies to support email")}</div>
    </>
  );
}

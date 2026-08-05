import { Icon } from "@/components/ui/Icon.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import {
  AuthMode_name,
  FreeMarginMode_name,
  HistoryLimit_name,
  MailMode_name,
  MarginFreeProfitMode_name,
  MarginMode_name,
  NewsMode_name,
  PermissionFlag_forceOtp,
  PermissionFlag_labels,
  PermissionFlag_notify,
  ReportsFlag_labels,
  ReportsMode_name,
  StopOutMode_name,
  TradeFlag_labels,
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

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

function Field({ label, value, onChange, wide, readOnly, suffix }) {
  return (
    <>
      <label>{label}</label>
      <span className={suffix ? "grp-suffixed" : undefined}>
        <input
          type="text"
          readOnly={readOnly || !onChange}
          value={value ?? ""}
          className={wide ? "wide" : ""}
          onChange={onChange ? (e) => onChange(e.target.value) : undefined}
        />
        {suffix && <span className="grp-suffix">{suffix}</span>}
      </span>
    </>
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

function SelectField({ label, value, names, onChange }) {
  return (
    <>
      <label>{label}</label>
      <PropSelect fill value={value ?? 0} options={enumOptions(names)} onChange={onChange} />
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

export function GroupCommonTab({ g, set, isNew }) {
  return (
    <>
      <GroupTabIntro>
        Group is a set of users that have the same permission settings and service conditions.
        Please specify name of group, deposit currency, trade server, and authentication type.
      </GroupTabIntro>
      <div className="form-grid sym-form-two-col">
        <Field label="Name" value={g.group} onChange={isNew ? (v) => set("group", v) : undefined} wide />
        <Field label="Currency" value={g.currency} onChange={(v) => set("currency", v.toUpperCase())} />
        <Field label="Trade server" value="Trade Server (Live)" readOnly />
        <NumField label="Digits" value={g.currency_digits} onChange={(v) => set("currency_digits", v)} />
        <SelectField label="Authentication" value={g.auth_mode} names={AuthMode_name} onChange={(v) => set("auth_mode", v)} />
        <NumField label="Minimum password length" value={g.auth_password_min} onChange={(v) => set("auth_password_min", v)} />
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
      <fieldset className="fieldset">
        <legend>Push notifications sent from the trade server</legend>
        {PermissionFlag_notify.map(({ bit, label }) => (
          <FlagCheck
            key={bit}
            flags={g.permission_flags}
            bit={bit}
            label={label}
            onChange={(v) => set("permission_flags", v)}
          />
        ))}
      </fieldset>
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
        <Field label="Support site" value={g.company_support_page} onChange={(v) => set("company_support_page", v)} wide />
        <Field label="Support email" value={g.company_support_email} onChange={(v) => set("company_support_email", v)} wide />
        <Field label="Templates folder" value={g.company_catalog} onChange={(v) => set("company_catalog", v)} wide />
      </div>
    </>
  );
}

export function GroupNewsMailTab({ g, set }) {
  return (
    <>
      <GroupTabIntro>
        Please specify the settings of news received by the group and the possibility of using the
        mail system.
      </GroupTabIntro>
      <div className="form-grid">
        <SelectField label="News" value={g.news_mode} names={NewsMode_name} onChange={(v) => set("news_mode", v)} />
        <Field label="News categories" value={g.news_category} onChange={(v) => set("news_category", v)} wide />
        <SelectField label="Internal mail" value={g.mail_mode} names={MailMode_name} onChange={(v) => set("mail_mode", v)} />
      </div>
    </>
  );
}

const unlimited = (v) => (v === 0 || v == null ? "unlimited" : String(v));
const fromUnlimited = (v) => (v === "unlimited" || v === "" ? 0 : Number(v) || 0);

export function GroupPermissionsTab({ g, set }) {
  return (
    <>
      <GroupTabIntro>
        Please specify the group permissions for symbols, orders, use of Expert Advisors, etc.
      </GroupTabIntro>
      <div className="form-grid sym-form-two-col">
        <Field label="Maximum symbols" value={unlimited(g.limit_symbols)} onChange={(v) => set("limit_symbols", fromUnlimited(v))} />
        <SelectField label="Available history" value={g.limit_history} names={HistoryLimit_name} onChange={(v) => set("limit_history", v)} />
        <Field label="Maximum positions" value={unlimited(g.limit_positions)} onChange={(v) => set("limit_positions", fromUnlimited(v))} />
        <Field label="Maximum orders" value={unlimited(g.limit_orders)} onChange={(v) => set("limit_orders", fromUnlimited(v))} />
        <NumField label="Deposit by default" value={g.demo_deposit ?? 0} onChange={(v) => set("demo_deposit", v)} />
        <NumField label="Leverage by default" value={g.demo_leverage ?? 0} onChange={(v) => set("demo_leverage", v)} />
        <NumField label="Annual interest rate" value={g.trade_interest_rate} suffix="%" onChange={(v) => set("trade_interest_rate", v)} />
        <SelectField label="Transfer of funds" value={g.trade_transfer_mode} names={TransferMode_name} onChange={(v) => set("trade_transfer_mode", v)} />
      </div>
      <div className="form-grid grp-check-stack">
        {TradeFlag_labels.map(({ bit, label }) => (
          <FlagCheck key={bit} flags={g.trade_flags} bit={bit} label={label} onChange={(v) => set("trade_flags", v)} />
        ))}
      </div>
    </>
  );
}

export function GroupMarginTab({ g, set }) {
  const compensates = ((g.trade_flags ?? 0) & TradeFlag_soCompensation) !== 0;
  return (
    <>
      <GroupTabIntro>
        Please set up the mechanism of margin calculation for the group: choose the method of
        calculation and margin requirements.
      </GroupTabIntro>
      <div className="form-grid">
        <SelectField label="Risk management" value={g.margin_mode} names={MarginMode_name} onChange={(v) => set("margin_mode", v)} />
        <NumField label="Margin call level" value={g.margin_call} onChange={(v) => set("margin_call", v)} />
        <NumField label="Stop out level" value={g.margin_stop_out} onChange={(v) => set("margin_stop_out", v)} />
        <SelectField label="in" value={g.margin_so_mode} names={StopOutMode_name} onChange={(v) => set("margin_so_mode", v)} />
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
          label="Withdraw credit when compensating negative balance"
          disabled={!compensates}
          onChange={(v) => set("trade_flags", v)}
        />
      </div>
      <fieldset className="fieldset">
        <legend>Profit/loss in free margin</legend>
        <div className="form-grid">
          <SelectField label="Unrealized profit" value={g.margin_free_mode} names={FreeMarginMode_name} onChange={(v) => set("margin_free_mode", v)} />
          <SelectField label="Daily fixed profit" value={g.margin_free_profit_mode} names={MarginFreeProfitMode_name} onChange={(v) => set("margin_free_profit_mode", v)} />
          <NumField label="Virtual credit" value={g.trade_virtual_credit} onChange={(v) => set("trade_virtual_credit", v)} />
        </div>
      </fieldset>
    </>
  );
}

export function GroupReportsTab({ g, set }) {
  const disabled = (g.reports_mode ?? 0) === 0;
  return (
    <>
      <GroupTabIntro>
        Please specify the settings of daily and monthly reports generated for the group.
      </GroupTabIntro>
      <div className="form-grid">
        <SelectField label="Generate report data" value={g.reports_mode} names={ReportsMode_name} onChange={(v) => set("reports_mode", v)} />
        <Field label="Mail server" value={g.reports_smtp} onChange={disabled ? undefined : (v) => set("reports_smtp", v)} readOnly={disabled} />
        <Field label="SMTP login" value={g.reports_smtp_login} onChange={disabled ? undefined : (v) => set("reports_smtp_login", v)} readOnly={disabled} />
        <Field label="Reports email" value={g.reports_email} onChange={disabled ? undefined : (v) => set("reports_email", v)} readOnly={disabled} />
      </div>
      <div className="form-grid grp-check-stack">
        {ReportsFlag_labels.map(({ bit, label }) => (
          <label key={bit} className={`sym-check${disabled ? " grp-check-disabled" : ""}`}>
            <input
              type="checkbox"
              disabled={disabled}
              checked={((g.reports_flags ?? 0) & bit) !== 0}
              onChange={() => set("reports_flags", (g.reports_flags ?? 0) ^ bit)}
            />{" "}
            {label}
          </label>
        ))}
      </div>
    </>
  );
}

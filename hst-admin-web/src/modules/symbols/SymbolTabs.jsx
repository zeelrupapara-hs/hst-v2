import { Icon } from "@/components/ui/Icon.jsx";
import { PropSelect } from "@/components/ui/PropSelect.jsx";
import {
  CalcMode_name,
  ExecMode_name,
  ExpirFlag_labels,
  FillFlag_labels,
  GTCMode_name,
  OrderFlag_labels,
  SwapDays_options,
  SwapMode_name,
  SymbolSector_name,
  TradeMode_name,
  toLots,
  fromLots,
} from "@/constants/symbols.js";
export function TabIntro({ icon = "symbols-tree", children }) {
  return (
    <div className="sym-sessions-intro">
      <span className="sym-tab-intro-icon" aria-hidden="true">
        <Icon id={icon} size={48} />
      </span>
      <p>{children}</p>
    </div>
  );
}

const enumOptions = (names) =>
  Object.entries(names).map(([value, label]) => ({ value: Number(value), label }));

function Field({ label, value, onChange, wide, readOnly }) {
  return (
    <>
      <label>{label}</label>
      <input
        type="text"
        readOnly={readOnly || !onChange}
        value={value ?? ""}
        className={wide ? "wide" : ""}
        onChange={onChange ? (e) => onChange(e.target.value) : undefined}
      />
    </>
  );
}

function NumField({ label, value, onChange, readOnly }) {
  return (
    <Field
      label={label}
      value={value}
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

function FlagChecks({ labels, value, onChange }) {
  return labels.map(({ bit, label }) => (
    <label key={bit} className="sym-check">
      <input
        type="checkbox"
        checked={((value ?? 0) & bit) !== 0}
        onChange={(e) => onChange((value ?? 0) ^ bit, e.target.checked)}
      />{" "}
      {label}
    </label>
  ));
}

export function CommonTab({ s, set, isNew }) {
  return (
    <>
      <TabIntro>
        The setting up of main parameters of the symbol. Please specify its name, description, and
        other parameters.
      </TabIntro>
      <div className="form-grid sym-form-two-col">
        <Field label="Symbol" value={s.symbol} onChange={isNew ? (v) => set("symbol", v) : undefined} />
        <Field label="Description" value={s.description} onChange={(v) => set("description", v)} wide />
        <Field label="Exchange" value={s.exchange} onChange={(v) => set("exchange", v)} />
        <Field label="International" value={s.international} onChange={(v) => set("international", v)} />
        <Field label="ISIN" value={s.isin} onChange={(v) => set("isin", v)} />
        <Field label="Sector" value={SymbolSector_name[s.sector] ?? "Undefined"} />
        <Field label="CFI" value={s.cfi} onChange={(v) => set("cfi", v)} />
        <Field label="Country" value={s.country} onChange={(v) => set("country", v)} />
        <Field label="Basis" value={s.basis} onChange={(v) => set("basis", v)} />
        <Field label="Category" value={s.category} onChange={(v) => set("category", v)} />
        <Field label="Source" value={s.source} onChange={(v) => set("source", v)} />
        <Field label="Page" value={s.page} onChange={(v) => set("page", v)} wide />
        <NumField label="Digits" value={s.digits} onChange={(v) => set("digits", v)} />
        <Field
          label="Market depth"
          value={s.tick_book_depth ? s.tick_book_depth : "off"}
          onChange={(v) => set("tick_book_depth", v === "off" ? 0 : Number(v) || 0)}
        />
        <Field
          label="Spread"
          value={s.spread === 0 ? "off" : s.spread}
          onChange={(v) => set("spread", v === "off" ? 0 : Number(v) || 0)}
        />
        <NumField label="Spread balance" value={s.spread_balance} onChange={(v) => set("spread_balance", v)} />
      </div>
    </>
  );
}

export function CurrencyTab({ s, set }) {
  return (
    <>
      <TabIntro>The setting up of base, profit, and margin currencies of the symbol.</TabIntro>
      <div className="form-grid">
        <Field label="Base currency" value={s.currency_base} onChange={(v) => set("currency_base", v.toUpperCase())} />
        <NumField label="Base currency digits" value={s.currency_base_digits} onChange={(v) => set("currency_base_digits", v)} />
        <Field label="Profit currency" value={s.currency_profit} onChange={(v) => set("currency_profit", v.toUpperCase())} />
        <NumField label="Profit currency digits" value={s.currency_profit_digits} onChange={(v) => set("currency_profit_digits", v)} />
        <Field label="Margin currency" value={s.currency_margin} onChange={(v) => set("currency_margin", v.toUpperCase())} />
        <NumField label="Margin currency digits" value={s.currency_margin_digits} onChange={(v) => set("currency_margin_digits", v)} />
      </div>
    </>
  );
}

export function QuotesTab({ s, set }) {
  return (
    <>
      <TabIntro>
        The setting up of different filtration levels is intended for cutting off the incorrect
        data.
      </TabIntro>
      <div className="form-grid sym-quotes-grid">
        <NumField label="Soft filtration level" value={s.filter_soft} onChange={(v) => set("filter_soft", v)} />
        <NumField label="Soft filter ticks" value={s.filter_soft_ticks} onChange={(v) => set("filter_soft_ticks", v)} />
        <NumField label="Hard filtration level" value={s.filter_hard} onChange={(v) => set("filter_hard", v)} />
        <NumField label="Hard filter ticks" value={s.filter_hard_ticks} onChange={(v) => set("filter_hard_ticks", v)} />
        <NumField label="Discard level" value={s.filter_discard} onChange={(v) => set("filter_discard", v)} />
        <NumField label="Gap level" value={s.filter_gap} onChange={(v) => set("filter_gap", v)} />
        <NumField label="Gap level ticks" value={s.filter_gap_ticks} onChange={(v) => set("filter_gap_ticks", v)} />
        <NumField label="Minimum spread" value={s.filter_spread_min} onChange={(v) => set("filter_spread_min", v)} />
        <NumField label="Maximum spread" value={s.filter_spread_max} onChange={(v) => set("filter_spread_max", v)} />
        <NumField label="Delay for subscriptions" value={s.subscriptions_delay} onChange={(v) => set("subscriptions_delay", v)} />
      </div>
    </>
  );
}

export function TradeTab({ s, set }) {
  return (
    <>
      <TabIntro>
        The setting up of symbol trading. Please specify the type of profit calculation, type of
        placed orders, allowed trade volumes, etc.
      </TabIntro>
      <div className="form-grid sym-form-two-col">
        <NumField label="Contract size" value={s.contract_size} onChange={(v) => set("contract_size", v)} />
        <NumField label="Limit & stop level" value={s.stops_level} onChange={(v) => set("stops_level", v)} />
        <SelectField label="Calculation" value={s.calc_mode} names={CalcMode_name} onChange={(v) => set("calc_mode", v)} />
        <NumField label="Freeze level" value={s.freeze_level} onChange={(v) => set("freeze_level", v)} />
        <SelectField label="Trade" value={s.trade_mode} names={TradeMode_name} onChange={(v) => set("trade_mode", v)} />
        <NumField label="Max quote delay" value={s.quotes_timeout} onChange={(v) => set("quotes_timeout", v)} />
        <SelectField label="GTC mode" value={s.gtc_mode} names={GTCMode_name} onChange={(v) => set("gtc_mode", v)} />
        <NumField label="Tick size" value={s.tick_size} onChange={(v) => set("tick_size", v)} />
        <NumField label="Tick value" value={s.tick_value} onChange={(v) => set("tick_value", v)} />
      </div>
      <fieldset className="fieldset sym-volumes">
        <legend>Volumes</legend>
        <div className="form-grid">
          <NumField label="Minimum" value={toLots(s.volume_min)} onChange={(v) => set("volume_min", fromLots(v))} />
          <NumField label="Step" value={toLots(s.volume_step)} onChange={(v) => set("volume_step", fromLots(v))} />
          <NumField label="Maximum" value={toLots(s.volume_max)} onChange={(v) => set("volume_max", fromLots(v))} />
          <NumField label="Limit" value={toLots(s.volume_limit)} onChange={(v) => set("volume_limit", fromLots(v))} />
        </div>
      </fieldset>
      <fieldset className="fieldset">
        <legend>Filling</legend>
        <FlagChecks labels={FillFlag_labels} value={s.fill_flags} onChange={(v) => set("fill_flags", v)} />
      </fieldset>
      <fieldset className="fieldset">
        <legend>Expiration</legend>
        <FlagChecks labels={ExpirFlag_labels} value={s.expir_flags} onChange={(v) => set("expir_flags", v)} />
      </fieldset>
      <fieldset className="fieldset">
        <legend>Orders</legend>
        <FlagChecks labels={OrderFlag_labels} value={s.order_flags} onChange={(v) => set("order_flags", v)} />
      </fieldset>
    </>
  );
}

export function ExecutionTab({ s, set }) {
  return (
    <>
      <TabIntro>
        The setting up of execution of orders by the symbol. Please specify the execution type and
        its parameters.
      </TabIntro>
      <div className="form-grid">
        <SelectField label="Execution" value={s.exec_mode} names={ExecMode_name} onChange={(v) => set("exec_mode", v)} />
        <NumField label="Timeout" value={s.ie_timeout} onChange={(v) => set("ie_timeout", v)} />
        <NumField label="Profit slippage" value={s.ie_slip_profit} onChange={(v) => set("ie_slip_profit", v)} />
        <NumField label="Losing slippage" value={s.ie_slip_losing} onChange={(v) => set("ie_slip_losing", v)} />
        <NumField label="Max volume" value={toLots(s.ie_volume_max)} onChange={(v) => set("ie_volume_max", fromLots(v))} />
      </div>
    </>
  );
}

export function MarginTab({ s, set }) {
  return (
    <>
      <TabIntro>The setting up of margin requirements for the symbol.</TabIntro>
      <div className="form-grid">
        <NumField label="Initial margin" value={s.margin_initial} onChange={(v) => set("margin_initial", v)} />
        <NumField label="Maintenance margin" value={s.margin_maintenance} onChange={(v) => set("margin_maintenance", v)} />
        <NumField label="Hedged margin" value={s.margin_hedged} onChange={(v) => set("margin_hedged", v)} />
      </div>
    </>
  );
}

const RATE_ROWS = [
  ["Initial Buy", "margin_initial_buy", "margin_initial_buy_limit", "margin_initial_buy_stop", "margin_initial_buy_stop_limit"],
  ["Initial Sell", "margin_initial_sell", "margin_initial_sell_limit", "margin_initial_sell_stop", "margin_initial_sell_stop_limit"],
  ["Maintenance Buy", "margin_maintenance_buy", "margin_maintenance_buy_limit", "margin_maintenance_buy_stop", "margin_maintenance_buy_stop_limit"],
  ["Maintenance Sell", "margin_maintenance_sell", "margin_maintenance_sell_limit", "margin_maintenance_sell_stop", "margin_maintenance_sell_stop_limit"],
];

export function MarginRatesTab({ s, set }) {
  return (
    <>
      <TabIntro>Please specify margin rates for different types of trade operations.</TabIntro>
      <div className="form-grid">
        <NumField label="Liquidity margin rate" value={s.margin_rate_liquidity} onChange={(v) => set("margin_rate_liquidity", v)} />
        <NumField label="Currency margin rate" value={s.margin_rate_currency} onChange={(v) => set("margin_rate_currency", v)} />
      </div>
      <table className="sym-margin-rates-table">
        <thead>
          <tr>
            <th />
            <th>Market Order</th>
            <th>Limit Order</th>
            <th>Stop Order</th>
            <th>Stop Limit Order</th>
          </tr>
        </thead>
        <tbody>
          {RATE_ROWS.map(([label, ...keys]) => (
            <tr key={label}>
              <td>{label}</td>
              {keys.map((key) => (
                <td key={key}>
                  <input
                    type="text"
                    className="df-cell-input"
                    value={s[key] ?? 0}
                    onChange={(e) => set(key, Number(e.target.value) || 0)}
                  />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}

const SWAP_RATE_KEYS = [
  ["Sunday", "swap_rate_sunday"],
  ["Monday", "swap_rate_monday"],
  ["Tuesday", "swap_rate_tuesday"],
  ["Wednesday", "swap_rate_wednesday"],
  ["Thursday", "swap_rate_thursday"],
  ["Friday", "swap_rate_friday"],
  ["Saturday", "swap_rate_saturday"],
];

export function SwapsTab({ s, set }) {
  return (
    <>
      <TabIntro>The setting up of charging swaps by the symbol.</TabIntro>
      <div className="form-grid">
        <SelectField label="Type" value={s.swap_mode} names={SwapMode_name} onChange={(v) => set("swap_mode", v)} />
        <NumField label="Long positions" value={s.swap_long} onChange={(v) => set("swap_long", v)} />
        <NumField label="Short positions" value={s.swap_short} onChange={(v) => set("swap_short", v)} />
        <SelectField
          label="Days in year"
          value={s.swap_year_day}
          names={Object.fromEntries(SwapDays_options.map((d) => [d, String(d)]))}
          onChange={(v) => set("swap_year_day", v)}
        />
      </div>
      <table className="sym-sessions-table sym-swaps-table">
        <thead>
          <tr>
            <th>Day of week</th>
            <th>Multiplier</th>
          </tr>
        </thead>
        <tbody>
          {SWAP_RATE_KEYS.map(([name, key]) => (
            <tr key={key}>
              <td>{name}</td>
              <td>
                <input
                  type="text"
                  className="df-cell-input"
                  value={s[key] ?? 0}
                  onChange={(e) => set(key, Number(e.target.value) || 0)}
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}

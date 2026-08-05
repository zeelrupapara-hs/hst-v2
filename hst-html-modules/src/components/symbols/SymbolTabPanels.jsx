import { Icon } from "../ui/Icon.jsx";

export function SymbolTabIntro({ children }) {
  return (
    <div className="sym-sessions-intro">
      <span className="sym-tab-intro-icon" aria-hidden="true">
        <Icon name="symbols-tree" size={48} />
      </span>
      <p>{children}</p>
    </div>
  );
}

function Field({ label, value, wide, readOnly, onChange }) {
  return (
    <>
      <label>{label}</label>
      <input
        type="text"
        readOnly={readOnly}
        value={value ?? ""}
        className={wide ? "wide" : ""}
        onChange={onChange ? (e) => onChange(e.target.value) : undefined}
      />
    </>
  );
}

export function SymbolCommonTab({ s, onFieldChange, isNew }) {
  if (!s) return null;
  const canEdit = !!onFieldChange;
  return (
    <>
      <SymbolTabIntro>
        The common symbol parameters are set up on this tab.
      </SymbolTabIntro>
      <div className="form-grid sym-form-two-col">
        <Field
          label="Symbol"
          value={s.symbol}
          readOnly={!canEdit || !isNew}
          onChange={isNew ? (v) => onFieldChange("symbol", v) : undefined}
        />
        <Field
          label="Description"
          value={s.description}
          readOnly={!canEdit}
          onChange={canEdit ? (v) => onFieldChange("description", v) : undefined}
          wide
        />
        <Field label="Exchange" value={s.exchange || ""} readOnly={!canEdit} onChange={canEdit ? (v) => onFieldChange("exchange", v) : undefined} />
        <Field label="International" value={s.international || ""} readOnly />
        <Field label="ISIN" value={s.isin || ""} readOnly />
        <Field label="Sector" value={s.sector === 12 ? "Currency" : "Undefined"} readOnly />
        <Field label="CFI" value={s.cfi || ""} readOnly />
        <Field label="Industry" value="Undefined" readOnly />
        <Field
          label="Digits"
          value={s.digits}
          readOnly={!canEdit}
          onChange={canEdit ? (v) => onFieldChange("digits", Number(v) || 0) : undefined}
        />
        <Field
          label="Spread"
          value={s.spread === 0 ? "off" : s.spread}
          readOnly={!canEdit}
          onChange={canEdit ? (v) => onFieldChange("spread", v === "off" ? 0 : Number(v) || 0) : undefined}
        />
        <Field label="Market depth" value={s.tick_book_depth ? s.tick_book_depth : "off"} readOnly />
      </div>
    </>
  );
}

export function SymbolCurrencyTab({ s }) {
  if (!s) return null;
  return (
    <>
      <SymbolTabIntro>
        The setting up of base, profit, and margin currencies of the symbol.
      </SymbolTabIntro>
      <div className="form-grid">
        <Field label="Base currency" value={s.currency_base} />
        <Field label="Base currency digits" value={s.currency_base_digits} />
        <Field label="Profit currency" value={s.currency_profit} />
        <Field label="Profit currency digits" value={s.currency_profit_digits} />
        <Field label="Margin currency" value={s.currency_margin} />
        <Field label="Margin currency digits" value={s.currency_margin_digits} />
      </div>
    </>
  );
}

export function SymbolQuotesTab({ s }) {
  if (!s) return null;
  return (
    <>
      <SymbolTabIntro>
        The setting up of different filtration levels is intended for cutting off
        the incorrect data.
      </SymbolTabIntro>
      <div className="form-grid sym-quotes-grid">
        <label className="sym-check">
          <input type="checkbox" checked readOnly /> Allow realtime quotes from datafeeds
        </label>
        <label className="sym-check">
          <input type="checkbox" readOnly /> Allow negative quotes
        </label>
        <Field label="Soft filtration level" value={`${s.filter_soft ?? 0} points`} />
        <Field label="Filter (Soft)" value={s.filter_soft_ticks ?? 5} />
        <Field label="Hard filtration level" value={`${s.filter_hard ?? 0} points`} />
        <Field label="Filter (Hard)" value={s.filter_hard_ticks ?? 5} />
        <Field label="Minimum spread" value={s.filter_spread_min ? s.filter_spread_min : "off"} />
        <Field label="Maximum spread" value={s.filter_spread_max ? s.filter_spread_max : "off"} />
        <Field label="Delay for subscriptions" value={`${s.subscriptions_delay ?? 15} min`} />
      </div>
    </>
  );
}

export function SymbolTradeTab({ s }) {
  if (!s) return null;
  const vol = (v) => (v != null ? (v / 10000).toFixed(2) : "—");
  return (
    <>
      <SymbolTabIntro>
        The setting up of symbol trading. Please specify the type of profit
        calculation, type of placed orders, allowed trade volumes, etc.
      </SymbolTabIntro>
      <div className="form-grid sym-form-two-col">
        <Field label="Contract size" value={s.contract_size} />
        <Field label="Limit &amp; stop level" value={`${s.stops_level ?? 0} pt`} />
        <Field label="Calculation" value={s.calc_mode === 0 ? "Forex" : s.calc_mode} />
        <Field label="Freeze level" value={`${s.freeze_level ?? 0} pt`} />
        <Field label="Trade" value={["Disabled", "Long only", "Short only", "Close only", "Full"][s.trade_mode] ?? s.trade_mode} />
        <Field label="Max quote delay" value={s.quotes_timeout ? `${s.quotes_timeout} sec` : "off"} />
        <Field label="GTC" value={s.gtc_mode === 0 ? "Good till cancelled" : s.gtc_mode} />
      </div>
      <fieldset className="fieldset sym-volumes">
        <legend>Volumes</legend>
        <div className="form-grid">
          <Field label="Minimum" value={vol(s.volume_min)} />
          <Field label="Step" value={vol(s.volume_step)} />
          <Field label="Maximum" value={vol(s.volume_max)} />
          <Field label="Limit" value={s.volume_limit ? vol(s.volume_limit) : ""} />
        </div>
      </fieldset>
    </>
  );
}

export function SymbolExecutionTab({ s }) {
  if (!s) return null;
  const modes = ["Request", "Instant", "Market", "Exchange"];
  return (
    <>
      <SymbolTabIntro>
        The setting up of execution of orders by the symbol. Please specify the
        execution type and its parameters.
      </SymbolTabIntro>
      <div className="form-grid">
        <Field label="Execution" value={modes[s.exec_mode] ?? s.exec_mode} />
      </div>
    </>
  );
}

export function SymbolMarginTab({ s }) {
  if (!s) return null;
  return (
    <>
      <SymbolTabIntro>
        The setting up of margin requirements for the symbol.
      </SymbolTabIntro>
      <div className="form-grid">
        <Field label="Initial margin" value={(s.margin_initial ?? 0).toFixed(2)} />
        <Field label="Hedged margin" value={s.margin_hedged ?? 0} />
        <Field label="Maintenance margin" value={(s.margin_maintenance ?? 0).toFixed(2)} />
        <Field label="Additional margin checks" value="None" />
      </div>
    </>
  );
}

export function SymbolMarginRatesTab({ s }) {
  if (!s) return null;
  return (
    <>
      <SymbolTabIntro>
        Please specify margin rates for different types of trade operations.
      </SymbolTabIntro>
      <div className="form-grid">
        <Field label="Liquidity margin rate" value={(s.margin_rate_liquidity ?? 1).toFixed(3)} />
        <Field label="Currency margin rate" value={(s.margin_rate_currency ?? 0).toFixed(8)} />
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
          <tr>
            <td>Initial Buy</td>
            <td>{(s.margin_initial_buy ?? 1).toFixed(7)}</td>
            <td>0.0000000</td>
            <td>0.0000000</td>
            <td>0.0000000</td>
          </tr>
          <tr>
            <td>Initial Sell</td>
            <td>{(s.margin_initial_sell ?? 1).toFixed(7)}</td>
            <td>0.0000000</td>
            <td>0.0000000</td>
            <td>0.0000000</td>
          </tr>
        </tbody>
      </table>
    </>
  );
}

export function SymbolSwapsTab({ s }) {
  if (!s) return null;
  const days = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
  const rates = [
    s.swap_rate_sunday,
    s.swap_rate_monday,
    s.swap_rate_tuesday,
    s.swap_rate_wednesday,
    s.swap_rate_thursday,
    s.swap_rate_friday,
    s.swap_rate_saturday,
  ];
  return (
    <>
      <SymbolTabIntro>
        The setting up of charging swaps by the symbol.
      </SymbolTabIntro>
      <div className="form-grid">
        <label className="sym-check">
          <input type="checkbox" checked={s.swap_mode !== 0} readOnly /> Enable swaps
        </label>
        <Field label="Type" value={s.swap_mode === 1 ? "In points" : "Disabled"} />
        <Field label="Long positions" value={s.swap_long} />
        <Field label="Short positions" value={s.swap_short} />
        <Field label="Days in year" value={s.swap_year_day ?? 360} />
      </div>
      <table className="sym-sessions-table sym-swaps-table">
        <thead>
          <tr>
            <th>Day of week</th>
            <th>Multiplier</th>
          </tr>
        </thead>
        <tbody>
          {days.map((name, i) => (
            <tr key={name}>
              <td>{name}</td>
              <td>{rates[i] != null && rates[i] !== 0 ? rates[i] : ""}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}

export const SYMBOL_TABS = [
  { id: "common", label: "Common", Panel: SymbolCommonTab },
  { id: "currency", label: "Currency", Panel: SymbolCurrencyTab },
  { id: "quotes", label: "Quotes", Panel: SymbolQuotesTab },
  { id: "trade", label: "Trade", Panel: SymbolTradeTab },
  { id: "execution", label: "Execution", Panel: SymbolExecutionTab },
  { id: "margin", label: "Margin", Panel: SymbolMarginTab },
  { id: "margin_rates", label: "Margin Rates", Panel: SymbolMarginRatesTab },
  { id: "swaps", label: "Swaps", Panel: SymbolSwapsTab },
  { id: "sessions", label: "Sessions", Panel: null },
];

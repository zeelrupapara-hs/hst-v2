import { useEffect, useState } from "react";
import { Divider } from "antd";
import { LuX } from "react-icons/lu";
import Icon from "../../common/Icon";
import ModalComponent from "../../modal/ModalComponent";
import Bid from "./Bid";
import Ask from "./Ask";
import Spread from "./Spread";
import { getSymbolSessions } from "../../../api/request/symbol";
import useLiveSymbolStore from "../../../store/useLiveSymbolStore";
import {
  SECTOR,
  CALC_MODE,
  TRADE_MODE,
  CHART_MODE,
  GTC_MODE,
  SWAP_MODE,
  lookup,
  fmtNum,
  formatFillFlags,
  formatExpirFlags,
  formatOrderFlags,
  displayInitialMargin,
} from "../../../utils/symbolSpec";

const DAY_NAMES = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

const SWAP_DAY_LABELS = [
  ["Sunday", "sunday"],
  ["Monday", "monday"],
  ["Tuesday", "tuesday"],
  ["Wednesday", "wednesday"],
  ["Thursday", "thursday"],
  ["Friday", "friday"],
  ["Saturday", "saturday"],
];

const RANGE_MODE_LABEL = {
  0: "volume",
  1: "volume per symbol",
  2: "notional value",
  3: "notional value per symbol",
};

const fmt = (value, digits = 2) => {
  if (value == null || value === "") return "—";
  const n = Number(value);
  if (Number.isNaN(n)) return String(value);
  return n.toFixed(digits);
};

const fmtVolume = (value) => {
  if (value == null || value === "") return "—";
  const n = Number(value);
  if (Number.isNaN(n)) return "—";
  return String(parseFloat(n.toFixed(8)));
};

const fmtTierRange = (from, to) => {
  const start = Number(from) || 0;
  const end = Number(to);
  if (!end) return `${start} – ∞`;
  return `${start} – ${end}`;
};

const formatSessionWindows = (windows = []) => {
  if (!windows.length) return "—";
  return windows
    .map((w) => `${w.from ?? "00:00"} - ${w.to ?? "00:00"}`)
    .join(", ");
};

const emptySessions = () =>
  DAY_NAMES.map((day) => ({
    key: day,
    day,
    quote: "—",
    trade: "—",
  }));

const SpecRow = ({ label, value }) => (
  <div className="flex justify-between gap-4 py-0.5">
    <span className="text-gray-500 shrink-0">{label}</span>
    <span className="text-right tabular-nums">{value ?? "—"}</span>
  </div>
);

const HeaderStat = ({ label, value }) => (
  <div className="text-center min-w-[80px]">
    <div className="text-base font-semibold tabular-nums">{value ?? "—"}</div>
    <div className="text-xs text-gray-500 mt-0.5">{label}</div>
  </div>
);

const SymbolInfoModal = ({ isOpen, setIsOpen, data }) => {
  const [sessions, setSessions] = useState(emptySessions);
  const live = useLiveSymbolStore((state) => state.liveSymbols?.[data?.id]);

  useEffect(() => {
    if (!isOpen || !data?.symbol) return;

    let cancelled = false;

    getSymbolSessions(data.symbol)
      .then((payload) => {
        if (cancelled) return;

        const days = payload?.days ?? [];
        setSessions(
          DAY_NAMES.map((day, idx) => ({
            key: day,
            day,
            quote: formatSessionWindows(days[idx]?.quote),
            trade: formatSessionWindows(days[idx]?.trade),
          }))
        );
      })
      .catch(() => {
        if (!cancelled) setSessions(emptySessions());
      });

    return () => {
      cancelled = true;
    };
  }, [isOpen, data?.symbol]);

  if (!data) return null;

  const calcMode = data.calc_mode ?? data.calculation;
  const tradeMode = data.trade_mode ?? data.trade_level;
  const swapMode = data.swap_mode ?? data.swap_type;
  const marginSpec = data?.margin_spec;
  const sector =
    data.sector ??
    (String(data.path || "").split("\\")[0] === "Forex" ? 12 : undefined);

  const contractSpecLeft = [
    { label: "Margin currency", value: data.currency_margin },
    { label: "Calculation", value: lookup(CALC_MODE, calcMode) },
    { label: "Chart mode", value: lookup(CHART_MODE, data.tick_chart_mode, 0) },
    { label: "GTC mode", value: lookup(GTC_MODE, data.gtc_mode, 0) },
    { label: "Expiration", value: formatExpirFlags(data.expir_flags) },
    { label: "Minimal volume", value: fmtVolume(data.min_value ?? data.volume_min) },
    { label: "Volume step", value: fmtVolume(data.step ?? data.volume_step) },
    { label: "Swap long", value: fmt(data.swap_long, 4) },
  ];

  const contractSpecRight = [
    { label: "Profit currency", value: data.currency_profit },
    { label: "Initial margin", value: displayInitialMargin(data) },
    { label: "Trade", value: lookup(TRADE_MODE, tradeMode) },
    { label: "Filling", value: formatFillFlags(data.fill_flags) },
    { label: "Orders", value: formatOrderFlags(data.order_flags) },
    { label: "Maximal volume", value: fmtVolume(data.max_value ?? data.volume_max) },
    { label: "Swap type", value: lookup(SWAP_MODE, swapMode) },
    { label: "Swap short", value: fmt(data.swap_short, 4) },
  ];

  const symbolInfo = [
    { label: "Bid", value: <Bid symbolId={data?.id} /> },
    { label: "Ask", value: <Ask symbolId={data?.id} /> },
    { label: "Spread", value: <Spread symbolId={data?.id} /> },
    { label: "High", value: fmt(live?.high_bid ?? data?.high_bid, data?.digits) },
    { label: "Low", value: fmt(live?.low_bid ?? data?.low_bid, data?.digits) },
    { label: "Open", value: fmt(data?.open, data?.digits) },
    { label: "Close", value: fmt(data?.close, data?.digits) },
    {
      label: "Change %",
      value:
        live?.percentChange != null
          ? live.percentChange
          : data?.change_percent != null
            ? fmt(data.change_percent, 2)
            : "—",
    },
  ];

  return (
    <ModalComponent isOpen={isOpen} width={860}>
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>{data?.symbol} contract specification</span>

        <button onClick={() => setIsOpen(false)}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-5 max-h-[80vh] overflow-y-auto">
        <div className="pb-4 mb-4 border-b border-theme-border">
          <p className="text-lg font-bold">{data.symbol}</p>
          <p className="text-sm text-gray-500 mt-0.5">{data.description || "—"}</p>

          <div className="flex flex-wrap justify-around gap-6 mt-4 pt-2 pb-4 border-b border-theme-border/60">
            <HeaderStat label="Bid" value={<Bid symbolId={data?.id} />} />
            <HeaderStat label="Ask" value={<Ask symbolId={data?.id} />} />
            <HeaderStat label="Spread" value={<Spread symbolId={data?.id} />} />
          </div>

          <div className="flex flex-wrap justify-around gap-6 mt-4 pt-2">
            <HeaderStat label="Sector" value={lookup(SECTOR, sector)} />
            <HeaderStat label="Digits" value={data.digits} />
            <HeaderStat label="Contract size" value={fmtNum(data.contract_size, 0)} />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-x-16 gap-y-0 text-sm">
          <div className="space-y-0.5">
            {contractSpecLeft.map((row) => (
              <SpecRow key={row.label} {...row} />
            ))}
          </div>
          <div className="space-y-0.5">
            {contractSpecRight.map((row) => (
              <SpecRow key={row.label} {...row} />
            ))}
          </div>
        </div>

        <Divider className="!border-theme-border !my-4" />

        <div className="space-y-2">
          <p className="font-semibold text-sm">Swap rates</p>
          <table className="text-sm border-collapse">
            <thead>
              <tr className="border-b border-theme-border">
                <th className="text-left py-1 pr-8 font-semibold">Day</th>
                <th className="text-right py-1 pr-12 font-semibold w-16">Multiplier</th>
                <th className="text-left py-1 pr-8 font-semibold">Day</th>
                <th className="text-right py-1 font-semibold w-16">Multiplier</th>
              </tr>
            </thead>
            <tbody>
              {SWAP_DAY_LABELS.slice(0, 4).map(([leftLabel, leftKey], i) => {
                const right = SWAP_DAY_LABELS[i + 4];
                return (
                  <tr key={leftKey} className="border-b border-theme-border/50">
                    <td className="py-1.5 pr-8 text-gray-500">{leftLabel}</td>
                    <td className="py-1.5 pr-12 text-right tabular-nums">
                      {fmt(data?.swap_rates?.[leftKey], 0)}
                    </td>
                    {right ? (
                      <>
                        <td className="py-1.5 pr-8 text-gray-500">{right[0]}</td>
                        <td className="py-1.5 text-right tabular-nums">
                          {fmt(data?.swap_rates?.[right[1]], 0)}
                        </td>
                      </>
                    ) : (
                      <td colSpan={2} />
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        <Divider className="!border-theme-border !my-4" />

        <div className="space-y-3">
          <p className="font-semibold text-sm">Margin rates</p>

          {marginSpec?.floating ? (
            <>
              <p className="text-sm text-gray-600">
                floating, {RANGE_MODE_LABEL[marginSpec.range_mode] ?? "volume"}
                {marginSpec.rule_path ? `, ${marginSpec.rule_path}` : ""}
              </p>
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr className="border-b border-theme-border">
                    <th className="text-left py-1 font-semibold">Range (lots)</th>
                    <th className="text-right py-1 font-semibold">Initial</th>
                    <th className="text-right py-1 font-semibold">Maintenance</th>
                  </tr>
                </thead>
                <tbody>
                  {(marginSpec.tiers ?? []).map((tier, i) => (
                    <tr key={i} className="border-b border-theme-border/50">
                      <td className="py-1">{fmtTierRange(tier.range_from, tier.range_to)}</td>
                      <td className="text-right py-1">{fmt(tier.margin_rate_initial, 2)}</td>
                      <td className="text-right py-1">{fmt(tier.margin_rate_maintenance, 2)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          ) : (
            <div className="grid grid-cols-2 gap-x-16 gap-y-0.5 text-sm">
              {[
                { label: "Initial Buy Margin", value: fmt(data?.margin_buy, 5) },
                { label: "Initial Sell Margin", value: fmt(data?.margin_sell, 5) },
                { label: "Maintenance Buy Margin", value: fmt(data?.maintenance_margin_buy, 5) },
                { label: "Maintenance Sell Margin", value: fmt(data?.maintenance_margin_sell, 5) },
              ].map((row) => (
                <SpecRow key={row.label} {...row} />
              ))}
            </div>
          )}
        </div>

        <Divider className="!border-theme-border !my-4" />

        <div className="space-y-3">
          <p className="font-semibold text-sm">Sessions</p>
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="border-b border-theme-border">
                <th className="text-left py-1.5 pr-3 font-semibold w-[100px]">Day</th>
                <th className="text-left py-1.5 px-3 font-semibold">
                  <span className="inline-flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-red shrink-0" />
                    Quotes
                  </span>
                </th>
                <th className="text-left py-1.5 pl-3 font-semibold">
                  <span className="inline-flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-green shrink-0" />
                    Trade
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((row) => (
                <tr key={row.key} className="border-b border-theme-border/50 align-top">
                  <td className="py-2 pr-3 font-medium">{row.day}</td>
                  <td className="py-2 px-3">
                    <div className="space-y-1">
                      <span className="tabular-nums">{row.quote}</span>
                      <div className="h-1.5 w-full max-w-[220px] rounded-full bg-red/80" />
                    </div>
                  </td>
                  <td className="py-2 pl-3">
                    <div className="space-y-1">
                      <span className="tabular-nums">{row.trade}</span>
                      <div className="h-1.5 w-full max-w-[220px] rounded-full bg-green/80" />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <Divider className="!border-theme-border !my-4" />

        <div className="grid grid-cols-2 gap-x-16 gap-y-0.5 text-sm">
          {symbolInfo.map((item) => (
            <SpecRow key={item.label} label={item.label} value={item.value} />
          ))}
        </div>
      </div>
    </ModalComponent>
  );
};

export default SymbolInfoModal;

import api from "../api";
import { qs } from "../../utils/utils";
import { adaptSymbol, mapList } from "../adapt";

export const getAllSymbols = () =>
  api.get("/symbols").then((r) => mapList(r, adaptSymbol));

// The chart suffixes its resolutions before sending; ours keys on the plain TradingView value,
// so they are put back. Minutes carry no suffix, days and above keep their letter.
const RESOLUTION = {
  "1m": "1",
  "5m": "5",
  "15m": "15",
  "30m": "30",
  "60m": "60",
  "240m": "240",
  "1d": "1D",
  "1w": "1W",
  "1mo": "1M",
};

// chart bars: ours answers in columns, the chart wants rows, and its times are milliseconds
export const marketHistory = (params) =>
  api
    .get(
      `/history?${qs({
        symbol: params.symbol_id,
        resolution: RESOLUTION[params.resolution] ?? params.resolution,
        from: params.from,
        to: params.to,
        countback: params.count_back,
      })}`
    )
    .then((r) => {
      const b = r.data ?? {};
      const rows =
        b.s === "ok"
          ? (b.t ?? []).map((t, i) => ({
              time: t * 1000,
              open: b.o?.[i],
              high: b.h?.[i],
              low: b.l?.[i],
              close: b.c?.[i],
              volume: b.v?.[i],
            }))
          : [];

      return { ...r, data: { data: rows, next_time: (b.nextTime ?? 0) * 1000 } };
    });

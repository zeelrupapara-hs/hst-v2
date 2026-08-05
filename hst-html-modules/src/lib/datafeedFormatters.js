import { fmtTs } from "./formatters.js";

/** mode: 1=quotes, 2=news, etc. — show N or Q per MT5 list. */
export function fmtDatafeedSource(row) {
  const mode = row.mode ?? 1;
  if (mode & 2 || row.news_count > 0 && !row.ticks_count) return "N";
  return "Q";
}

export function fmtDatafeedState(row) {
  const ticks = row.ticks_count ?? 0;
  const books = row.books_count ?? 0;
  const news = row.news_count ?? 0;
  if (news && !ticks) return String(news);
  if (books) return `${ticks} / ${books}`;
  return String(ticks);
}

export function fmtDatafeedLastActive(row) {
  const t = row.sys_last_time ?? row.updated_at;
  if (!t) return "—";
  return fmtTs(t);
}

export function fmtDatafeedSymbols(row) {
  const n = row.symbol_count;
  if (n != null) return String(n);
  return row.feed_symbols?.length != null ? String(row.feed_symbols.length) : "—";
}

export function fmtFeederMode(mode) {
  const m = Number(mode) || 0;
  const parts = [];
  if (m & 1) parts.push("Quotes");
  if (m & 2) parts.push("News");
  if (m & 4) parts.push("Market books");
  return parts.join(", ") || "Quotes";
}

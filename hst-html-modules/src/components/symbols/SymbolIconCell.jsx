import { fmtTradeMode, tradeModeClass } from "../../lib/symbolFormatters.js";
import {
  sessionDotClass,
  sessionDotStatus,
} from "../../lib/symbolSessionStatus.js";

export function SymbolIconCell({ row, sessions }) {
  const trade = tradeModeClass(row.trade_mode ?? 4);
  const status = sessionDotStatus(sessions);
  const sess = sessionDotClass(status);
  let title = fmtTradeMode(row.trade_mode ?? 4);
  if (status === "closed") title += " — market closed";
  else if (status === "quote") title += " — quote only";

  return (
    <span className="sym-icon-cell" title={title}>
      <span className={`sym-trade-icon ${trade}`} />
      <span className={`sym-session-dot ${sess}`} />
    </span>
  );
}

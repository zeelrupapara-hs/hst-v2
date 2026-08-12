import { request } from "@/api/client.js";

export const throwQuote = (symbol, bid, ask) =>
  request("/api/v1/quotes", { method: "POST", body: { symbol, bid, ask } });

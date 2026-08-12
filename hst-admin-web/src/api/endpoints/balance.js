import { request } from "@/api/client.js";

const op = (kind) => (login, amount, comment = "") =>
  request(`/api/v1/balance/${kind}`, { method: "POST", body: { login, amount, comment } });

export const createDeposit = op("deposit");
export const createWithdrawal = op("withdrawal");
export const createCredit = op("credit");
export const createCorrection = op("correction");

// Any balance-type deal the engine accepts: balance 2, credit 3, charge 4, correction 5,
// bonus 6, commission 7. The sign of the amount decides deposit or withdrawal.
export const createBalance = (login, action, amount, comment = "") =>
  request("/api/v1/balance", { method: "POST", body: { login, action, amount, comment } });

// The integrity audit: stored money beside what the account's deals add up to.
export const checkBalances = () => request("/api/v1/balance/check");
export const fixBalance = (login) =>
  request("/api/v1/balance/fix", { method: "POST", body: { login } });

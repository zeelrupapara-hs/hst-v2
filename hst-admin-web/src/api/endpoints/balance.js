import { request } from "@/api/client.js";

const op = (kind) => (login, amount, comment = "") =>
  request(`/api/v1/balance/${kind}`, { method: "POST", body: { login, amount, comment } });

export const createDeposit = op("deposit");
export const createWithdrawal = op("withdrawal");
export const createCredit = op("credit");
export const createCorrection = op("correction");

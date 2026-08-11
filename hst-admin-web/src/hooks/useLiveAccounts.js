import { useEffect, useState } from "react";
import { onEvent } from "@/api/socket.js";

// Live account money, parsed from the engine's summary lines:
// summary,login,balance,credit,equity,margin,free,level%,profit[,positionId,profit]...
// One shared cache; every consumer re-renders only when a line lands.
const accounts = new Map();
const listeners = new Set();
let watching = false;
let notifyTimer = 0;
const changedLogins = new Set();

// a burst of summaries paints once, not once per line
function notifySoon(login) {
  changedLogins.add(login);
  if (notifyTimer) return;
  notifyTimer = setTimeout(() => {
    notifyTimer = 0;
    const changed = [...changedLogins];
    changedLogins.clear();
    listeners.forEach((fn) => changed.forEach((l) => fn(l)));
  }, 150);
}

function readLine(line) {
  const parts = line.split(",");
  if (parts.length < 9) return;

  const login = Number(parts[1]);
  const positions = {};
  for (let i = 9; i + 1 < parts.length; i += 2) {
    positions[Number(parts[i])] = Number(parts[i + 1]);
  }

  accounts.set(login, {
    login,
    balance: Number(parts[2]),
    credit: Number(parts[3]),
    equity: Number(parts[4]),
    margin: Number(parts[5]),
    free: Number(parts[6]),
    level: Number(parts[7].replace("%", "")),
    profit: Number(parts[8]),
    positions,
    at: Date.now(),
  });

  notifySoon(login);
}

function watch() {
  if (watching) return;
  watching = true;
  onEvent("account_summary", (event) => readLine(event.payload));
}

/** @returns {Map<number, object>} live money per login, updated as summary lines arrive */
export function useLiveAccounts() {
  const [, bump] = useState(0);

  useEffect(() => {
    watch();
    const fn = () => bump((n) => n + 1);
    listeners.add(fn);
    return () => listeners.delete(fn);
  }, []);

  return accounts;
}

/** @returns {object|null} the live money of one login */
export function useLiveAccount(login) {
  const [account, setAccount] = useState(accounts.get(Number(login)) ?? null);

  useEffect(() => {
    watch();
    const wanted = Number(login);
    setAccount(accounts.get(wanted) ?? null);
    const fn = (changed) => changed === wanted && setAccount({ ...accounts.get(wanted) });
    listeners.add(fn);
    return () => listeners.delete(fn);
  }, [login]);

  return account;
}

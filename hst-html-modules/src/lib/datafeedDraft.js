import { FEEDER_MODE_QUOTES } from "./datafeedModules.js";

/** Default draft for Add → Data Feed settings. */
export function newDatafeedDraft() {
  return {
    name: "",
    module: "fix44",
    enable: 1,
    mode: FEEDER_MODE_QUOTES,
    feed_server: "",
    gateway_server: "",
    feed_login: 0,
    feed_password: "",
    gateway_login: 0,
    gateway_password: "",
    timeout: 60,
    timeout_reconnect: 5,
    timeout_sleep: 600,
    attempts_sleep: 3,
    allow_import_symbols: 0,
    params: [],
    feed_symbols: [],
    translates: [],
  };
}

export function isNewDatafeedId(id) {
  return id === "new" || id === "0";
}

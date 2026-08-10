import { useEffect, useRef, useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import { formatNs } from "@/lib/time.js";
import {
  fetchEndOfDay,
  fetchTimeSettings,
  runEndOfDay,
  updateEndOfDay,
  updateTimeSettings,
} from "@/api/endpoints/endOfDay.js";

// the browser already ships the IANA list; older engines get a single safe zone
const ZONES = typeof Intl.supportedValuesOf === "function" ? Intl.supportedValuesOf("timeZone") : ["UTC"];

const DST_RULES = [
  ["none", "None"],
  ["europe", "Europe"],
  ["usa", "USA"],
  ["australia", "Australia"],
];

/** The server clock: time zone, daylight saving rule, time source, and the daily rollover. */
export function TimeModule() {
  const [zone, setZone] = useState("UTC");
  const [dst, setDst] = useState("none");
  const [ntp, setNtp] = useState("");
  const [at, setAt] = useState("");
  const [updatedAt, setUpdatedAt] = useState(0);
  const [status, setStatus] = useState("");
  const [now, setNow] = useState(Date.now());
  const skewRef = useRef(0);

  const load = () => {
    fetchTimeSettings().then((res) => {
      if (!res.ok) {
        setStatus(res.message || "unavailable");
        return;
      }
      setZone(res.data.time_zone || "UTC");
      setDst(res.data.time_dst || "none");
      setNtp(res.data.time_ntp_server || "");
      skewRef.current = res.data.server_now / 1e6 - Date.now();
      setNow(Date.now() + skewRef.current);
    });
    fetchEndOfDay().then((res) => {
      if (res.ok) {
        setAt(res.data.at);
        setUpdatedAt(res.data.updated_at);
      }
    });
  };

  useEffect(() => {
    load();
  }, []);

  useEffect(() => {
    const t = setInterval(() => setNow(Date.now() + skewRef.current), 1000);
    return () => clearInterval(t);
  }, []);

  async function save() {
    const time = await updateTimeSettings({ time_zone: zone, time_dst: dst, time_ntp_server: ntp });
    if (!time.ok) {
      setStatus(time.message || "save failed");
      return;
    }
    const eod = await updateEndOfDay(at);
    setStatus(eod.ok ? "Time settings saved." : eod.message || "save failed");
    load();
  }

  async function run() {
    if (!window.confirm("Run the end-of-day rollover now? Swaps accrue and daily reports generate.")) return;
    const res = await runEndOfDay();
    setStatus(res.ok ? "End of day started." : res.message || "run failed");
  }

  let clock;
  try {
    clock = new Intl.DateTimeFormat("en-GB", {
      timeZone: zone,
      dateStyle: "medium",
      timeStyle: "medium",
    }).format(now);
  } catch {
    clock = new Date(now).toISOString().slice(0, 19).replace("T", " ");
  }

  return (
    <div className="module-root time-module">
      <div className="sym-sessions-intro">
        <span className="sym-tab-intro-icon" aria-hidden="true">
          <Icon id="time" size={48} />
        </span>
        <p>
          The server clock: the time zone every quote, order and report is stamped in, the daylight
          saving rule that shifts it, and the source it is synchronised from.
        </p>
      </div>
      <div className="form-grid time-grid">
        <label>Server time</label>
        <span className="time-clock">{clock}</span>
        <label>Time zone</label>
        <select value={zone} onChange={(e) => setZone(e.target.value)}>
          {!ZONES.includes(zone) && <option value={zone}>{zone}</option>}
          {ZONES.map((z) => (
            <option key={z} value={z}>{z}</option>
          ))}
        </select>
        <label>Daylight saving</label>
        <select value={dst} onChange={(e) => setDst(e.target.value)}>
          {DST_RULES.map(([value, text]) => (
            <option key={value} value={value}>{text}</option>
          ))}
        </select>
        <label>Time synchronization with</label>
        <input
          type="text"
          value={ntp}
          placeholder="pool.ntp.org"
          onChange={(e) => setNtp(e.target.value)}
        />
        <label>End of day at</label>
        <span className="grp-suffixed">
          <input
            type="text"
            value={at}
            placeholder="00:00"
            onChange={(e) => setAt(e.target.value)}
          />
          <span className="grp-suffix">server time, HH:MM</span>
        </span>
      </div>
      <div className="config-actions time-actions">
        <button type="button" className="config-ok" onClick={save}>Apply</button>
        <button type="button" onClick={run}>Run now</button>
      </div>
      {status && <p className="module-note">{status}</p>}
      {updatedAt > 0 && <p className="module-note">Last changed {formatNs(updatedAt)}</p>}
    </div>
  );
}

// the nav route still points at the end-of-day path this screen grew out of
export const EndOfDayModule = TimeModule;

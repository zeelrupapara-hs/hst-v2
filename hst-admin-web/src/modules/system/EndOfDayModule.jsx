import { useEffect, useState } from "react";
import { Icon } from "@/components/ui/Icon.jsx";
import { formatNs } from "@/lib/time.js";
import { fetchEndOfDay, runEndOfDay, updateEndOfDay } from "@/api/endpoints/endOfDay.js";

/** The daily rollover: when it runs, and running it by hand. */
export function EndOfDayModule() {
  const [at, setAt] = useState("");
  const [updatedAt, setUpdatedAt] = useState(0);
  const [status, setStatus] = useState("");

  const load = () =>
    fetchEndOfDay().then((res) => {
      if (res.ok) {
        setAt(res.data.at);
        setUpdatedAt(res.data.updated_at);
      } else {
        setStatus(res.message || "unavailable");
      }
    });

  useEffect(() => {
    load();
  }, []);

  async function save() {
    const res = await updateEndOfDay(at);
    setStatus(res.ok ? "Rollover hour saved." : res.message || "save failed");
    if (res.ok) load();
  }

  async function run() {
    if (!window.confirm("Run the end-of-day rollover now? Swaps accrue and daily reports generate.")) return;
    const res = await runEndOfDay();
    setStatus(res.ok ? "End of day started." : res.message || "run failed");
  }

  return (
    <div className="module-root eod-module">
      <div className="sym-sessions-intro">
        <span className="sym-tab-intro-icon" aria-hidden="true">
          <Icon id="time" size={48} />
        </span>
        <p>
          The end-of-day rollover accrues swaps, settles daily profit and generates reports. It
          runs every day at the server hour set here.
        </p>
      </div>
      <div className="form-grid eod-grid">
        <label>Run at</label>
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
      <div className="config-actions eod-actions">
        <button type="button" className="config-ok" onClick={save}>Apply</button>
        <button type="button" onClick={run}>Run now</button>
      </div>
      {status && <p className="module-note">{status}</p>}
      {updatedAt > 0 && <p className="module-note">Last changed {formatNs(updatedAt)}</p>}
    </div>
  );
}

import { useState } from "react";

export function Toolbox({ panel }) {
  const [tab, setTab] = useState("journal");
  const tabs =
    panel === "admin"
      ? [
          { id: "journal", label: "Journal" },
          { id: "news", label: "News" },
          { id: "mail", label: "Mail" },
          { id: "search", label: "Search" },
        ]
      : [
          { id: "journal", label: "Journal" },
          { id: "news", label: "News" },
        ];

  return (
    <div className="toolbox" id="toolbox">
      <div className="toolbox-tabs">
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            className={tab === t.id ? "active" : undefined}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>
      {tab === "journal" && (
        <div className="toolbox-panel active">
          <table className="journal-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Server</th>
                {panel === "admin" && <th>IP</th>}
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              {panel === "admin" ? (
                <>
                  <tr>
                    <td>2026.08.03 10:00:01</td>
                    <td>Trade</td>
                    <td>127.0.0.1</td>
                    <td>Configuration synchronized</td>
                  </tr>
                  <tr>
                    <td>2026.08.03 09:58:12</td>
                    <td>Trade</td>
                    <td>127.0.0.1</td>
                    <td>Manager login 1000 connected</td>
                  </tr>
                </>
              ) : (
                <tr>
                  <td>2026.08.03 10:01:05</td>
                  <td>Trade</td>
                  <td>Dealer 1001 connected</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
      {tab !== "journal" && (
        <div className="toolbox-panel active">
          <p style={{ padding: 8, color: "#666" }}>
            {tabs.find((t) => t.id === tab)?.label} — phase 2
          </p>
        </div>
      )}
    </div>
  );
}

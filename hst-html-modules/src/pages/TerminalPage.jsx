import { Navigate, useParams } from "react-router-dom";
import { TerminalShell } from "../components/layout/TerminalShell.jsx";
import { adminNav } from "../features/navigation/adminNav.js";
import { managerNav } from "../features/navigation/managerNav.js";
import {
  DEFAULT_MODULE,
  getModuleLabel,
} from "../features/modules/registry.js";
import { useDemo } from "../hooks/useDemo.js";

const TITLES = {
  admin: "HST Administrator — MetaTrader Server (Live)",
  manager: "HST Manager — MetaTrader Server (Live)",
};

export function TerminalPage({ panel }) {
  const { moduleId, recordId } = useParams();
  const demo = useDemo();
  const nav = panel === "admin" ? adminNav : managerNav;

  if (!moduleId) {
    return <Navigate to={`/${panel}/${DEFAULT_MODULE[panel]}`} replace />;
  }

  const moduleLabel = recordId
    ? `${getModuleLabel(panel, moduleId)} #${recordId}`
    : getModuleLabel(panel, moduleId);

  return (
    <TerminalShell
      panel={panel}
      title={TITLES[panel]}
      nav={nav}
      moduleLabel={moduleLabel}
      demo={demo}
    />
  );
}

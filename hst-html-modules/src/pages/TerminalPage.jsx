import { Navigate, useParams, useSearchParams } from "react-router-dom";
import { TerminalShell } from "../components/layout/TerminalShell.jsx";
import {
  DEFAULT_MODULE,
  getModuleLabel,
} from "../features/modules/registry.js";
import { useDemo } from "../hooks/useDemo.js";
import { useAdminNavTree } from "../hooks/useAdminNavTree.js";

const TITLES = {
  admin: "HST Administrator — Trade Server (Live)",
  manager: "HST Manager — Trade Server (Live)",
};

export function TerminalPage({ panel }) {
  const { moduleId, recordId } = useParams();
  const [searchParams] = useSearchParams();
  const demo = useDemo();
  const activeFolder =
    panel === "admin" && moduleId === "symbols"
      ? searchParams.get("folder") || ""
      : "";
  const activeDatafeedId =
    panel === "admin" &&
    moduleId === "datafeeds" &&
    recordId &&
    recordId !== "new"
      ? recordId
      : "";
  const { nav, refreshNav } = useAdminNavTree(
    panel,
    activeFolder,
    activeDatafeedId
  );

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
      refreshNav={refreshNav}
      onNavRefresh={refreshNav}
      moduleLabel={moduleLabel}
      demo={demo}
    />
  );
}

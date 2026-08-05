import { lazy, Suspense } from "react";
import { Navigate, Route, Routes, useParams } from "react-router-dom";
import { ModuleView } from "../components/module/ModuleView.jsx";

const HomePage = lazy(() =>
  import("../pages/HomePage.jsx").then((m) => ({ default: m.HomePage }))
);
const LoginPage = lazy(() =>
  import("../pages/LoginPage.jsx").then((m) => ({ default: m.LoginPage }))
);
const TerminalPage = lazy(() =>
  import("../pages/TerminalPage.jsx").then((m) => ({ default: m.TerminalPage }))
);

function PageFallback() {
  return (
    <div className="picker">
      <div className="picker-card">
        <p>Loading…</p>
      </div>
    </div>
  );
}

function PanelModuleRoute({ panel }) {
  const { moduleId, recordId } = useParams();
  return <ModuleView panel={panel} moduleId={moduleId} recordId={recordId} />;
}

export function AppRoutes() {
  return (
    <Suspense fallback={<PageFallback />}>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<TerminalPage panel="admin" />}>
          <Route index element={<Navigate to="clients" replace />} />
          <Route path=":moduleId" element={<PanelModuleRoute panel="admin" />} />
          <Route
            path=":moduleId/:recordId"
            element={<PanelModuleRoute panel="admin" />}
          />
        </Route>
        <Route path="/manager" element={<TerminalPage panel="manager" />}>
          <Route index element={<Navigate to="clients" replace />} />
          <Route path=":moduleId" element={<PanelModuleRoute panel="manager" />} />
          <Route
            path=":moduleId/:recordId"
            element={<PanelModuleRoute panel="manager" />}
          />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
}

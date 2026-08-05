import { Navigate, Route, Routes } from "react-router-dom";
import { LoginPage } from "@/app/LoginPage.jsx";
import { SessionProvider } from "@/app/session.jsx";
import { TerminalPage } from "@/app/TerminalPage.jsx";

export function AppRoutes() {
  return (
    <SessionProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin/*" element={<TerminalPage panel="admin" />} />
        <Route path="/manager/*" element={<TerminalPage panel="manager" />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </SessionProvider>
  );
}

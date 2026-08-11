import { Navigate, Route, Routes } from "react-router-dom";
import { TerminalShell } from "@/components/layout/TerminalShell.jsx";
import { SymbolsModule } from "@/modules/symbols/SymbolsModule.jsx";
import { DatafeedsModule } from "@/modules/datafeeds/DatafeedsModule.jsx";
import { GroupsModule } from "@/modules/groups/GroupsModule.jsx";
import { AccountsModule } from "@/modules/accounts/AccountsModule.jsx";
import { ClientsModule } from "@/modules/clients/ClientsModule.jsx";
import { LeveragesModule } from "@/modules/leverages/LeveragesModule.jsx";
import { RoutingModule } from "@/modules/routing/RoutingModule.jsx";
import { HolidaysModule } from "@/modules/holidays/HolidaysModule.jsx";
import { EndOfDayModule } from "@/modules/system/EndOfDayModule.jsx";
import { ManagersModule } from "@/modules/managers/ManagersModule.jsx";
import { DealsModule, OrdersModule, PositionsModule } from "@/modules/trades/TradeBlotters.jsx";
import { DealingModule } from "@/modules/dealing/DealingModule.jsx";
import { BalanceModule } from "@/modules/balance/BalanceModule.jsx";
import { MailServersModule } from "@/modules/mail/MailServersModule.jsx";
import { useSession } from "@/hooks/useSession.js";

function ModulePlaceholder() {
  return (
    <div style={{ padding: 16, color: "#666" }}>
      <p>Select a module in the Navigator. Modules land phase by phase.</p>
    </div>
  );
}

/** @param {{panel: "admin"|"manager"}} props */
export function TerminalPage({ panel }) {
  const session = useSession();

  if (session.status === "loading") return null;
  if (session.status !== "ready") return <Navigate to="/login" replace />;

  // the session's terminal decides the panel; the URL cannot claim the other one
  const owned = session.terminal === "administrator" ? "admin" : "manager";
  if (panel !== owned) return <Navigate to={`/${owned}`} replace />;

  return (
    <Routes>
      <Route element={<TerminalShell panel={panel} />}>
        <Route index element={<ModulePlaceholder />} />
        <Route path="symbols" element={<SymbolsModule />} />
        <Route path="datafeeds" element={<DatafeedsModule />} />
        <Route path="groups" element={<GroupsModule />} />
        <Route path="users" element={<AccountsModule />} />
        <Route path="clients" element={<ClientsModule />} />
        <Route path="leverage-profiles" element={<LeveragesModule />} />
        <Route path="routing" element={<RoutingModule />} />
        <Route path="holidays" element={<HolidaysModule />} />
        <Route path="system/end-of-day" element={<EndOfDayModule />} />
        <Route path="managers" element={<ManagersModule />} />
        <Route path="positions" element={<PositionsModule />} />
        <Route path="orders" element={<OrdersModule />} />
        <Route path="deals" element={<DealsModule />} />
        <Route path="dealing" element={<DealingModule />} />
        <Route path="balance" element={<BalanceModule />} />
        <Route path="mail-servers" element={<MailServersModule />} />
        <Route path="*" element={<ModulePlaceholder />} />
      </Route>
    </Routes>
  );
}

import { Link } from "react-router-dom";
import { CenteredCard } from "../components/ui/CenteredCard.jsx";

export function HomePage() {
  return (
    <CenteredCard
      title="HST Admin / Manager — React Prototype"
      description="MT5-style terminal with demo data pre-loaded. Edit React components under src/ and see changes instantly with hot reload."
    >
      <div className="picker-links">
        <Link to="/login">
          Connect API
          <small>Sign in to hst-server — required before modules load live data</small>
        </Link>
        <Link to="/admin/clients">
          Administrator Panel
          <small>Platform setup — Groups, Symbols, Routing, Data Feeds…</small>
        </Link>
        <Link to="/manager/clients">
          Manager Panel
          <small>Day-to-day — Clients, Orders, Positions, Dealing, Balance…</small>
        </Link>
      </div>
    </CenteredCard>
  );
}

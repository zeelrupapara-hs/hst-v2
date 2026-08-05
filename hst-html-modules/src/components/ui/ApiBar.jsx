import { Link } from "react-router-dom";

export function ApiBar({ state }) {
  let dotClass = "api-dot";
  let message = state.message || "…";

  if (state.loading) {
    dotClass += " load";
    message = "Loading…";
  } else if (state.error) {
    dotClass += " err";
    message = state.error;
  } else if (state.demo) {
    dotClass += " warn";
    if (!state.message) {
      message = "Demo data — sign in via Connect API for live server";
    }
  } else {
    dotClass += " ok";
  }

  return (
    <div className="api-bar">
      <span className={dotClass} />
      <span>
        {state.demo && state.message?.includes("Connect API") ? (
          <>
            Demo mode —{" "}
            <Link to="/login" target="_top">
              Connect API
            </Link>{" "}
            for live data
          </>
        ) : (
          message
        )}
      </span>
    </div>
  );
}

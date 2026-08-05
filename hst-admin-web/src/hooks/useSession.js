import { createContext, useContext } from "react";

// The provider lives in app/session.jsx; modules and layout reach the session through here.
export const SessionContext = createContext(null);

export const useSession = () => useContext(SessionContext);

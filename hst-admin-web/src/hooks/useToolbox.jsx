import { createContext, useCallback, useContext, useState } from "react";

export const ToolboxContext = createContext(null);

/** Provider lives in TerminalShell; modules request journal/search focus. */
export function ToolboxProvider({ children }) {
  const [journalQuery, setJournalQuery] = useState("");
  const [tab, setTab] = useState("journal");
  const [tick, setTick] = useState(0);

  const openJournal = useCallback((query = "") => {
    setJournalQuery(query);
    setTab("journal");
    setTick((n) => n + 1);
  }, []);

  return (
    <ToolboxContext.Provider value={{ tab, setTab, journalQuery, tick, openJournal }}>
      {children}
    </ToolboxContext.Provider>
  );
}

export const useToolbox = () => useContext(ToolboxContext);

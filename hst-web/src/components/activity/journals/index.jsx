import useJournalStore from "../../../store/useJournalStore";
import useJournals from "../../../columns/useJournals";
import SimpleTable from "../../table/SimpleTable";

const Journals = () => {
  const logs = useJournalStore((state) => state.logs);
  const columns = useJournals();

  return <SimpleTable columns={columns} data={logs} />;
};

export default Journals;

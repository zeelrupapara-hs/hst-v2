import useHistoryStore from "../../../store/useHistoryStore";
import useDeals from "../../../columns/useDeals";
import SimpleTable from "../../table/SimpleTable";

const Deals = () => {
  const deals = useHistoryStore((state) => state.deals);
  const columns = useDeals();

  return <SimpleTable columns={columns} data={deals} />;
};

export default Deals;

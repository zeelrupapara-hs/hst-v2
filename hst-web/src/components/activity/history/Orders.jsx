import useHistoryStore from "../../../store/useHistoryStore";
import useOrders from "../../../columns/useOrders";
import SimpleTable from "../../table/SimpleTable";

const Orders = () => {
  const orders = useHistoryStore((state) => state.orders);
  const columns = useOrders();

  return <SimpleTable columns={columns} data={orders} />;
}

export default Orders
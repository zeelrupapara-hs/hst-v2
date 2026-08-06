import useHistoryStore from "../../../store/useHistoryStore";
import useClosedPositions from "../../../columns/useClosedPositions";
import SimpleTable from "../../table/SimpleTable";

const ClosedPositions = () => {
  const positions = useHistoryStore((state) => state.positions);
  const columns = useClosedPositions();

  return <SimpleTable columns={columns} data={positions} />;
};

export default ClosedPositions;

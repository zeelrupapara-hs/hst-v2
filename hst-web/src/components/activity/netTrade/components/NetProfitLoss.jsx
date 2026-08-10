import usePositionStore from "../../../../store/usePositionStore";
import { formatMoney } from "../../../../utils/utils";

const NetProfitLoss = ({ positions }) => {
  const positionPL = usePositionStore((state) => state.positionPL);
  const digits = usePositionStore((state) => state.currencyDigits);

  const netProfit = positions.reduce((sum, pos) => {
    return sum + Number(positionPL[pos?.id] || 0);
  }, 0);

  return (
    <span className={`${netProfit >= 0 ? "text-green" : "text-red"}`}>
      {formatMoney(netProfit, digits)}
    </span>
  );
};

export default NetProfitLoss;

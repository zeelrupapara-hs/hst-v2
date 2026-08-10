import usePositionStore from "../../../../store/usePositionStore";
import { formatMoney } from "../../../../utils/utils";

const ProfitLoss = ({ positionId }) => {
  const profit = usePositionStore((state) => state.positionPL[positionId]);
  const digits = usePositionStore((state) => state.currencyDigits);

  return (
    <div className={`${profit >= 0 ? "text-green" : "text-red"}`}>
      {formatMoney(profit, digits)}
    </div>
  );
};

export default ProfitLoss;

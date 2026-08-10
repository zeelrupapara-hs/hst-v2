import usePositionStore from "../../../../store/usePositionStore";
import { formatMoney } from "../../../../utils/utils";

const TotalPNL = () => {
  const { profit } = usePositionStore((state) => state.summary);
  const digits = usePositionStore((state) => state.currencyDigits);

  return (
    <span className={`${profit >= 0 ? "text-green" : "text-red"}`}>
      {formatMoney(profit, digits)}
    </span>
  );
};

export default TotalPNL;

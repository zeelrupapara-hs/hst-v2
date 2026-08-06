import usePositionStore from "../../../../store/usePositionStore";

const ProfitLoss = ({ positionId }) => {
  const profit = usePositionStore((state) => state.positionPL[positionId]);

  return (
    <div className={`${profit >= 0 ? "text-green" : "text-red"}`}>{profit}</div>
  );
};

export default ProfitLoss;

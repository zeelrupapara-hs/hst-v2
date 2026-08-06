import usePositionStore from "../../../../store/usePositionStore";

const NetProfitLoss = ({ positions }) => {
  const positionPL = usePositionStore((state) => state.positionPL);

  const netProfit = positions.reduce((sum, pos) => {
    return sum + (positionPL[pos?.id] || 0);
  }, 0);

  return (
    <span className={`${netProfit >= 0 ? "text-green" : "text-red"}`}>
      {Number(netProfit.toFixed(2)) || 0}
    </span>
  );
};

export default NetProfitLoss;

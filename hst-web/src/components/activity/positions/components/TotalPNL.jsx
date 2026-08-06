import usePositionStore from "../../../../store/usePositionStore";

const TotalPNL = () => {
  const { profit } = usePositionStore((state) => state.summary);

  return (
    <span className={`${profit >= 0 ? "text-green" : "text-red"}`}>
      {profit}
    </span>
  );
};

export default TotalPNL;

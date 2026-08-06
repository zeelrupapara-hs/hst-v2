import useGlobalStore from "../../store/useGlobalStore";
import OrderModal from "./components/OrderModal";

const MultiOrderScreen = () => {
  const orderModals = useGlobalStore((state) => state.orderModals);

  return (
    <>
      {orderModals.map((symbolId) => (
        <OrderModal key={symbolId} symbolId={symbolId} />
      ))}
    </>
  );
};

export default MultiOrderScreen;

import useLiveSymbolStore from "../../../../store/useLiveSymbolStore";
import { calculateNetValue } from "../../../../utils/calculations";

const NetValue = ({ position }) => {
  const [symbolId, side] = position?.id?.split("_");

  const { last_ask, last_bid } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  // the position's close-out value: a buy closes at bid, a sell at ask
  const currentPrice = Number(side) === 0 ? last_bid : last_ask;

  const netValue = calculateNetValue({ ...position, price: currentPrice });

  return <span>{netValue || 0}</span>;
};

export default NetValue;

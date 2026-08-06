import useLiveSymbolStore from "../../../store/useLiveSymbolStore";
import Ask from "../../marketWatch/components/Ask";
import Bid from "../../marketWatch/components/Bid";

const BuySell = ({ symbolId, formValues, disabled, createOrder }) => {
  const { last_bid } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  const price = last_bid || 0;
  const { stop_loss, take_profit } = formValues;

  const disableBuy =
    (stop_loss && stop_loss >= price) ||
    (take_profit && take_profit <= price) ||
    disabled;

  const disableSell =
    (stop_loss && stop_loss <= price) ||
    (take_profit && take_profit >= price) ||
    disabled;

  return (
    <div className="grid grid-cols-2 gap-2">
      <div className="flex flex-col items-center gap-3">
        <Bid symbolId={symbolId} />

        <button
          className="w-full bg-red text-white p-2 rounded-sm"
          onClick={() => createOrder(1)}
          disabled={disableSell}
        >
          Sell
        </button>
      </div>

      <div className="flex flex-col items-center gap-3">
        <Ask symbolId={symbolId} />

        <button
          className="w-full bg-green text-white p-2 rounded-sm"
          onClick={() => createOrder(0)}
          disabled={disableBuy}
        >
          Buy
        </button>
      </div>
    </div>
  );
};

export default BuySell;

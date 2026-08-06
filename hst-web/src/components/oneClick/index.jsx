import { useEffect, useState } from "react";
import useSymbolStore from "../../store/useSymbolStore";
import useGlobalStore from "../../store/useGlobalStore";
import useSymbolLive from "../../hooks/useSymbolLive";
import Value from "../common/Value";
import { validateOrderField } from "../../utils/validation";
import Sell from "./components/Sell";
import Buy from "./components/Buy";
import { useSocket } from "../../socket";
import { SOCKET_EVENTS } from "../../socket/events";

const OneClick = ({ symbolId, variant }) => {
  const { sendEvent } = useSocket();
  const { min_value, max_value } = useSymbolStore(
    (state) => state.symbols?.[symbolId]
  );
  const storedVolume = useGlobalStore((state) => state.volume?.[symbolId]);
  const setSymbolVolume = useGlobalStore((state) => state.setSymbolVolume);
  const [volume, setVolume] = useState(storedVolume || min_value || 0.01);
  const [error, setError] = useState(false);
  const isLive = useSymbolLive(symbolId);

  const min = min_value || 0.01;
  const max = max_value || 100;
  const hasError = error || !volume;
  const isDisabled = hasError || !isLive;

  useEffect(() => {
    setVolume(storedVolume || min_value || 0.01);
  }, [symbolId, storedVolume]);

  const getNewValue = (newValue) => {
    return volume === null ? min || max : newValue;
  };

  const updateValue = (key, value, range) => {
    const error = validateOrderField(key, value, range);
    setError(error);
    setVolume(value);
  };

  const createOrder = (side = 0) => {
    if (isDisabled) return;

    const payload = {
      symbol_id: symbolId,
      type: 0,
      side,
      volume,
      order_price: 1,
      fill_policy: 0,
    };

    sendEvent(SOCKET_EVENTS.ORDER_CREATE, payload);
  };

  return (
    <div className="flex">
      <Sell
        symbolId={symbolId}
        disabled={isDisabled}
        createOrder={createOrder}
        variant={variant}
      />

      <div className="w-[120px]">
        <Value
          size="small"
          min={min}
          max={max}
          value={volume}
          disabled={!isLive}
          setValue={(val, inputType) => {
            const newValue = inputType === "manual" ? val : getNewValue(val);
            updateValue("volume", newValue, { min, max });
            setSymbolVolume(symbolId, newValue);
          }}
          className="!rounded-none !border-x-0"
        />
      </div>

      <Buy
        symbolId={symbolId}
        disabled={isDisabled}
        createOrder={createOrder}
        variant={variant}
      />
    </div>
  );
};

export default OneClick;

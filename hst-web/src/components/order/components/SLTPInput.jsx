import { useEffect } from "react";
import { Form } from "antd";
import useLiveSymbolStore from "../../../store/useLiveSymbolStore";
import Value from "../../common/Value";
import { getPriceRange, validateOrderField } from "../../../utils/validation";

const SLTPInput = ({
  mode = "add",
  name,
  label,
  symbolId,
  data,
  value,
  setValue,
  setError,
}) => {
  const { last_ask, last_bid, digits } = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]
  );

  const { type, side, order_price } = data;

  const { min, max, range } = getPriceRange({
    mode,
    name,
    type,
    side,
    last_ask,
    last_bid,
    order_price,
  });

  useEffect(() => {
    if (setError) {
      const error = validateOrderField(name, value, { min, max }, data);
      setError(error);
    }
  }, [min, max]);

  const getNewValue = (newValue) => {
    return value === null ? min || last_bid || max : newValue;
  };

  return (
    <Form.Item name={name}>
      <div>
        <div className="flex items-center justify-between gap-2 pb-1 text-xs">
          <span>{label}</span>
          <span>{range}</span>
        </div>

        <Value
          min={min}
          max={max}
          digits={digits}
          value={value}
          setValue={(val, inputType) => {
            const newValue = inputType === "manual" ? val : getNewValue(val);
            setValue(newValue, { min, max });
          }}
        />
      </div>
    </Form.Item>
  );
};

export default SLTPInput;

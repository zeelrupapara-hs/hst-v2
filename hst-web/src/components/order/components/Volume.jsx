import { Form } from "antd";
import useSymbolStore from "../../../store/useSymbolStore";
import Value from "../../common/Value";

const Volume = ({
  mode = "add",
  name,
  label,
  symbolId,
  data,
  value,
  setValue,
}) => {
  const { min_value, max_value } = useSymbolStore(
    (state) => state.symbols?.[symbolId]
  );

  const min = mode === "edit" ? 0.01 : min_value || 0.01;
  const max = mode === "edit" ? data?.volume : max_value || 100;
  const range = `Min: ${min} | Max: ${max}`;

  const getNewValue = (newValue) => {
    return value === null ? min || max : newValue;
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

export default Volume;

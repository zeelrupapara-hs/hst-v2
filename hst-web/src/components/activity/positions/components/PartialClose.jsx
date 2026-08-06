import { useEffect, useMemo, useState } from "react";
import { Form, Popover } from "antd";
import { LuCircleMinus } from "react-icons/lu";
import { useSocket } from "../../../../socket";
import usePositionStore from "../../../../store/usePositionStore";
import Volume from "../../../order/components/Volume";
import { validateOrderField } from "../../../../utils/validation";
import { SOCKET_EVENTS } from "../../../../socket/events";
import Icon from "../../../common/Icon";

const PartialClose = ({ record }) => {
  const { sendEvent } = useSocket();
  const profit = usePositionStore((state) => state.positionPL[record?.id]);
  const [form] = Form.useForm();
  const [newValue, setNewValue] = useState(null);
  const [error, setError] = useState(false);
  const [open, setOpen] = useState(false);

  const symbolId = record?.symbol_id;
  const hasError = error || !newValue;

  const partialProfit = useMemo(() => {
    if (!profit || !newValue || !record?.volume) return 0;

    const partialProfit = profit * (newValue / record?.volume);

    if (profit > 0) {
      return Math.min(profit, partialProfit);
    } else {
      return Math.max(profit, partialProfit);
    }
  }, [profit, newValue, record?.volume]);

  useEffect(() => {
    setNewValue(record?.volume || null);
  }, [record]);

  const updateValue = (key, value, range) => {
    const error = validateOrderField(key, value, range);
    setError(error);
    setNewValue(value);
  };

  const handleCancel = () => {
    form.resetFields();
    setNewValue(null);
    setOpen(false);
  };

  const handleClose = () => {
    sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
      position_id: record?.id,
      volume: newValue,
    });
  };

  const content = (
    <Form form={form} layout="vertical" size="medium" className="custom-form">
      <Volume
        mode={"edit"}
        name={"volume"}
        label={"Lot"}
        symbolId={symbolId}
        data={record}
        value={newValue}
        setValue={(val, range) => updateValue("volume", val, range)}
      />

      <div className="flex items-center justify-between gap-2 pb-2 text-xs">
        <span>Partial P/L</span>
        <span className={`${partialProfit >= 0 ? "text-green" : "text-red"}`}>
          {partialProfit?.toFixed(2)}
        </span>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <button
          className="w-full bg-red text-white p-1 rounded-sm"
          onClick={handleCancel}
        >
          Cancel
        </button>

        <button
          className="w-full bg-green text-white p-1 rounded-sm"
          onClick={handleClose}
          disabled={hasError}
        >
          Close
        </button>
      </div>
    </Form>
  );

  return (
    <Popover
      content={content}
      trigger="click"
      arrow={false}
      open={open}
      onOpenChange={() => setOpen(!open)}
    >
      <button>
        <Icon Icon={LuCircleMinus} size={18} />
      </button>
    </Popover>
  );
};

export default PartialClose;

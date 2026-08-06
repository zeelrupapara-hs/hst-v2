import { useEffect, useState } from "react";
import { Form, Popover } from "antd";
import { LuX } from "react-icons/lu";
import { useSocket } from "../../../../socket";
import SLTPInput from "../../../order/components/SLTPInput";
import { validateOrderField } from "../../../../utils/validation";
import { SOCKET_EVENTS } from "../../../../socket/events";
import Icon from "../../../common/Icon";

const EditSLTP = ({ name, label, value, record }) => {
  const { sendEvent } = useSocket();
  const [form] = Form.useForm();
  const [newValue, setNewValue] = useState(null);
  const [error, setError] = useState(false);
  const [open, setOpen] = useState(false);

  const symbolId = record?.symbol_id;
  const isPendingOrder = record?.type && record?.type !== 0;
  const hasError = error || !newValue;

  useEffect(() => {
    setNewValue(value || null);
  }, [value]);

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

  const handleSave = (newValue) => {
    if (isPendingOrder) updateOrder(newValue);
    else updatePosition(newValue);
  };

  const updateOrder = (newValue) => {
    if (hasError) return;

    sendEvent(SOCKET_EVENTS.ORDER_UPDATE, {
      order_id: record?.id,
      ...record,
      [name]: newValue,
    });

    handleCancel();
  };

  const updatePosition = (newValue) => {
    if (hasError) return;

    sendEvent(SOCKET_EVENTS.POSITION_UPDATE, {
      position_id: record?.id,
      stop_loss: record?.stop_loss,
      take_profit: record?.take_profit,
      [name]: newValue,
    });

    handleCancel();
  };

  const content = (
    <Form form={form} layout="vertical" size="medium" className="custom-form">
      <SLTPInput
        mode={"edit"}
        name={name}
        label={label}
        symbolId={symbolId}
        data={{ ...record, order_price: record?.order_limit_price }}
        value={newValue}
        setValue={(val, range) => updateValue(name, val, range)}
      />

      <div className="grid grid-cols-2 gap-2">
        <button
          className="w-full bg-red text-white p-1 rounded-sm"
          onClick={handleCancel}
        >
          Cancel
        </button>

        <button
          className="w-full bg-green text-white p-1 rounded-sm"
          onClick={() => handleSave(newValue)}
          disabled={hasError}
        >
          Save
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
      <div className="cursor-pointer">
        {value ? (
          <div className="flex items-center gap-2">
            <span className="leading-none">{value}</span>

            <button
              onClick={(e) => {
                e.stopPropagation();
                handleSave(null);
              }}
            >
              <Icon Icon={LuX} size={14} className="stroke-red" />
            </button>
          </div>
        ) : (
          <button>Add</button>
        )}
      </div>
    </Popover>
  );
};

export default EditSLTP;

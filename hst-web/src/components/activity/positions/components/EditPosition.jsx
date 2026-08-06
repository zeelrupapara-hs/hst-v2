import { useEffect, useState } from "react";
import { Form, Popover } from "antd";
import { LuPencil } from "react-icons/lu";
import { useSocket } from "../../../../socket";
import Volume from "../../../order/components/Volume";
import PriceInput from "../../../order/components/PriceInput";
import SLTPInput from "../../../order/components/SLTPInput";
import { validateOrderField } from "../../../../utils/validation";
import { SOCKET_EVENTS } from "../../../../socket/events";
import Icon from "../../../common/Icon";

const EditPosition = ({ record }) => {
  const { sendEvent } = useSocket();
  const [form] = Form.useForm();
  const [formValues, setFormValues] = useState({});
  const [errors, setErrors] = useState({});
  const [open, setOpen] = useState(false);

  const symbolId = record?.symbol_id;
  const isPendingOrder = record?.type && record?.type !== 0;
  const hasErrors = Object.values(errors).some((val) => val === true);

  useEffect(() => {
    setFormValues({
      type: record?.type,
      side: record?.side,
      volume: record?.volume,
      order_price: record?.open_price || record?.order_limit_price,
      stop_loss: record?.stop_loss,
      take_profit: record?.take_profit,
    });
  }, [record]);

  const updateFormValue = (key, value, range) => {
    const error = validateOrderField(key, value, range);
    setErrors({ ...error, [key]: error });
    setFormValues({ ...formValues, [key]: value });
  };

  const updateError = (key, value) => {
    if (errors?.[key] === value) return;
    setErrors((prev) => ({ ...prev, [key]: value }));
  };

  const handleCancel = () => {
    form.resetFields();
    setOpen(false);
  };

  const handleSave = () => {
    if (isPendingOrder) updateOrder();
    else updatePosition();
  };

  const updateOrder = () => {
    if (hasErrors) return;

    sendEvent(SOCKET_EVENTS.ORDER_UPDATE, {
      order_id: record?.id,
      ...record,
      volume: formValues?.volume,
      order_limit_price: formValues?.order_price,
      stop_loss: formValues?.stop_loss,
      take_profit: formValues?.take_profit,
    });

    handleCancel();
  };

  const updatePosition = () => {
    if (hasErrors) return;

    sendEvent(SOCKET_EVENTS.POSITION_UPDATE, {
      position_id: record?.id,
      stop_loss: formValues?.stop_loss,
      take_profit: formValues?.take_profit,
    });

    handleCancel();
  };

  const content = (
    <Form
      form={form}
      layout="vertical"
      size="medium"
      className="custom-form"
      initialValues={formValues}
    >
      {isPendingOrder ? (
        <div className="grid grid-cols-2 gap-2">
          <Volume
            name={"volume"}
            label={"Lot"}
            symbolId={symbolId}
            value={formValues?.volume}
            setValue={(val, range) => updateFormValue("volume", val, range)}
          />

          <PriceInput
            mode={"edit"}
            name={"order_price"}
            label={"Price"}
            symbolId={symbolId}
            data={formValues}
            value={formValues?.order_price}
            setValue={(val, range) =>
              updateFormValue("order_price", val, range)
            }
            setError={(error) => updateError("order_price", error)}
          />
        </div>
      ) : null}

      <div className="grid grid-cols-2 gap-2">
        <SLTPInput
          mode={"edit"}
          name={"stop_loss"}
          label={"SL"}
          symbolId={symbolId}
          data={formValues}
          value={formValues?.stop_loss}
          setValue={(val, range) => updateFormValue("stop_loss", val, range)}
          setError={(error) => updateError("stop_loss", error)}
        />

        <SLTPInput
          mode={"edit"}
          name={"take_profit"}
          label={"TP"}
          symbolId={symbolId}
          data={formValues}
          value={formValues?.take_profit}
          setValue={(val, range) => updateFormValue("take_profit", val, range)}
          setError={(error) => updateError("take_profit", error)}
        />
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
          onClick={handleSave}
          disabled={hasErrors}
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
      <button>
        <Icon Icon={LuPencil} size={16} />
      </button>
    </Popover>
  );
};

export default EditPosition;

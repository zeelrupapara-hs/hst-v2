import { useEffect, useState } from "react";
import dayjs from "dayjs";
import { DatePicker, Form, Input, Select, TimePicker } from "antd";
import { LuX } from "react-icons/lu";
import { useSocket } from "../../socket";
import useGlobalStore from "../../store/useGlobalStore";
import useSymbolStore from "../../store/useSymbolStore";
import Icon from "../common/Icon";
import Volume from "./components/Volume";
import PriceInput from "./components/PriceInput";
import SLTPInput from "./components/SLTPInput";
import BuySell from "./components/BuySell";
import { getOptions } from "../../utils/utils";
import { EXPIRATION_POLICY, ORDER_TYPES } from "../../utils/constants";
import { getSide, validateOrderField } from "../../utils/validation";
import { SOCKET_EVENTS } from "../../socket/events";
import useSymbolLive from "../../hooks/useSymbolLive";

const Order = ({ symbolId, showHeader = true }) => {
  const { sendEvent } = useSocket();
  const [form] = Form.useForm();
  const togglePanel = useGlobalStore((state) => state.togglePanel);
  const storedVolume = useGlobalStore((state) => state.volume?.[symbolId]);
  const symbolDeatils = useSymbolStore((state) => state.symbols?.[symbolId]);
  const [formValues, setFormValues] = useState({ type: 0 });
  const [errors, setErrors] = useState({});
  const isLive = useSymbolLive(symbolId);

  const minVolume = storedVolume || symbolDeatils?.min_value || 0.01;

  const defaultValues = {
    volume: minVolume,
    order_price: null,
    stop_loss: null,
    take_profit: null,
    expiration_policy: 0,
    expiry_at: null,
    comment: null,
  };

  const hasErrors = Object.values(errors).some((val) => val === true);

  const resetForm = () => {
    const updatedValues = { ...formValues, ...defaultValues };
    setFormValues(updatedValues);
    form.setFieldsValue(updatedValues);
  };

  useEffect(() => {
    resetForm();
  }, [symbolId, formValues?.type]);

  const updateFormValue = (key, value, range) => {
    const error = validateOrderField(key, value, range, formValues);
    setErrors((prev) => ({ ...prev, [key]: error }));
    setFormValues((prev) => ({ ...prev, [key]: value }));
  };

  const updateError = (key, value) => {
    if (errors?.[key] === value) return;
    setErrors({ ...errors, [key]: value });
  };

  useEffect(() => {
    if (
      [2, 3].includes(formValues?.expiration_policy) &&
      !formValues?.expiry_at
    ) {
      const now = dayjs();
      const minutes = now.minute();
      const roundedMinutes = Math.ceil((minutes + 1) / 5) * 5;

      const adjusted = now.minute(roundedMinutes).second(0);
      const defaultExpiry = adjusted.unix();

      updateFormValue("expiry_at", defaultExpiry);
      form.setFieldsValue({ expiry_at: adjusted });
    }
  }, [formValues?.expiration_policy]);

  const createOrder = (side = 0) => {
    if (hasErrors || !isLive) return;

    const payload = {
      ...formValues,
      ...(formValues?.type === 0 && { order_price: 1 }),
      side: getSide(formValues?.type, side),
      symbol_id: symbolId,
      fill_policy: 0,
    };

    sendEvent(SOCKET_EVENTS.ORDER_CREATE, payload);
  };

  return (
    <div className="h-full w-full overflow-auto scrollbar-hide">
      {showHeader && (
        <div className="sticky top-0 z-10 bg-theme-bg flex items-center justify-between gap-2 p-2 border-b border-theme-border">
          <span>{symbolDeatils?.symbol}</span>

          <button onClick={() => togglePanel("order", false)}>
            <Icon Icon={LuX} size={16} />
          </button>
        </div>
      )}

      <div className="p-2">
        <Form
          form={form}
          layout="vertical"
          size="medium"
          className="custom-form"
          initialValues={formValues}
        >
          <Form.Item label={"Type"} name="type">
            <Select
              value={formValues?.type}
              onChange={(val) => updateFormValue("type", val)}
              options={getOptions(ORDER_TYPES)}
            />
          </Form.Item>

          <Volume
            name={"volume"}
            label={"Lot"}
            symbolId={symbolId}
            value={formValues?.volume}
            setValue={(val, range) => updateFormValue("volume", val, range)}
          />

          {formValues?.type !== 0 && (
            <PriceInput
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
          )}

          <div className="grid grid-cols-2 gap-5">
            <SLTPInput
              name={"stop_loss"}
              label={"SL"}
              symbolId={symbolId}
              data={formValues}
              value={formValues?.stop_loss}
              setValue={(val, range) =>
                updateFormValue("stop_loss", val, range)
              }
              setError={(error) => updateError("stop_loss", error)}
            />

            <SLTPInput
              name={"take_profit"}
              label={"TP"}
              symbolId={symbolId}
              data={formValues}
              value={formValues?.take_profit}
              setValue={(val, range) =>
                updateFormValue("take_profit", val, range)
              }
              setError={(error) => updateError("take_profit", error)}
            />
          </div>

          {formValues?.type !== 0 && (
            <Form.Item label={"Expiration Policy"} name="expiration_policy">
              <Select
                value={formValues?.expiration_policy}
                onChange={(val) => updateFormValue("expiration_policy", val)}
                options={getOptions(EXPIRATION_POLICY)}
              />
            </Form.Item>
          )}

          {formValues?.expiration_policy === 2 && (
            <Form.Item name="expiry_at" label="Select Time">
              <TimePicker
                use12Hours
                format="h:mm A"
                minuteStep={5}
                value={
                  formValues?.expiry_at
                    ? dayjs.unix(formValues.expiry_at)
                    : null
                }
                onChange={(val) => {
                  const timestamp = val ? dayjs(val).unix() : null;
                  updateFormValue("expiry_at", timestamp);
                }}
                disabledTime={() => {
                  const now = dayjs();
                  return {
                    disabledHours: () => [...Array(now.hour()).keys()],
                    disabledMinutes: (selectedHour) =>
                      selectedHour === now.hour()
                        ? [...Array(now.minute()).keys()]
                        : [],
                  };
                }}
                className="!w-full"
              />
            </Form.Item>
          )}

          {formValues?.expiration_policy === 3 && (
            <Form.Item name="expiry_at" label="Select Date & Time">
              <DatePicker
                showTime={{ minuteStep: 5 }}
                format="YYYY-MM-DD h:mm A"
                value={
                  formValues?.expiry_at
                    ? dayjs.unix(formValues.expiry_at)
                    : null
                }
                onChange={(val) => {
                  const timestamp = val ? dayjs(val).unix() : null;
                  updateFormValue("expiry_at", timestamp);
                }}
                disabledDate={(current) => {
                  return current && current < dayjs().startOf("day");
                }}
                disabledTime={(date) => {
                  if (!date) return {};
                  const now = dayjs();
                  if (date.isSame(now, "day")) {
                    return {
                      disabledHours: () => [...Array(now.hour()).keys()],
                      disabledMinutes: (selectedHour) =>
                        selectedHour === now.hour()
                          ? [...Array(now.minute()).keys()]
                          : [],
                    };
                  }
                  return {};
                }}
                className="!w-full"
              />
            </Form.Item>
          )}

          <Form.Item label={"Comment"} name="comment">
            <Input
              placeholder="Comment"
              onChange={(e) => updateFormValue("comment", e.target.value)}
            />
          </Form.Item>

          {formValues?.type === 0 ? (
            <BuySell
              symbolId={symbolId}
              formValues={formValues}
              disabled={errors?.volume || !isLive}
              createOrder={createOrder}
            />
          ) : (
            <button
              className="btn-primary w-full"
              onClick={createOrder}
              disabled={hasErrors || !isLive}
            >
              Place Order
            </button>
          )}
        </Form>
      </div>
    </div>
  );
};

export default Order;

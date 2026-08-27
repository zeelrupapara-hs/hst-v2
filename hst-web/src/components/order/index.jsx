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
import { EXPIRATION_POLICY, ORDER_TYPES, FILL_POLICY } from "../../utils/constants";
import { ORDER_FLAG, EXPIR_FLAG, FILL_FLAG, hasFlag, canBuy, canSell } from "../../utils/symbol";
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

  // what the admin allowed on this symbol (Trade tab: Orders, Expiration, Filling)
  const orderFlags = symbolDeatils?.order_flags ?? 0;
  const expirFlags = symbolDeatils?.expir_flags ?? 0;
  const fillFlags = symbolDeatils?.fill_flags ?? 0;
  const execMode = symbolDeatils?.exec_mode ?? 0;
  const typeAllowed = { 0: ORDER_FLAG.MARKET, 1: ORDER_FLAG.LIMIT, 2: ORDER_FLAG.STOP, 3: ORDER_FLAG.LIMIT, 4: ORDER_FLAG.STOP };
  const typeOptions = getOptions(ORDER_TYPES).filter((o) => hasFlag(orderFlags, typeAllowed[o.value]));
  const expiryAllowed = { 0: EXPIR_FLAG.GTC, 1: EXPIR_FLAG.DAY, 2: EXPIR_FLAG.SPECIFIED, 3: EXPIR_FLAG.SPECIFIED_DAY };
  const expiryOptions = getOptions(EXPIRATION_POLICY).filter((o) => hasFlag(expirFlags, expiryAllowed[o.value]));
  // the filling policy is the trader's choice for market orders only; instant and request execution are always fill-or-kill
  const fillOptions =
    execMode === 0 || execMode === 1
      ? [{ value: 0, label: FILL_POLICY[0] }]
      : getOptions(FILL_POLICY).filter((o) => (o.value === 0 && hasFlag(fillFlags, FILL_FLAG.FOK)) || (o.value === 1 && hasFlag(fillFlags, FILL_FLAG.IOC)) || (o.value === 2 && execMode === 3));
  const allowSL = hasFlag(orderFlags, ORDER_FLAG.SL);
  const allowTP = hasFlag(orderFlags, ORDER_FLAG.TP);
  const buyAllowed = canBuy(symbolDeatils);
  const sellAllowed = canSell(symbolDeatils);

  useEffect(() => {
    if (typeOptions.length && !typeOptions.some((o) => o.value === formValues?.type)) updateFormValue("type", typeOptions[0].value);
    if (expiryOptions.length && !expiryOptions.some((o) => o.value === formValues?.expiration_policy)) updateFormValue("expiration_policy", expiryOptions[0].value);
    if (fillOptions.length && !fillOptions.some((o) => o.value === (formValues?.fill_policy ?? 0))) updateFormValue("fill_policy", fillOptions[0].value);
  }, [orderFlags, expirFlags, fillFlags, execMode]);

  const minVolume = storedVolume || symbolDeatils?.min_value || 0.01;

  const defaultValues = {
    volume: minVolume,
    order_price: null,
    stop_loss: null,
    take_profit: null,
    expiration_policy: expiryOptions[0]?.value ?? 0,
    expiry_at: null,
    fill_policy: fillOptions[0]?.value ?? 0,
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
      fill_policy: formValues?.type === 0 ? formValues?.fill_policy ?? 0 : 0,
      ...(!allowSL && { stop_loss: null }),
      ...(!allowTP && { take_profit: null }),
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
              options={typeOptions}
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

          {formValues?.type === 0 && fillOptions.length > 1 && (
            <Form.Item label={"Filling"} name="fill_policy">
              <Select
                value={formValues?.fill_policy ?? 0}
                onChange={(val) => updateFormValue("fill_policy", val)}
                options={fillOptions}
              />
            </Form.Item>
          )}

          <div className="grid grid-cols-2 gap-5">
            {allowSL && (
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
            )}

            {allowTP && (
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
            )}
          </div>

          {formValues?.type !== 0 && (
            <Form.Item label={"Expiration Policy"} name="expiration_policy">
              <Select
                value={formValues?.expiration_policy}
                onChange={(val) => updateFormValue("expiration_policy", val)}
                options={expiryOptions}
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
              disableBuy={!buyAllowed}
              disableSell={!sellAllowed}
              createOrder={createOrder}
            />
          ) : (
            <button
              className="btn-primary w-full"
              onClick={createOrder}
              disabled={hasErrors || !isLive || !typeOptions.length || ([1, 2].includes(formValues?.type) ? !buyAllowed : !sellAllowed)}
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

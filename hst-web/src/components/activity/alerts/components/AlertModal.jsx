import { useEffect, useState } from "react";
import { Form, Select } from "antd";
import { LuX } from "react-icons/lu";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";
import useSymbolStore from "../../../../store/useSymbolStore";
import useLiveSymbolStore from "../../../../store/useLiveSymbolStore";
import Bid from "../../../marketWatch/components/Bid";
import Ask from "../../../marketWatch/components/Ask";
import Value from "../../../common/Value";
import { createAlert, updateAlert } from "../../../../api/request/alert";
import { errorToast } from "../../../common/CustomToast";

const AlertModal = ({ isOpen, setIsOpen, mode = "add", record }) => {
  const [form] = Form.useForm();
  const symbols = useSymbolStore((state) => state.symbols);
  const getSymbolInfo = useLiveSymbolStore((state) => state.getSymbolInfo);
  const [formValues, setFormValues] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  const isSymbolType = ["market_ask", "market_bid"].includes(formValues?.type);
  const requiredFields = ["type", "condition", "value"];
  const hasError =
    requiredFields.some((field) => !formValues?.[field]) ||
    (isSymbolType && !formValues?.symbol);

  useEffect(() => {
    if (record) {
      const [type, condition, value, symbol] = record?.formula?.split(",");
      setFormValues({ type, condition, value, symbol: Number(symbol) });
    }
  }, [record, isOpen]);

  const updateFormValue = (key, value) => {
    setFormValues({ ...formValues, [key]: value });
  };

  const typeOptions = [
    { value: "market_ask", label: "Market Ask" },
    { value: "market_bid", label: "Market Bid" },
    { value: "balance", label: "Balance" },
    { value: "equity", label: "Equity" },
    { value: "margin_level", label: "Margin Level" },
  ];

  const conditionOptions = [
    { value: "greater_than", label: "Greater Than" },
    { value: "less_than", label: "Less Than" },
  ];

  const symbolOptions = Object.values(symbols)?.map((symbol) => ({
    value: symbol?.id,
    label: symbol?.symbol,
  }));

  const getNewValue = (newValue) => {
    const { last_ask } = getSymbolInfo(formValues?.symbol);

    if (isSymbolType && !formValues?.value && last_ask) return last_ask;
    else return newValue;
  };

  const getButtonText = () => {
    if (isLoading) return mode === "edit" ? "Saving..." : "Adding...";
    return mode === "edit" ? "Save" : "Add";
  };

  const clearAll = () => {
    form.resetFields();
    setFormValues({});
  };

  const handleSubmit = async () => {
    if (hasError) return;

    setIsLoading(true);
    try {
      const { type, symbol, condition, value } = formValues;
      const symbolId = isSymbolType ? symbol : "";
      const payload = { formula: `${type},${condition},${value},${symbolId}` };

      let data = null;

      if (mode === "edit" && record) {
        const res = await updateAlert(record?.id, payload);
        data = res?.data;
      } else {
        const res = await createAlert(payload);
        data = res?.data;
      }

      if (data?.success) {
        clearAll();
        setIsOpen(false);
      }
    } catch (error) {
      errorToast(error?.response?.data?.message || error?.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    clearAll();
    setIsOpen(false);
  };

  return (
    <ModalComponent isOpen={isOpen}>
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>{mode === "edit" ? "Edit" : "Add"} Alert</span>

        <button onClick={handleClose}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        {formValues?.symbol && (
          <div className="grid grid-cols-2 gap-5 mb-5">
            <div className="flex flex-col items-center gap-2">
              <span>Bid</span>
              <Bid symbolId={formValues?.symbol} />
            </div>

            <div className="flex flex-col items-center gap-2">
              <span>Ask</span>
              <Ask symbolId={formValues?.symbol} />
            </div>
          </div>
        )}

        <Form
          form={form}
          layout="vertical"
          size="medium"
          className="custom-form"
          initialValues={mode === "edit" ? formValues : {}}
        >
          <div className="grid grid-cols-2 gap-x-5">
            <Form.Item name="type" label="Type">
              <Select
                placeholder="Select Type"
                options={typeOptions}
                onChange={(value) => updateFormValue("type", value)}
              />
            </Form.Item>

            {isSymbolType && (
              <Form.Item name="symbol" label="Symbol">
                <Select
                  placeholder="Select symbol"
                  options={symbolOptions}
                  showSearch
                  optionFilterProp="label"
                  filterOption={(input, option) =>
                    option?.label?.toLowerCase().includes(input?.toLowerCase())
                  }
                  onChange={(value) => updateFormValue("symbol", value)}
                />
              </Form.Item>
            )}

            <Form.Item name="condition" label="Condition">
              <Select
                placeholder="Select Condition"
                options={conditionOptions}
                onChange={(value) => updateFormValue("condition", value)}
              />
            </Form.Item>

            {formValues?.type && (
              <Form.Item name="value">
                <div>
                  <div className="pb-1 text-xs leading-[22px]">
                    <span>Value</span>
                  </div>

                  <Value
                    min={0.01}
                    max={Infinity}
                    value={formValues?.value}
                    setValue={(val, inputType) => {
                      const newValue =
                        inputType === "manual" ? val : getNewValue(val);
                      updateFormValue("value", newValue);
                    }}
                    className="h-[32px]"
                  />
                </div>
              </Form.Item>
            )}
          </div>

          <button
            className="btn-primary w-full"
            onClick={handleSubmit}
            disabled={hasError || isLoading}
          >
            {getButtonText()}
          </button>
        </Form>
      </div>
    </ModalComponent>
  );
};

export default AlertModal;

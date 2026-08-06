import { useState } from "react";
import { DatePicker, Form, Select } from "antd";
import { LuX } from "react-icons/lu";
import dayjs from "dayjs";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";
import useAuthStore from "../../../../store/useAuthStore";
import { createReport } from "../../../../api/request/report";
import { errorToast } from "../../../common/CustomToast";

const ReportModal = ({ isOpen, setIsOpen }) => {
  const [form] = Form.useForm();
  const user = useAuthStore((state) => state.user);
  const [formValues, setFormValues] = useState({});
  const [isLoading, setIsLoading] = useState(false);

  const requiredFields = ["type", "format", "from", "to"];
  const hasError = requiredFields.some((field) => !formValues?.[field]);

  const updateFormValue = (key, value) => {
    setFormValues({ ...formValues, [key]: value });
  };

  const typeOptions = [
    { value: 10, label: "Account Statement" },
    { value: 6, label: "Money Transaction" },
    { value: 11, label: "History" },
    { value: 12, label: "Journal" },
  ];

  const formatOptions = [
    { value: 1, label: "HTML" },
    { value: 2, label: "CSV" },
    { value: 3, label: "PDF" },
  ];

  const historyTypeOptions = [
    { value: 2, label: "Position" },
    { value: 0, label: "Order" },
    { value: 1, label: "Deal" },
  ];

  const clearAll = () => {
    form.resetFields();
    setFormValues({});
  };

  const handleSubmit = async () => {
    if (hasError) return;

    setIsLoading(true);
    try {
      const payload = {
        acc_id: user?.account_id,
        borker_id: 1,
        report_type: formValues?.type,
        report_data_type: formValues?.format,
        filter: {
          fromdate: formValues?.from?.format("YYYY-MM-DD"),
          todate: formValues?.to?.format("YYYY-MM-DD"),
        },
        history: { history_type: formValues?.history_type || 0 },
      };

      const { data } = await createReport(payload);
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
        <span>Add Report</span>

        <button onClick={handleClose}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        <Form
          form={form}
          layout="vertical"
          size="medium"
          className="custom-form"
          // initialValues={formValues}
        >
          <Form.Item name="type" label="Report Type">
            <Select
              placeholder="Select Type"
              options={typeOptions}
              onChange={(value) => updateFormValue("type", value)}
            />
          </Form.Item>

          {formValues?.type === 3 && (
            <Form.Item name="history_type" label="History Type">
              <Select
                placeholder="Select History Type"
                options={historyTypeOptions}
                onChange={(value) => updateFormValue("history_type", value)}
              />
            </Form.Item>
          )}

          <Form.Item name="format" label="Report Format">
            <Select
              placeholder="Select Format"
              options={formatOptions}
              onChange={(value) => updateFormValue("format", value)}
            />
          </Form.Item>

          <Form.Item name="from" label="Start Date">
            <DatePicker
              className="w-full"
              onChange={(date) => updateFormValue("from", date)}
              disabledDate={(current) =>
                current && current > dayjs().endOf("day")
              }
            />
          </Form.Item>

          <Form.Item name="to" label="End Date">
            <DatePicker
              className="w-full"
              onChange={(date) => updateFormValue("to", date)}
              disabledDate={(current) =>
                current && current > dayjs().endOf("day")
              }
            />
          </Form.Item>

          <button
            className="btn-primary w-full"
            onClick={handleSubmit}
            disabled={hasError || isLoading}
          >
            {isLoading ? "Saving..." : "Save"}
          </button>
        </Form>
      </div>
    </ModalComponent>
  );
};

export default ReportModal;

import { useEffect, useState } from "react";
import { Form, Input, Select } from "antd";
import { LuX } from "react-icons/lu";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";
import Editor from "../../../common/Editor";
import { createMail, updateMail } from "../../../../api/request/mail";
import { errorToast } from "../../../common/CustomToast";
import ConfirmationModal from "../../../modal/ConfirmationModal";

const MailModal = ({ isOpen, setIsOpen, mode = "add", record }) => {
  const [form] = Form.useForm();
  const [formValues, setFormValues] = useState({});
  const [isLoading, setIsLoading] = useState(false);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [isConfirmLoading, setIsConfirmLoading] = useState(false);

  const requiredFields = ["subject", "body"];
  const hasError = requiredFields.some((field) => !formValues?.[field]);

  useEffect(() => {
    if (record) {
      const { to_account_id: to, subject, body } = record;
      setFormValues({ to, subject, body });
    }
  }, [record, isOpen]);

  const updateFormValue = (key, value) => {
    setFormValues({ ...formValues, [key]: value });
  };

  const getButtonText = () => {
    if (isLoading) return mode === "edit" ? "Saving..." : "Sending...";
    return mode === "edit" ? "Save" : "Send";
  };

  const clearAll = () => {
    form.resetFields();
    setFormValues({});
  };

  const handleSubmit = async () => {
    if (hasError) return;

    setIsLoading(true);
    try {
      const payload = { ...formValues, to: [25100007], status: 1 };

      let data = null;

      if (mode === "edit" && record) {
        const res = await updateMail(record?.id, payload);
        data = res?.data;
      } else {
        const res = await createMail(payload);
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
    const isMailData = Object.keys(formValues).some((key) => formValues?.[key]);
    if (mode === "add" && isMailData) return setIsConfirmOpen(true);

    clearAll();
    setIsOpen(false);
  };

  const handleConfirm = async () => {
    setIsConfirmLoading(true);

    try {
      const payload = { ...formValues, to: [25100007], status: 3 };
      const { data } = await createMail(payload);
      if (data?.success) {
        clearAll();
        setIsConfirmOpen(false);
        setIsOpen(false);
      }
    } catch (error) {
      errorToast(error?.response?.data?.message || error?.message);
    } finally {
      setIsConfirmLoading(false);
    }
  };

  return (
    <>
      <ModalComponent isOpen={isOpen} width={800}>
        <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
          <span>{mode === "edit" ? "Edit" : "Send"} Mail</span>

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
            initialValues={mode === "edit" ? formValues : { to: 25100007 }}
          >
            <div className="grid grid-cols-2 gap-5">
              <Form.Item name="to" label="Recipient">
                <Select
                  placeholder="Select Recipient"
                  options={[{ value: 25100007, label: "Admin" }]}
                  onChange={(value) => updateFormValue("to", value)}
                  disabled
                />
              </Form.Item>

              <Form.Item name="subject" label="Subject">
                <Input
                  placeholder="Subject"
                  onChange={(e) => updateFormValue("subject", e.target.value)}
                />
              </Form.Item>
            </div>

            <Form.Item name="body" label="Message">
              <Editor
                value={formValues?.body || ""}
                onChange={(data) => updateFormValue("body", data)}
              />
            </Form.Item>

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

      <ConfirmationModal
        isOpen={isConfirmOpen}
        setIsOpen={setIsConfirmOpen}
        title="Discard changes?"
        message="Are you sure you want to move this mail to draft?"
        onConfirm={handleConfirm}
        isLoading={isConfirmLoading}
      />
    </>
  );
};

export default MailModal;

import { useEffect, useState } from "react";
import { Form, Input, Select, Upload } from "antd";
import { LuPaperclip, LuX } from "react-icons/lu";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";
import Editor from "../../../common/Editor";
import { createMail, updateMail, getMailboxes, uploadMailAttachments } from "../../../../api/request/mail";
import { errorToast } from "../../../common/CustomToast";
import ConfirmationModal from "../../../modal/ConfirmationModal";

// MT5 attachment limits: 5 files, 8MB each, 16MB in total, whitelisted types
const ATTACH_MAX_FILES = 5;
const ATTACH_MAX_FILE = 8 * 1024 * 1024;
const ATTACH_MAX_TOTAL = 16 * 1024 * 1024;
const ATTACH_EXTS = [
  "png", "jpg", "jpeg", "bmp", "gif", "zip", "7z", "doc", "xls",
  "docx", "xlsx", "odt", "rtf", "csv", "txt", "log",
];

const checkAttachments = (files) => {
  if (files.length > ATTACH_MAX_FILES) return `Up to ${ATTACH_MAX_FILES} files can be attached`;
  let total = 0;
  for (const f of files) {
    if (!ATTACH_EXTS.includes(f.name.split(".").pop().toLowerCase())) {
      return `"${f.name}" cannot be attached`;
    }
    if (f.size > ATTACH_MAX_FILE) return `"${f.name}" is over 8MB`;
    total += f.size;
  }
  return total > ATTACH_MAX_TOTAL ? "Attachments exceed 16MB in total" : "";
};

const MailModal = ({ isOpen, setIsOpen, mode = "add", record, replyTo }) => {
  const [form] = Form.useForm();
  const [formValues, setFormValues] = useState({});
  const [mailboxes, setMailboxes] = useState([]);
  const [fileList, setFileList] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);

  const requiredFields = ["to", "subject", "body"];
  const hasError = requiredFields.some((field) => !formValues?.[field]);

  useEffect(() => {
    if (!isOpen) return;
    getMailboxes()
      .then((r) => setMailboxes(r?.data?.data || []))
      .catch(() => setMailboxes([]));
  }, [isOpen]);

  useEffect(() => {
    if (record) {
      const { to_account_id: to, subject, body } = record;
      const values = { to: to || undefined, subject, body };
      setFormValues(values);
      form.setFieldsValue(values);
    } else if (replyTo) {
      // a reply is locked to the original sender and quotes the message under a rule
      const values = {
        to: replyTo.sender_login,
        subject: /^re:/i.test(replyTo.subject) ? replyTo.subject : `Re: ${replyTo.subject}`,
        body: `<p><br></p><hr><blockquote>${replyTo.body}</blockquote>`,
      };
      setFormValues(values);
      form.setFieldsValue(values);
    }
  }, [record, replyTo, isOpen, form]);

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
    setFileList([]);
  };

  const stageAttachments = async () => {
    if (!fileList.length) return [];
    const files = fileList.map((f) => f.originFileObj || f);
    const bad = checkAttachments(files);
    if (bad) throw new Error(bad);
    const res = await uploadMailAttachments(files);
    return (res?.data?.data || []).map((a) => a.attachment_id);
  };

  const handleSubmit = async () => {
    if (hasError) return;

    setIsLoading(true);
    try {
      const payload = {
        ...formValues,
        status: 1,
        replyTo: replyTo?.tracking_id,
        attachmentIds: await stageAttachments(),
      };

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

  // discarding really discards; a draft is only kept when the trader saves one on purpose
  const handleConfirm = () => {
    clearAll();
    setIsConfirmOpen(false);
    setIsOpen(false);
  };

  return (
    <>
      <ModalComponent isOpen={isOpen} width={800}>
        <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
          <span>{mode === "edit" ? "Edit" : replyTo ? "Reply" : "Send"} Mail</span>

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
          >
            <div className="grid grid-cols-2 gap-5">
              <Form.Item name="to" label="To">
                <Select
                  placeholder={mailboxes.length ? "Select mailbox" : "No mailboxes available"}
                  options={mailboxes.map((m) => ({ value: m.login, label: m.mailbox }))}
                  onChange={(value) => updateFormValue("to", value)}
                  disabled={Boolean(replyTo) || !mailboxes.length}
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

            <Form.Item label="Attachments">
              <Upload
                multiple
                fileList={fileList}
                beforeUpload={() => false}
                accept={ATTACH_EXTS.map((e) => `.${e}`).join(",")}
                onChange={({ fileList: next }) => {
                  const bad = checkAttachments(next.map((f) => f.originFileObj || f));
                  if (bad) return errorToast(bad);
                  setFileList(next);
                }}
                onRemove={(file) => setFileList(fileList.filter((f) => f.uid !== file.uid))}
              >
                <button type="button" className="flex items-center gap-1">
                  <Icon Icon={LuPaperclip} size={16} /> Attach files
                </button>
              </Upload>
            </Form.Item>

            <div className="flex justify-end gap-3">
              <button type="button" className="btn-outline min-w-24" onClick={handleClose}>
                Cancel
              </button>
              <button
                className={`btn-primary min-w-24 ${hasError || isLoading || !mailboxes.length ? "btn-disabled" : ""}`}
                onClick={handleSubmit}
                disabled={hasError || isLoading || !mailboxes.length}
              >
                {getButtonText()}
              </button>
            </div>
          </Form>
        </div>
      </ModalComponent>

      <ConfirmationModal
        isOpen={isConfirmOpen}
        setIsOpen={setIsConfirmOpen}
        title="Discard changes?"
        message="This message has not been sent. Discard it?"
        onConfirm={handleConfirm}
      />
    </>
  );
};

export default MailModal;

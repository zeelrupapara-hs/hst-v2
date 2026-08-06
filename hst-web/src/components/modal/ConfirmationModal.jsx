import { Modal } from "antd";
import { LuX } from "react-icons/lu";
import Icon from "../common/Icon";

const ConfirmationModal = ({
  isOpen = false,
  setIsOpen,
  width,
  title = "Confirmation",
  message = "Are you sure?",
  onCancel,
  onConfirm,
  isLoading,
  children,
}) => {
  const handleConfirm = () => {
    onConfirm?.();
    setIsOpen(false);
  };

  const handleCancel = () => {
    onCancel?.();
    setIsOpen(false);
  };

  return (
    <Modal
      open={isOpen}
      closeIcon={false}
      footer={false}
      width={width}
      centered
      className="custom-modal bg-theme-bg border border-theme-border shadow-lg rounded-md"
    >
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>{title}</span>

        <button onClick={() => setIsOpen(false)}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        <p className="text-base">{message}</p>

        {children && <div className="mt-5">{children}</div>}

        <div className="flex justify-end items-center gap-3 mt-5">
          <button
            className="btn-outline"
            onClick={handleCancel}
            disabled={isLoading}
          >
            No
          </button>

          <button
            className="btn-primary"
            onClick={handleConfirm}
            disabled={isLoading}
          >
            Yes
          </button>
        </div>
      </div>
    </Modal>
  );
};

export default ConfirmationModal;

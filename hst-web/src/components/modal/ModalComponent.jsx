import { Modal } from "antd";

const ModalComponent = ({ isOpen = false, width, children }) => {
  return (
    <Modal
      open={isOpen}
      closeIcon={false}
      footer={false}
      width={width}
      centered
      className="custom-modal bg-theme-bg border border-theme-border shadow-lg rounded-md"
    >
      {children}
    </Modal>
  );
};

export default ModalComponent;

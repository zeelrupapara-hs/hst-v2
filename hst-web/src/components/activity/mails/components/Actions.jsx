import { useState } from "react";
import { LuEye, LuPencil, LuTrash2 } from "react-icons/lu";
import Icon from "../../../common/Icon";
import { deleteMail } from "../../../../api/request/mail";
import { errorToast } from "../../../common/CustomToast";
import MailModal from "./MailModal";
import PreviewModal from "./PreviewModal";

const Actions = ({ record, activeTab }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);

  const handleEdit = () => {
    setIsModalOpen(true);
  };

  const handleDelete = async () => {
    if (!record) return;

    setIsLoading(true);
    try {
      await deleteMail(record?.id);
    } catch (error) {
      errorToast(error?.response?.data?.message || error?.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleView = () => {
    setIsPreviewOpen(true);
  };

  return (
    <>
      <div className="flex items-center gap-2">
        {activeTab === 3 && (
          <button onClick={handleEdit}>
            <Icon Icon={LuPencil} size={18} />
          </button>
        )}

        <button onClick={handleDelete}>
          <Icon Icon={LuTrash2} size={18} />
        </button>

        <button onClick={handleView}>
          <Icon Icon={LuEye} size={18} />
        </button>
      </div>

      <MailModal
        isOpen={isModalOpen}
        setIsOpen={setIsModalOpen}
        mode={"edit"}
        record={record}
      />

      <PreviewModal
        isOpen={isPreviewOpen}
        setIsOpen={setIsPreviewOpen}
        data={record}
      />
    </>
  );
};

export default Actions;

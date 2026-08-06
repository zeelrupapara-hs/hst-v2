import { useState } from "react";
import { LuPencil, LuTrash2 } from "react-icons/lu";
import Icon from "../../../common/Icon";
import { deleteAlert } from "../../../../api/request/alert";
import { errorToast } from "../../../common/CustomToast";
import AlertModal from "./AlertModal";

const Actions = ({ record }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleEdit = () => {
    setIsModalOpen(true);
  };

  const handleDelete = async () => {
    if (!record) return;

    setIsLoading(true);
    try {
      await deleteAlert(record?.id);
    } catch (error) {
      errorToast(error?.response?.data?.message || error?.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <div className="flex items-center gap-2">
        {record?.status === 0 && (
          <button onClick={handleEdit}>
            <Icon Icon={LuPencil} size={18} />
          </button>
        )}

        <button onClick={handleDelete}>
          <Icon Icon={LuTrash2} size={18} />
        </button>
      </div>

      <AlertModal
        isOpen={isModalOpen}
        setIsOpen={setIsModalOpen}
        mode={"edit"}
        record={record}
      />
    </>
  );
};

export default Actions;

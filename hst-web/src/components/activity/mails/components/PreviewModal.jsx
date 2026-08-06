import { LuX } from "react-icons/lu";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";

const PreviewModal = ({ isOpen, setIsOpen, data }) => {
  return (
    <ModalComponent isOpen={isOpen} width={800}>
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>Preview</span>

        <button onClick={() => setIsOpen(false)}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        <div className="flex flex-col gap-5">
          <div className="p-2 border border-theme-border rounded">
            <p className="font-semibold mb-1">Subject: </p>
            <p>{data?.subject || ""}</p>
          </div>

          <div className="p-2 border border-theme-border rounded">
            <p className="font-semibold mb-1">Message:</p>
            <div
              className="prose dark:prose-invert"
              dangerouslySetInnerHTML={{ __html: data?.body || "" }}
            />
          </div>
        </div>
      </div>
    </ModalComponent>
  );
};

export default PreviewModal;

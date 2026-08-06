import { useState } from "react";
import { LuDownload, LuTrash2 } from "react-icons/lu";
import Icon from "../../../common/Icon";
import { deleteReport, downloadReport } from "../../../../api/request/report";
import { errorToast } from "../../../common/CustomToast";

const Actions = ({ record }) => {
  const [isLoading, setIsLoading] = useState(false);

  const handleDownload = async () => {
    try {
      const headers = { responseType: "blob" };
      const { data } = await downloadReport(record?.request_id, headers);

      const url = window.URL.createObjectURL(new Blob([data]));
      const link = document.createElement("a");
      link.href = url;

      const filename = record?.url;

      link.setAttribute("download", filename);
      document.body.appendChild(link);
      link.click();

      link.remove();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Report download error:", error);
    }
  };

  const handleDelete = async () => {
    if (!record) return;

    setIsLoading(true);
    try {
      await deleteReport(record?.id);
    } catch (error) {
      errorToast(error?.response?.data?.message || error?.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <div className="flex items-center gap-2">
        <button onClick={handleDownload}>
          <Icon Icon={LuDownload} size={18} />
        </button>

        <button onClick={handleDelete}>
          <Icon Icon={LuTrash2} size={18} />
        </button>
      </div>
    </>
  );
};

export default Actions;

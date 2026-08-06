import { useState } from "react";
import { LuCalendarPlus } from "react-icons/lu";
import useReportStore from "../../../store/useReportStore";
import useReports from "../../../columns/useReports";
import SimpleTable from "../../table/SimpleTable";
import Icon from "../../common/Icon";
import ReportModal from "./components/ReportModal";

const Reports = () => {
  const reports = useReportStore((state) => state.reports);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const baseColumns = useReports();

  const columns = baseColumns.map((col) =>
    col.key === "actions"
      ? {
          ...col,
          title: (
            <div className="flex items-center justify-between">
              <span>{col?.title}</span>

              <button onClick={() => setIsModalOpen(true)}>
                <Icon Icon={LuCalendarPlus} size={18} />
              </button>
            </div>
          ),
        }
      : col
  );

  return (
    <>
      <SimpleTable columns={columns} data={reports} />

      <ReportModal isOpen={isModalOpen} setIsOpen={setIsModalOpen} />
    </>
  );
};

export default Reports;

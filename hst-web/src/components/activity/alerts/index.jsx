import { useState } from "react";
import { LuBellPlus } from "react-icons/lu";
import useAlertStore from "../../../store/useAlertStore";
import useAlerts from "../../../columns/useAlerts";
import SimpleTable from "../../table/SimpleTable";
import Icon from "../../common/Icon";
import AlertModal from "./components/AlertModal";

const Alerts = () => {
  const alerts = useAlertStore((state) => state.alerts);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const baseColumns = useAlerts();

  const columns = baseColumns.map((col) =>
    col.key === "actions"
      ? {
          ...col,
          title: (
            <div className="flex items-center justify-between">
              <span>{col?.title}</span>

              <button onClick={() => setIsModalOpen(true)}>
                <Icon Icon={LuBellPlus} size={18} />
              </button>
            </div>
          ),
        }
      : col
  );

  return (
    <>
      <SimpleTable columns={columns} data={alerts} />

      <AlertModal
        isOpen={isModalOpen}
        setIsOpen={setIsModalOpen}
        mode={"add"}
      />
    </>
  );
};

export default Alerts;

import { useMemo } from "react";
import useSymbolStore from "../store/useSymbolStore";
import { defaultRender, formatDate } from "../utils/utils";
import {
  ALERT_CONDITIONS,
  ALERT_STATUS,
  ALERT_TYPES,
} from "../utils/constants";
import Actions from "../components/activity/alerts/components/Actions";

const useAlerts = () => {
  const symbols = useSymbolStore((state) => state.symbols);

  const columns = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      render: defaultRender,
    },
    {
      title: "Symbol",
      dataIndex: "symbol_id",
      key: "symbol_id",
      render: (value) => symbols?.[value]?.symbol,
    },
    {
      title: "Condition",
      dataIndex: "condition",
      key: "condition",
      render: (value) => {
        const [type, condition] = value?.split(" ");
        const alertType = ALERT_TYPES[type] || type;
        const alertCondition = ALERT_CONDITIONS[condition] || condition;
        return `${alertType || ""} ${alertCondition || ""}`;
      },
    },
    {
      title: "Trigger Price",
      dataIndex: "trigger_price",
      key: "trigger_price",
      render: defaultRender,
    },
    {
      title: "Trigger Time",
      dataIndex: "created_at",
      key: "created_at",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (value) => ALERT_STATUS[value],
    },
    {
      title: "Actions",
      dataIndex: "actions",
      key: "actions",
      render: (_, record) => <Actions record={record} />,
      width: 150,
    },
  ];

  return useMemo(() => columns, []);
};

export default useAlerts;

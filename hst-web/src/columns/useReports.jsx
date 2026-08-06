import { useMemo } from "react";
import { defaultRender, formatDate } from "../utils/utils";
import { REPORT_STATUS } from "../utils/constants";
import Actions from "../components/activity/reports/components/Actions";

const useReports = () => {
  const columns = [
    {
      title: "Name",
      dataIndex: "desc",
      key: "name",
      render: defaultRender,
      width: 350,
    },
    {
      title: "Date/Time",
      dataIndex: "created_at",
      key: "created_at",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (value) => REPORT_STATUS[value],
    },
    {
      title: "Actions",
      key: "actions",
      dataIndex: "actions",
      render: (_, record) => <Actions record={record} />,
      width: 150,
    },
  ];

  return useMemo(() => columns, []);
};

export default useReports;

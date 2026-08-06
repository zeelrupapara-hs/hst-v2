import { useMemo } from "react";
import { defaultRender, formatDate } from "../utils/utils";
import Actions from "../components/activity/mails/components/Actions";

const useMails = (activeTab) => {
  const columns = [
    {
      title: "Date/Time",
      dataIndex: "created_at",
      key: "created_at",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "Subject",
      dataIndex: "subject",
      key: "subject",
      render: defaultRender,
    },
    {
      title: "Actions",
      key: "actions",
      dataIndex: "actions",
      render: (_, record) => <Actions record={record} activeTab={activeTab} />,
      width: 150,
    },
  ];

  return useMemo(() => columns, [activeTab]);
};

export default useMails;

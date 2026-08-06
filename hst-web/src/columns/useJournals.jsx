import { useMemo } from "react";
import { defaultRender, formatDate } from "../utils/utils";
import { CHANNEL } from "../utils/constants";

const useJournals = () => {
  const columns = [
    {
      title: "Date/Time",
      dataIndex: "created_at",
      key: "created_at",
      render: (value) => formatDate(value),
      width: 165,
    },
    {
      title: "Channel",
      dataIndex: "channel",
      key: "channel",
      render: (value) => CHANNEL[value],
      width: 110,
    },
    {
      title: "OS",
      dataIndex: "os",
      key: "os",
      render: defaultRender,
      width: 110,
    },
    {
      title: "IP Address",
      dataIndex: "ip",
      key: "ip",
      render: defaultRender,
      width: 150,
    },
    {
      // the widest thing in the table, and the only column worth reading at a glance
      title: "Description",
      dataIndex: "desc",
      key: "desc",
      render: defaultRender,
      width: 620,
    },
  ];

  return useMemo(() => columns, []);
};

export default useJournals;

import { useMemo } from "react";
import { LuMail, LuMailOpen, LuPaperclip } from "react-icons/lu";
import { defaultRender, formatDate } from "../utils/utils";
import Icon from "../components/common/Icon";
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
      title: "From",
      dataIndex: "sender_name",
      key: "sender_name",
      render: (value, record) => value || String(record?.sender_login ?? ""),
      width: 120,
    },
    {
      title: "Subject",
      dataIndex: "subject",
      key: "subject",
      render: (value, record) => (
        <span className={`flex items-center gap-1 ${record?.folder === 1 && !record?.read_at ? "font-semibold" : ""}`}>
          {activeTab === 1 && (
            <Icon Icon={record?.read_at ? LuMailOpen : LuMail} size={14} />
          )}
          {record?.attach_id && <Icon Icon={LuPaperclip} size={14} />}
          {defaultRender(value)}
        </span>
      ),
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

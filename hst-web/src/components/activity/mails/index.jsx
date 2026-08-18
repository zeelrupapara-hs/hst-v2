import { useEffect, useState } from "react";
import { Tabs } from "antd";
import { LuMailPlus } from "react-icons/lu";
import useMailStore from "../../../store/useMailStore";
import useMails from "../../../columns/useMails";
import SimpleTable from "../../table/SimpleTable";
import Icon from "../../common/Icon";
import MailModal from "./components/MailModal";

const Mails = () => {
  const { inboxMails, outboxMails, draftMails, trashMails } = useMailStore();
  const [activeTab, setActiveTab] = useState(1);
  const [mails, setMails] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const columns = useMails(activeTab);

  const tabs = [
    { key: 1, label: "Inbox" },
    { key: 2, label: "Outbox" },
    { key: 3, label: "Draft" },
    { key: 4, label: "Trash" },
  ];

  useEffect(() => {
    // replies nest under the newest mail of their thread, antd tree-data style
    const threaded = (rows) => {
      const byThread = new Map();
      for (const m of rows || []) {
        byThread.get(m?.thread_id)?.push(m) ?? byThread.set(m?.thread_id, [m]);
      }
      return [...byThread.values()].map(([head, ...rest]) =>
        rest.length ? { ...head, children: rest } : head
      );
    };

    if (activeTab === 1) setMails(threaded(inboxMails));
    if (activeTab === 2) setMails(threaded(outboxMails));
    if (activeTab === 3) setMails(draftMails);
    if (activeTab === 4) setMails(trashMails);
  }, [activeTab, inboxMails, outboxMails, draftMails, trashMails]);

  return (
    <>
      <div className="h-full flex flex-col">
        <div className="flex items-center justify-between">
          <Tabs
            items={tabs}
            onChange={(tab) => setActiveTab(tab)}
            className="small-tabs"
          />

          <button className="m-3" onClick={() => setIsModalOpen(true)}>
            <Icon Icon={LuMailPlus} size={18} />
          </button>
        </div>

        <div className="h-full">
          <SimpleTable columns={columns} data={mails} />
        </div>
      </div>

      <MailModal isOpen={isModalOpen} setIsOpen={setIsModalOpen} />
    </>
  );
};

export default Mails;

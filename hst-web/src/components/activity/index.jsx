import { useState } from "react";
import { Badge } from "antd";
import useMailStore from "../../store/useMailStore";
import { CiCircleList } from "react-icons/ci";
import { LuGitFork } from "react-icons/lu";
import { VscHistory } from "react-icons/vsc";
import { IoIosNotificationsOutline } from "react-icons/io";
import { IoMailOutline, IoNewspaperOutline } from "react-icons/io5";
import { BsJournals, BsJournalBookmark } from "react-icons/bs";
import Icon from "../common/Icon";
import Positions from "./positions";
import NetTrade from "./netTrade";
import History from "./history";
import Alerts from "./alerts";
import Mails from "./mails";
import Reports from "./reports";
import News from "./news";
import Journals from "./journals";

const Activity = () => {
  const [activeMenu, setActiveMenu] = useState(0);
  const unreadMails = useMailStore(
    (s) => (s.inboxMails || []).filter((m) => !m?.read_at).length
  );

  const menu = [
    {
      id: 0,
      title: "Positions",
      icon: <Icon Icon={CiCircleList} size={28} />,
    },
    {
      id: 1,
      title: "Net Trades",
      icon: <Icon Icon={LuGitFork} size={28} className="!stroke-1" />,
    },
    {
      id: 2,
      title: "History",
      icon: <Icon Icon={VscHistory} size={24} />,
    },
    {
      id: 3,
      title: "Alerts",
      icon: <Icon Icon={IoIosNotificationsOutline} size={28} />,
    },
    {
      id: 4,
      title: "Mails",
      icon: (
        <Badge count={unreadMails} size="small" offset={[4, -2]}>
          <Icon Icon={IoMailOutline} size={24} />
        </Badge>
      ),
    },
    {
      id: 5,
      title: "Reports",
      icon: <Icon Icon={BsJournals} size={24} />,
    },
    {
      id: 6,
      title: "News",
      icon: <Icon Icon={IoNewspaperOutline} size={24} />,
    },
    {
      id: 7,
      title: "Journals",
      icon: <Icon Icon={BsJournalBookmark} size={24} />,
    },
  ];

  return (
    <div className="h-full w-full flex">
      <div className="min-w-fit flex flex-col gap-2 p-2 border-r-4 border-theme-border overflow-auto scrollbar-hide">
        {menu.map((item, index) => (
          <button
            key={index}
            title={item.title}
            className="flex items-center justify-center btn-icon-hover"
            onClick={() => setActiveMenu(item.id)}
          >
            {item.icon}
          </button>
        ))}
      </div>

      <div className="h-full w-full">
        {activeMenu === 0 && <Positions />}
        {activeMenu === 1 && <NetTrade />}
        {activeMenu === 2 && <History />}
        {activeMenu === 3 && <Alerts />}
        {activeMenu === 4 && <Mails />}
        {activeMenu === 5 && <Reports />}
        {activeMenu === 6 && <News />}
        {activeMenu === 7 && <Journals />}
      </div>
    </div>
  );
};

export default Activity;

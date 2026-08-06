import { useState } from "react";
// import { Tabs } from "antd";
import Symbols from "./Symbols";
import WatchList from "./WatchList";

const MarketWatch = () => {
  const [activeTab, setActiveTab] = useState(1);

  // const tabs = [
  //   { key: 1, label: "All Symbols" },
  //   { key: 2, label: "Watchlists" },
  // ];

  return (
    <div className="h-full w-full">
      {/* <Tabs
        items={tabs}
        onChange={(tab) => setActiveTab(tab)}
        className="w-full custom-tabs small-tabs"
        centered
      /> */}

      <div className="h-full">
        {activeTab === 1 && <Symbols />}
        {activeTab === 2 && <WatchList />}
      </div>
    </div>
  );
};

export default MarketWatch;

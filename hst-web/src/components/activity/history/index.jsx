import { useState } from "react";
import { Tabs } from "antd";
import ClosedPositions from "./ClosedPositions";
import Orders from "./Orders";
import Deals from "./Deals";

const History = () => {
  const [activeTab, setActiveTab] = useState(1);

  const tabs = [
    { key: 1, label: "Positions" },
    { key: 2, label: "Orders" },
    { key: 3, label: "Deals" },
  ];

  return (
    <div className="h-full flex flex-col">
      <Tabs
        items={tabs}
        onChange={(tab) => setActiveTab(tab)}
        className="small-tabs"
      />

      <div className="h-full">
        {activeTab === 1 && <ClosedPositions />}
        {activeTab === 2 && <Orders />}
        {activeTab === 3 && <Deals />}
      </div>
    </div>
  );
};

export default History;

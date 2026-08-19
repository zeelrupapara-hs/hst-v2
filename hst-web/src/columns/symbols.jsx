import Ask from "../components/marketWatch/components/Ask";
import Bid from "../components/marketWatch/components/Bid";
import LiveDot from "../components/marketWatch/components/LiveDot";
import NetChange from "../components/marketWatch/components/NetChange";
import PercentChange from "../components/marketWatch/components/PercentChange";
import PriceArrow from "../components/marketWatch/components/PriceArrow";
import Spread from "../components/marketWatch/components/Spread";
import OneClick from "../components/oneClick";

const columns = [
  {
    title: "Symbol",
    dataIndex: "symbol",
    key: "symbol",
    width: 100,
    fixed: "left",
    render: (value, record) => (
      <div className="flex items-center gap-1.5">
        <LiveDot symbolId={record?.id} />
        <PriceArrow symbolId={record?.id} />
        <span className="truncate">{value}</span>
      </div>
    ),
  },
  {
    title: "One Click Trade",
    dataIndex: "oneClick",
    key: "oneClick",
    width: 400,
    render: (_, record) => <OneClick symbolId={record?.id} />,
  },
  {
    title: "Bid",
    dataIndex: "bid",
    key: "bid",
    width: 100,
    render: (_, record) => <Bid symbolId={record?.id} />,
  },
  {
    title: "Ask",
    dataIndex: "ask",
    key: "ask",
    width: 100,
    render: (_, record) => <Ask symbolId={record?.id} />,
  },
  {
    title: "Spread",
    dataIndex: "spread",
    key: "spread",
    width: 100,
    render: (_, record) => <Spread symbolId={record?.id} />,
  },
  {
    title: "Change",
    dataIndex: "netChange",
    key: "netChange",
    width: 100,
    render: (_, record) => <NetChange symbolId={record?.id} />,
  },
  {
    title: "Change (%)",
    dataIndex: "percentChange",
    key: "percentChange",
    width: 100,
    render: (_, record) => <PercentChange symbolId={record?.id} />,
  },
  {
    title: "Open",
    dataIndex: "open",
    key: "open",
    width: 100,
  },
  {
    title: "Close",
    dataIndex: "close",
    key: "close",
    width: 100,
  },
  {
    title: "High",
    dataIndex: "high_bid",
    key: "high",
    width: 100,
  },
  {
    title: "Low",
    dataIndex: "low_bid",
    key: "low",
    width: 100,
  },
  {
    title: "Min Lot",
    dataIndex: "min_value",
    key: "minAmount",
    width: 100,
  },
  {
    title: "Max Lot",
    dataIndex: "max_value",
    key: "maxAmount",
    width: 100,
  },
  {
    title: "Contract Size",
    dataIndex: "contract_size",
    key: "contractSize",
    width: 100,
  },
];

export default columns;

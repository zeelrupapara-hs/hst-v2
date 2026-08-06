import { useEffect, useState } from "react";
import { Divider, Table } from "antd";
import { LuX } from "react-icons/lu";
import Icon from "../../common/Icon";
import ModalComponent from "../../modal/ModalComponent";
import { SYMBOL_DETAILS } from "../../../utils/constants";
import columns from "../../../columns/symbolInfo";
import Bid from "./Bid";
import Ask from "./Ask";

const SymbolInfoModal = ({ isOpen, setIsOpen, data }) => {
  const [sessions, setSessions] = useState([]);

  const days = [
    "sunday",
    "monday",
    "tuesday",
    "wednesday",
    "thursday",
    "friday",
    "saturday",
  ];

  useEffect(() => {
    if (data) {
      const quoteSessions = data?.quote_sessions?.split(",") || [];
      const tradeSessions = data?.trade_sessions?.split(",") || [];

      const quoteRow = { id: "quote", key: "quote", bgColor: "bg-red" };
      const tradeRow = { id: "trade", key: "trade", bgColor: "bg-green" };

      days.forEach((day, idx) => {
        quoteRow[day] = quoteSessions[idx] || "00:00-00:00";
        tradeRow[day] = tradeSessions[idx] || "00:00-00:00";
      });

      setSessions([quoteRow, tradeRow]);
    }
  }, [data]);

  const calculation = SYMBOL_DETAILS.CALCULATION[data?.calculation];
  const tradeLevel = SYMBOL_DETAILS.TRADE_LEVEL[data?.trade_level];
  const swapType = SYMBOL_DETAILS.SWAP_TYPE[data?.swap_type];
  const execution = SYMBOL_DETAILS.EXECUTION[data?.execution];

  const symbolDetails = [
    { label: "Symbol Group", value: data?.symbol_class?.desc },
    { label: "Calculation", value: calculation },
    { label: "Trade Level", value: tradeLevel },
    { label: "Spread", value: data?.spread },
    { label: "Stop Level", value: data?.stop_level },
    { label: "Contract Size", value: data?.contract_size },
    { label: "Min Lot", value: data?.min_value },
    { label: "Max Lot", value: data?.max_value },
    { label: "Digits", value: data?.digits },
    { label: "Step", value: data?.step },
    { label: "Initial Margin", value: data?.margin_initial },
    { label: "Maintenance Margin", value: data?.margin_maintenance },
    { label: "Initial Buy Margin", value: data?.margin_buy },
    { label: "Initial Sell Margin", value: data?.margin_sell },
    { label: "Maintenance Buy Margin", value: data?.maintenance_margin_buy },
    { label: "Maintenance Sell Margin", value: data?.maintenance_margin_sell },
    { label: "Execution", value: execution },
    { label: "Swap Type", value: swapType },
    { label: "Swap Long", value: data?.swap_long },
    { label: "Swap Short", value: data?.swap_short },
  ];

  const symbolInfo = [
    { label: "Bid", value: <Bid symbolId={data?.id} /> },
    { label: "Ask", value: <Ask symbolId={data?.id} /> },
    { label: "High", value: data?.high_bid },
    { label: "Low", value: data?.low_ask },
    { label: "Open", value: data?.open },
    { label: "Close", value: data?.close },
  ];

  return (
    <ModalComponent isOpen={isOpen} width={800}>
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>{data?.symbol}</span>

        <button onClick={() => setIsOpen(false)}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        <div className="grid grid-cols-2 gap-y-2 gap-x-10">
          {symbolDetails?.map((item, index) => (
            <div key={index} className="flex justify-between gap-2">
              <p className="font-semibold">{item?.label}</p>
              <p>{item?.value ?? "--"}</p>
            </div>
          ))}
        </div>

        <Divider className="!border-theme-border" />

        <div className="space-y-5">
          <div className="flex items-center justify-center gap-10">
            <div className="flex items-center gap-2">
              <div className="h-2 w-2 rounded-full bg-red"></div>
              <span>Quote Session</span>
            </div>

            <div className="flex items-center gap-2">
              <div className="h-2 w-2 rounded-full bg-green"></div>
              <span>Trade Session</span>
            </div>
          </div>

          <Table columns={columns} dataSource={sessions} pagination={false} />
        </div>

        <Divider className="!border-theme-border" />

        <div className="grid grid-cols-2 gap-y-2 gap-x-10">
          {symbolInfo?.map((item, index) => (
            <div key={index} className="flex justify-between gap-2">
              <p className="font-semibold">{item?.label}</p>
              <p>{item?.value ?? "--"}</p>
            </div>
          ))}
        </div>
      </div>
    </ModalComponent>
  );
};

export default SymbolInfoModal;

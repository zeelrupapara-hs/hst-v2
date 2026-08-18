import { useMemo } from "react";
import useSymbolStore from "../store/useSymbolStore";
import { defaultRender, formatPrice, getOrderType } from "../utils/utils";
import { LuArrowRight } from "react-icons/lu";
import Icon from "../components/common/Icon";
import Bid from "../components/marketWatch/components/Bid";
import Ask from "../components/marketWatch/components/Ask";
import NetProfitLoss from "../components/activity/netTrade/components/NetProfitLoss";
import NetValue from "../components/activity/netTrade/components/NetValue";
import { getSide } from "../utils/validation";
import Actions from "../components/activity/netTrade/components/Actions";

const useNetTrades = () => {
  const symbols = useSymbolStore((state) => state.symbols);

  const columns = [
    {
      title: "Symbol",
      dataIndex: "symbol_id",
      key: "symbol_id",
      render: (value) => symbols?.[value]?.symbol,
    },
    {
      title: "Type",
      dataIndex: "type",
      key: "type",
      render: (value, record) => (
        <span
          className={`${
            !value ? (record?.side === 0 ? "text-green" : "text-red") : ""
          } capitalize`}
        >
          {getOrderType(value, record?.side)}
        </span>
      ),
    },
    {
      title: "Lot",
      dataIndex: "volume",
      key: "volume",
      render: defaultRender,
    },
    {
      title: "Net Hedge",
      dataIndex: "netHedge",
      render: defaultRender,
    },
    {
      title: "Entry / Market",
      dataIndex: "open_price",
      key: "open_price",
      render: (value, record) => (
        <div className="flex items-center gap-2">
          <span>{formatPrice(value, symbols?.[record?.symbol_id]?.digits)}</span>

          <Icon Icon={LuArrowRight} size={14} />

          {getSide(record?.type, record?.side) === 0 ? (
            <Ask symbolId={record?.symbol_id} />
          ) : (
            <Bid symbolId={record?.symbol_id} />
          )}
        </div>
      ),
      width: 200,
    },
    {
      title: "P/L",
      dataIndex: "profit",
      key: "profit",
      render: (_, record) => <NetProfitLoss positions={record?.positions} />,
    },
    {
      title: "Net Value",
      dataIndex: "netValue",
      key: "netValue",
      render: (_, record) => <NetValue position={record} />,
      width: 120,
    },
    {
      title: "Net Open Value",
      dataIndex: "netOpenValue",
      key: "netOpenValue",
      render: defaultRender,
      width: 120,
    },
    {
      title: "Actions",
      dataIndex: "actions",
      key: "actions",
      render: (_, record) => <Actions record={record} />,
      width: 150,
    },
  ];

  return useMemo(() => columns, []);
};

export default useNetTrades;

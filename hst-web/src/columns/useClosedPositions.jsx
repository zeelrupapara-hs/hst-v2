import { useMemo } from "react";
import useSymbolStore from "../store/useSymbolStore";
import {
  defaultRender,
  formatDate,
  formateProfit,
  getOrderType,
} from "../utils/utils";
import { LuArrowRight } from "react-icons/lu";

const useClosedPositions = () => {
  const symbols = useSymbolStore((state) => state.symbols);

  const columns = [
    {
      title: "ID",
      dataIndex: "id",
      key: "id",
      render: defaultRender,
    },
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
      title: "Entry / Exit",
      dataIndex: "open_price",
      key: "open_price",
      render: (value, record) => (
        <span className="flex items-center gap-2">
          <span>{value}</span>
          <LuArrowRight size={13} />
          <span className="text-green">{record?.close_price}</span>
        </span>
      ),
      width: 200,
    },
    {
      title: "S/L",
      dataIndex: "stop_loss",
      key: "stop_loss",
      render: defaultRender,
    },
    {
      title: "T/P",
      dataIndex: "take_profit",
      key: "take_profit",
      render: defaultRender,
    },
    {
      title: "Open Time",
      dataIndex: "created_at",
      key: "created_at",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "Close Time",
      dataIndex: "updated_at",
      key: "updated_at",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "P/L",
      dataIndex: "profit",
      key: "profit",
      render: (value) => (
        <span className={`${value < 0 ? "text-red" : "text-green"}`}>
          {formateProfit(value)}
        </span>
      ),
    },
    {
      title: "Comment",
      dataIndex: "comment",
      key: "comment",
      render: defaultRender,
      ellipsis: true,
      width: 150,
    },
  ];

  return useMemo(() => columns, []);
};

export default useClosedPositions;

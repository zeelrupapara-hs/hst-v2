import { useMemo } from "react";
import useSymbolStore from "../store/useSymbolStore";
import { defaultRender, formatDate, getOrderType } from "../utils/utils";
import { ORDER_STATUS } from "../utils/constants";

const useOrders = () => {
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
      key: "size",
      render: defaultRender,
    },
    {
      title: "Filled Price",
      dataIndex: "filled_price",
      key: "filled_price",
      render: defaultRender,
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
      key: "open_time",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (value) => ORDER_STATUS[value],
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

export default useOrders;

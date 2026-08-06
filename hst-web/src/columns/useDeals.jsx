import { useMemo } from "react";
import useSymbolStore from "../store/useSymbolStore";
import {
  defaultRender,
  formatDate,
  formateProfit,
  getOrderType,
} from "../utils/utils";

const useDeals = () => {
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
      title: "Open Time",
      dataIndex: "created_at",
      key: "created_at",
      render: (value) => formatDate(value),
      width: 150,
    },
    {
      title: "P/L",
      dataIndex: "profit",
      key: "profit",
      render: (value) => formateProfit(value),
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

export default useDeals;

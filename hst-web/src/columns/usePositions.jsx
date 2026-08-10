import { useMemo } from "react";
import useSymbolStore from "../store/useSymbolStore";
import usePositionStore from "../store/usePositionStore";
import {
  defaultRender,
  formatDate,
  formatMoney,
  getOrderType,
} from "../utils/utils";
import { LuArrowRight } from "react-icons/lu";
import Icon from "../components/common/Icon";
import Bid from "../components/marketWatch/components/Bid";
import Ask from "../components/marketWatch/components/Ask";
import ProfitLoss from "../components/activity/positions/components/ProfitLoss";
import EditSLTP from "../components/activity/positions/components/EditSLTP";
import Actions from "../components/activity/positions/components/Actions";
import { getSide } from "../utils/validation";

const usePositions = () => {
  const symbols = useSymbolStore((state) => state.symbols);
  const currencyDigits = usePositionStore((state) => state.currencyDigits);

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
      title: "Entry / Market",
      dataIndex: "open_price",
      key: "open_price",
      render: (value, record) => (
        <div className="flex items-center gap-2">
          <span>{value || record?.order_limit_price}</span>

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
      title: "S/L",
      dataIndex: "stop_loss",
      key: "stop_loss",
      render: (value, record) => (
        <EditSLTP
          name={"stop_loss"}
          label="SL"
          value={value}
          record={record}
        />
      ),
    },
    {
      title: "T/P",
      dataIndex: "take_profit",
      key: "take_profit",
      render: (value, record) => (
        <EditSLTP
          name={"take_profit"}
          label="TP"
          value={value}
          record={record}
        />
      ),
    },
    {
      title: "Swap",
      dataIndex: "swaps",
      key: "swaps",
      render: (value) => formatMoney(value, currencyDigits),
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
      render: (_, record) => <ProfitLoss positionId={record?.id} />,
    },
    {
      title: "Comment",
      dataIndex: "comment",
      key: "comment",
      render: defaultRender,
      ellipsis: true,
      width: 150,
    },
    {
      title: "Actions",
      dataIndex: "actions",
      key: "actions",
      render: (_, record) => <Actions record={record} />,
      width: 150,
      fixed: "right",
    },
  ];

  // the render closures capture these, so a memo that never re-runs would keep showing the old
  // symbol list and the old currency digits
  return useMemo(() => columns, [symbols, currencyDigits]);
};

export default usePositions;

import { useState } from "react";
import { Checkbox } from "antd";
import { LuX } from "react-icons/lu";
import { useSocket } from "../../../socket";
import useSymbolStore from "../../../store/useSymbolStore";
import usePositionStore from "../../../store/usePositionStore";
import usePositions from "../../../columns/usePositions";
import SimpleTable from "../../table/SimpleTable";
import Summary from "./components/Summary";
import TotalPNL from "./components/TotalPNL";
import ContextMenuTable from "../../table/ContextMenuTable";
import { SOCKET_EVENTS } from "../../../socket/events";
import Icon from "../../common/Icon";
import HedgeModal from "./components/HedgeModal";

const Positions = () => {
  const { sendEvent } = useSocket();
  const symbols = useSymbolStore((state) => state.symbols);
  const positions = usePositionStore((state) => state.positions);
  const orders = usePositionStore((state) => state.orders);
  const getPositionsByType = usePositionStore(
    (state) => state.getPositionsByType
  );
  const [selectedRecord, setSelectedRecord] = useState(null);
  const [selectedPositions, setSelectedPositions] = useState([]);
  const [selectedOrders, setSelectedOrders] = useState([]);
  const [isHedgeModalOpen, setIsHedgeModalOpen] = useState(false);

  const symbolDetails = symbols?.[selectedRecord?.symbol_id];

  const summaryRow = { id: "summary", isSummary: true };
  const mergedData = [...positions, summaryRow, ...orders];
  const selectedRowKeys = [...selectedPositions, ...selectedOrders];

  const baseColumns = usePositions();
  const profitColIndex = baseColumns.findIndex(
    (col) => col.dataIndex === "profit"
  );
  const totalColSpan = profitColIndex;
  const profitColSpan = baseColumns?.length - totalColSpan;

  const columns = baseColumns.map((col, index) => ({
    ...col,
    render: (value, record, rowIndex) => {
      if (record?.isSummary) {
        if (index === 0) return <Summary />;
        if (index === profitColIndex) return <TotalPNL />;
        return null;
      }

      return col.render ? col.render(value, record, rowIndex) : value;
    },
    onCell: (record) => {
      if (record?.isSummary) {
        if (index === 0) return { colSpan: totalColSpan };
        if (index === profitColIndex) return { colSpan: profitColSpan };
        return { colSpan: 0 };
      }

      return {};
    },
  }));

  const handleInversePosition = () => {
    if (!selectedRecord) return;

    const payload = {
      symbol_id: selectedRecord?.symbol_id,
      volume: selectedRecord?.volume,
      side: selectedRecord?.side === 0 ? 1 : 0,
      order_price: 1,
    };

    sendEvent(SOCKET_EVENTS.ORDER_CREATE, payload);
    sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
      position_id: selectedRecord?.id,
      volume: selectedRecord?.volume,
    });
  };

  const handleDoublePosition = () => {
    if (!selectedRecord) return;

    const payload = {
      symbol_id: selectedRecord?.symbol_id,
      volume: selectedRecord?.volume,
      side: selectedRecord?.side,
      order_price: 1,
    };

    sendEvent(SOCKET_EVENTS.ORDER_CREATE, payload);
  };

  const handleCloseByHedge = () => {
    setIsHedgeModalOpen(true);
  };

  const handleCloseAllSymbol = () => {
    if (!selectedRecord) return;

    const symbolPositions = positions.filter(
      (position) => position?.symbol_id === selectedRecord?.symbol_id
    );

    symbolPositions.forEach((position) => {
      sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
        position_id: position?.id,
        volume: position?.volume,
      });
    });
  };

  const handleCloseAllProfitable = () => {
    const profitablePositions = getPositionsByType("profit");

    profitablePositions.forEach((position) => {
      sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
        position_id: position?.id,
        volume: position?.volume,
      });
    });
  };

  const handleCloseAllUnprofitable = () => {
    const unprofitablePositions = getPositionsByType("loss");

    unprofitablePositions.forEach((position) => {
      sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
        position_id: position?.id,
        volume: position?.volume,
      });
    });
  };

  const positionMenu = {
    items: [
      {
        key: 0,
        label: "Inverse Position",
        onClick: handleInversePosition,
      },
      {
        key: 1,
        label: "Double Position",
        onClick: handleDoublePosition,
      },
      {
        key: 2,
        label: "Close By Hedge",
        onClick: handleCloseByHedge,
      },
      {
        key: 3,
        label: `Close All ${symbolDetails?.symbol || ""}`,
        onClick: handleCloseAllSymbol,
      },
      {
        key: 4,
        label: "Close All Profitable",
        onClick: handleCloseAllProfitable,
      },
      {
        key: 5,
        label: "Close All Unprofitable",
        onClick: handleCloseAllUnprofitable,
      },
    ],
  };

  const handleAllSelection = (e, type) => {
    const checked = e?.target?.checked;

    if (type === "position") {
      if (checked) setSelectedPositions(positions.map((k) => k?.id));
      else setSelectedPositions([]);
    } else {
      if (checked) setSelectedOrders(orders.map((k) => k?.id));
      else setSelectedOrders([]);
    }
  };

  const handleRowSelection = (e, record) => {
    const checked = e?.target?.checked;
    const isPendingOrder = record?.type && record?.type !== 0;

    if (isPendingOrder) {
      if (checked) setSelectedOrders([...selectedOrders, record?.id]);
      else setSelectedOrders(selectedOrders.filter((k) => k !== record?.id));
    } else {
      if (checked) setSelectedPositions([...selectedPositions, record?.id]);
      else
        setSelectedPositions(selectedPositions.filter((k) => k !== record?.id));
    }
  };

  const handleClosePosition = () => {
    selectedPositions.forEach((positionId) => {
      const position = positions.find((k) => k?.id === positionId);
      sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
        position_id: position?.id,
        volume: position?.volume,
      });
    });
    setSelectedPositions([]);
  };

  const handleCloseOrder = () => {
    selectedOrders.forEach((orderId) => {
      sendEvent(SOCKET_EVENTS.ORDER_CANCEL, { order_id: orderId });
    });
    setSelectedOrders([]);
  };

  const rowSelection = {
    columnWidth: 60,
    selectedRowKeys,
    getCheckboxProps: (record) => ({
      disabled: record?.isSummary,
    }),
    columnTitle: () => (
      <div className="flex items-center justify-between gap-2">
        <Checkbox
          checked={
            selectedPositions?.length &&
            positions?.length === selectedPositions?.length
          }
          onChange={(e) => handleAllSelection(e, "position")}
        />

        {selectedPositions?.length ? (
          <button
            onClick={handleClosePosition}
            disabled={!selectedPositions?.length}
          >
            <Icon Icon={LuX} size={18} className="!text-red" />
          </button>
        ) : null}
      </div>
    ),
    renderCell: (checked, record, index, originNode) => {
      if (record?.isSummary) {
        return (
          <div className="flex items-center justify-between gap-2">
            <Checkbox
              checked={
                selectedOrders?.length &&
                orders?.length === selectedOrders?.length
              }
              onChange={(e) => handleAllSelection(e, "order")}
            />

            {selectedOrders?.length ? (
              <button
                onClick={handleCloseOrder}
                disabled={!selectedOrders?.length}
              >
                <Icon Icon={LuX} size={18} className="!text-red" />
              </button>
            ) : null}
          </div>
        );
      }

      return (
        <div className="flex items-center justify-between gap-2">
          <Checkbox
            checked={selectedRowKeys.includes(record?.id)}
            onChange={(e) => handleRowSelection(e, record)}
          />
        </div>
      );
    },
  };

  return (
    <>
      <ContextMenuTable
        menu={positionMenu}
        setSelectedRecord={setSelectedRecord}
        isContextMenuAllowed={(record) => !record?.type && !record?.isSummary}
      >
        <SimpleTable
          columns={columns}
          data={mergedData}
          rowClassName={(record) => (record?.isSummary ? "table-summary" : "")}
          rowSelection={rowSelection}
          virtual={false}
        />
      </ContextMenuTable>

      <HedgeModal
        isOpen={isHedgeModalOpen}
        setIsOpen={setIsHedgeModalOpen}
        record={selectedRecord}
      />
    </>
  );
};

export default Positions;

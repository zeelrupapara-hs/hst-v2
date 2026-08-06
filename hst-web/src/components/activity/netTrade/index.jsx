import { useState } from "react";
import usePositionStore from "../../../store/usePositionStore";
import useNetTrades from "../../../columns/useNetTrades";
import ContextMenuTable from "../../table/ContextMenuTable";
import SimpleTable from "../../table/SimpleTable";
import { calculateNetValue } from "../../../utils/calculations";
import { useSocket } from "../../../socket";
import { SOCKET_EVENTS } from "../../../socket/events";

const NetTrade = () => {
  const { sendEvent } = useSocket();
  const positions = usePositionStore((state) => state.positions);
  const [selectedRecord, setSelectedRecord] = useState(null);

  const columns = useNetTrades();

  const handleHedgePosition = () => {
    if (!selectedRecord?.netHedge) return;

    const payload = {
      symbol_id: selectedRecord?.symbol_id,
      volume: Math.abs(selectedRecord?.netHedge),
      side: selectedRecord?.netHedge > 0 ? 1 : 0,
      order_price: 1,
    };

    sendEvent(SOCKET_EVENTS.ORDER_CREATE, payload);
  };

  const netTradeMenu = {
    items: [
      {
        key: 0,
        label: "Hedge Position",
        onClick: handleHedgePosition,
      },
    ],
  };

  const netTrades = Object.values(
    positions.reduce((acc, pos) => {
      const { symbol_id, side, volume, profit, open_price } = pos;
      const key = `${symbol_id}_${side}`;

      if (!acc[key]) {
        acc[key] = {
          ...pos,
          id: key,
          volume: 0,
          profit: 0,
          weightedPriceSum: 0,
          positions: [],
        };
      }

      acc[key].volume += volume;
      acc[key].profit += profit;
      acc[key].weightedPriceSum += open_price * volume;
      acc[key].positions.push(pos);

      return acc;
    }, {})
  ).map((pos) => {
    const volume = Number(pos.volume.toFixed(2));
    const profit = Number(pos.profit.toFixed(2));

    const avgOpenPrice = pos.weightedPriceSum / pos.volume;
    const open_price = Number(avgOpenPrice.toFixed(pos.digits || 2));

    const netOpenValue = calculateNetValue({ ...pos, price: open_price });

    return { ...pos, volume, profit, open_price, netOpenValue };
  });

  const finalNetTrades = netTrades.map((pos) => {
    const { symbol_id, side } = pos;

    const buy = netTrades?.find((n) => n.id === `${symbol_id}_0`)?.volume || 0;
    const sell = netTrades?.find((n) => n.id === `${symbol_id}_1`)?.volume || 0;
    const diff = buy - sell;

    let netHedge = 0;
    if (diff > 0 && side === 0) netHedge = diff;
    if (diff < 0 && side === 1) netHedge = diff;

    return { ...pos, netHedge: Number(netHedge.toFixed(2)) };
  });

  return (
    <ContextMenuTable menu={netTradeMenu} setSelectedRecord={setSelectedRecord}>
      <SimpleTable columns={columns} data={finalNetTrades} />
    </ContextMenuTable>
  );
};

export default NetTrade;

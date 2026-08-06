import { LuCircleX } from "react-icons/lu";
import { useSocket } from "../../../../socket";
import EditPosition from "./EditPosition";
import PartialClose from "./PartialClose";
import { SOCKET_EVENTS } from "../../../../socket/events";
import Icon from "../../../common/Icon";

const Actions = ({ record }) => {
  const { sendEvent } = useSocket();

  const isPendingOrder = record?.type && record?.type !== 0;

  const handleClose = () => {
    if (isPendingOrder) closeOrder();
    else closePosition();
  };

  const closeOrder = () => {
    sendEvent(SOCKET_EVENTS.ORDER_CANCEL, { order_id: record?.id });
  };

  const closePosition = () => {
    sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
      position_id: record?.id,
      volume: record?.volume,
    });
  };

  return (
    <div className="flex items-center gap-2">
      <EditPosition record={record} />

      {!isPendingOrder && <PartialClose record={record} />}

      <button onClick={handleClose}>
        <Icon Icon={LuCircleX} size={18} />
      </button>
    </div>
  );
};

export default Actions;

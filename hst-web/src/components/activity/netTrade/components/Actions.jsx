import { LuCircleX } from "react-icons/lu";
import { useSocket } from "../../../../socket";
import { SOCKET_EVENTS } from "../../../../socket/events";
import Icon from "../../../common/Icon";

const Actions = ({ record }) => {
  const { sendEvent } = useSocket();

  const handleClose = () => {
    record?.positions?.forEach((pos) => {
      sendEvent(SOCKET_EVENTS.POSITION_CLOSE, {
        position_id: pos?.id,
        volume: pos?.volume,
      });
    });
  };

  return (
    <div className="flex items-center gap-2">
      <button onClick={handleClose}>
        <Icon Icon={LuCircleX} size={18} />
      </button>
    </div>
  );
};

export default Actions;

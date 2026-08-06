import { Rnd } from "react-rnd";
import { LuX } from "react-icons/lu";
import Portal from "../../common/Portal";
import useSymbolStore from "../../../store/useSymbolStore";
import useGlobalStore from "../../../store/useGlobalStore";
import Icon from "../../common/Icon";
import Order from "../index";

const OrderModal = ({ symbolId }) => {
  const symbolDeatils = useSymbolStore((state) => state.symbols?.[symbolId]);
  const closeOrderModal = useGlobalStore((state) => state.closeOrderModal);

  return (
    <Portal>
      <Rnd
        default={{
          x: 50,
          y: 50,
          width: 400,
        }}
        className="z-10"
        dragHandleClassName="modal-header"
        enableResizing={false}
      >
        <div className="h-full w-full bg-theme-bg border border-theme-border shadow-lg rounded-md">
          <div className="modal-header cursor-move flex items-center justify-between gap-2 p-2 bg-primary text-white rounded-t-md">
            <span>{symbolDeatils?.symbol}</span>

            <button onClick={() => closeOrderModal(symbolId)}>
              <Icon Icon={LuX} size={18} className="!text-white" />
            </button>
          </div>

          <Order symbolId={symbolId} showHeader={false} />
        </div>
      </Rnd>
    </Portal>
  );
};

export default OrderModal;

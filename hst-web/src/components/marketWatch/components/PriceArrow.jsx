import { LuArrowUp } from "react-icons/lu";
import { GoDotFill } from "react-icons/go";
import Icon from "../../common/Icon";
import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

const PriceArrow = ({ symbolId }) => {
  const bidColor = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]?.bidColor
  );

  return (
    <div>
      <Icon
        Icon={bidColor ? LuArrowUp : GoDotFill}
        size={14}
        className={`${
          bidColor === "red"
            ? "!text-red rotate-135"
            : bidColor === "green" && "!text-green rotate-45"
        } transition-all duration-100`}
      />
    </div>
  );
};

export default PriceArrow;

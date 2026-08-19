import { LuArrowUp } from "react-icons/lu";
import Icon from "../../common/Icon";
import useLiveSymbolStore from "../../../store/useLiveSymbolStore";

// The tick-direction arrow. Until a price has moved there is nothing to say — the status
// lamp beside it already covers the idle state.
const PriceArrow = ({ symbolId }) => {
  const bidColor = useLiveSymbolStore(
    (state) => state.liveSymbols?.[symbolId]?.bidColor
  );

  if (!bidColor) return null;

  return (
    <div>
      <Icon
        Icon={LuArrowUp}
        size={14}
        className={`${
          bidColor === "red" ? "!text-red rotate-135" : "!text-green rotate-45"
        } transition-all duration-100`}
      />
    </div>
  );
};

export default PriceArrow;

import { useState } from "react";
import { Form, Select } from "antd";
import { LuX } from "react-icons/lu";
import Icon from "../../../common/Icon";
import ModalComponent from "../../../modal/ModalComponent";
import useAuthStore from "../../../../store/useAuthStore";
import usePositionStore from "../../../../store/usePositionStore";
import Value from "../../../common/Value";
import { validateOrderField } from "../../../../utils/validation";
import Bid from "../../../marketWatch/components/Bid";
import Ask from "../../../marketWatch/components/Ask";
import { useSocket } from "../../../../socket";
import { SOCKET_EVENTS } from "../../../../socket/events";

const HedgeModal = ({ isOpen, setIsOpen, record }) => {
  const { sendEvent } = useSocket();
  const [form] = Form.useForm();
  const user = useAuthStore((state) => state.user);
  const positions = usePositionStore((state) => state.positions);
  const [hedgePosition, setHedgePosition] = useState(null);
  const [volume, setVolume] = useState(null);
  const [error, setError] = useState(false);

  const hedgePositions = positions.filter(
    (position) =>
      position?.symbol_id === record?.symbol_id &&
      position?.side !== record?.side
  );

  const selectedPosition = hedgePositions.find(
    (position) => position?.id === hedgePosition
  );

  const min = 0.01;
  const max = Math.min(record?.volume ?? 100, selectedPosition?.volume ?? 100);
  const range = `Min: ${min} | Max: ${max}`;

  const hasError = error || !volume || !hedgePosition;

  const updateValue = (key, value, range) => {
    const error = validateOrderField(key, value, range);
    setError(error);
    setVolume(value);
  };

  const getNewValue = (newValue) => {
    return volume === null ? min || max : newValue;
  };

  const clearAll = () => {
    form.resetFields();
    setHedgePosition(null);
    setVolume(null);
    setError(false);
  };

  const handleCloseByHedge = () => {
    if (hasError) return;

    sendEvent(SOCKET_EVENTS.POSITION_HEDGE, {
      hedge_position: hedgePosition,
      volume: volume,
      position: record?.id,
      account_id: user?.account_id,
    });

    clearAll();
    setIsOpen(false);
  };

  const handleClose = () => {
    clearAll();
    setIsOpen(false);
  };

  return (
    <ModalComponent isOpen={isOpen}>
      <div className="flex items-center justify-between gap-2 px-4 py-2 bg-primary text-white rounded-t-md">
        <span>Close By Hedge</span>

        <button onClick={handleClose}>
          <Icon Icon={LuX} size={18} className="!text-white" />
        </button>
      </div>

      <div className="p-4">
        <div className="grid grid-cols-2 gap-5 mb-5">
          <div className="flex flex-col items-center gap-2">
            <span>Bid</span>
            <Bid symbolId={record?.symbol_id} />
          </div>

          <div className="flex flex-col items-center gap-2">
            <span>Ask</span>
            <Ask symbolId={record?.symbol_id} />
          </div>
        </div>

        <Form
          form={form}
          layout="vertical"
          size="medium"
          className="custom-form"
        >
          <div className="grid grid-cols-2 gap-5">
            <Form.Item name="hedge_position" label="Hedge Position">
              <Select
                placeholder="Select Hedge Position"
                options={hedgePositions.map((pos) => ({
                  label: pos?.id,
                  value: pos?.id,
                }))}
                onChange={(value) => setHedgePosition(value)}
              />
            </Form.Item>

            <Form.Item name="volume">
              <div>
                <div className="flex items-center justify-between gap-2 pb-1 text-xs leading-[22px]">
                  <span>Lot</span>
                  <span>{range}</span>
                </div>

                <Value
                  min={min}
                  max={max}
                  value={volume}
                  setValue={(val, inputType) => {
                    const newValue =
                      inputType === "manual" ? val : getNewValue(val);
                    updateValue("volume", newValue, { min, max });
                  }}
                  className="h-[32px]"
                />
              </div>
            </Form.Item>
          </div>

          <button
            className="btn-primary w-full"
            onClick={handleCloseByHedge}
            disabled={hasError}
          >
            Close By Hedge
          </button>
        </Form>
      </div>
    </ModalComponent>
  );
};

export default HedgeModal;

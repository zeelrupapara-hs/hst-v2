import { useRef } from "react";
import { MdFullscreen } from "react-icons/md";
import Icon from "./Icon";

const FullScreenToggle = () => {
  const elementRef = useRef(document.documentElement);

  const toggleFullScreen = () => {
    if (!document.fullscreenElement) elementRef.current.requestFullscreen();
    else document.exitFullscreen();
  };

  return (
    <button
      title="Toggle Fullscreen"
      className="p-2"
      onClick={toggleFullScreen}
    >
      <Icon Icon={MdFullscreen} size={20} />
    </button>
  );
};

export default FullScreenToggle;

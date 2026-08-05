import { useFullscreen } from "../../hooks/useFullscreen.js";

export function FullscreenButton() {
  const { fullscreen, toggle } = useFullscreen();

  return (
    <button
      type="button"
      className="title-btn"
      title={fullscreen ? "Exit Full Screen (F11)" : "Full Screen (F11)"}
      aria-label={fullscreen ? "Exit Full Screen" : "Full Screen"}
      onClick={toggle}
    >
      {fullscreen ? "❐" : "□"}
    </button>
  );
}

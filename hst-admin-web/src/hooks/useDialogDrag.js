import { useCallback, useEffect, useRef, useState } from "react";

/** Drag settings dialogs by the title bar; offset resets when resetKey changes. */
export function useDialogDrag(resetKey) {
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const dragRef = useRef({ active: false, startX: 0, startY: 0, origX: 0, origY: 0 });

  useEffect(() => {
    setOffset({ x: 0, y: 0 });
  }, [resetKey]);

  const onTitlePointerDown = useCallback(
    (e) => {
      if (e.button !== 0) return;
      e.preventDefault();
      dragRef.current = {
        active: true,
        startX: e.clientX,
        startY: e.clientY,
        origX: offset.x,
        origY: offset.y,
      };
    },
    [offset.x, offset.y],
  );

  useEffect(() => {
    function onMove(e) {
      if (!dragRef.current.active) return;
      setOffset({
        x: dragRef.current.origX + (e.clientX - dragRef.current.startX),
        y: dragRef.current.origY + (e.clientY - dragRef.current.startY),
      });
    }
    const onUp = () => (dragRef.current.active = false);
    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);
    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
    };
  }, []);

  return { offset, onTitlePointerDown };
}

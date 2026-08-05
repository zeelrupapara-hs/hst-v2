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
      const title = e.currentTarget.getBoundingClientRect();
      dragRef.current = {
        active: true,
        startX: e.clientX,
        startY: e.clientY,
        origX: offset.x,
        origY: offset.y,
        title,
      };
    },
    [offset.x, offset.y],
  );

  useEffect(() => {
    // the title bar must stay reachable: never above the top, never fully off any edge
    function onMove(e) {
      const d = dragRef.current;
      if (!d.active) return;
      let x = d.origX + (e.clientX - d.startX);
      let y = d.origY + (e.clientY - d.startY);
      const dx = x - d.origX;
      const dy = y - d.origY;
      if (d.title.top + dy < 0) y = d.origY - d.title.top;
      if (d.title.top + dy > window.innerHeight - 40) y = d.origY + (window.innerHeight - 40 - d.title.top);
      if (d.title.right + dx < 80) x = d.origX + (80 - d.title.right);
      if (d.title.left + dx > window.innerWidth - 80) x = d.origX + (window.innerWidth - 80 - d.title.left);
      setOffset({ x, y });
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

import { useCallback, useEffect, useState } from "react";

export function useFullscreen() {
  const [fullscreen, setFullscreen] = useState(false);

  useEffect(() => {
    function onChange() {
      setFullscreen(!!(document.fullscreenElement || document.webkitFullscreenElement));
    }
    document.addEventListener("fullscreenchange", onChange);
    document.addEventListener("webkitfullscreenchange", onChange);
    return () => {
      document.removeEventListener("fullscreenchange", onChange);
      document.removeEventListener("webkitfullscreenchange", onChange);
    };
  }, []);

  const toggle = useCallback(() => {
    const el = document.documentElement;
    const on = !!(document.fullscreenElement || document.webkitFullscreenElement);
    if (on) {
      (document.exitFullscreen || document.webkitExitFullscreen)?.call(document);
    } else {
      (el.requestFullscreen || el.webkitRequestFullscreen)?.call(el);
    }
  }, []);

  useEffect(() => {
    document.body.classList.toggle("is-fullscreen", fullscreen);
  }, [fullscreen]);

  useEffect(() => {
    function onKey(e) {
      if (e.key === "F11") {
        e.preventDefault();
        toggle();
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [toggle]);

  return { fullscreen, toggle };
}

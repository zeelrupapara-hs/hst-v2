export const applyChartInactiveState = (container, inactive, { darkMode = false } = {}) => {
  if (!container) return false;

  const iframe = container.querySelector("iframe");
  const doc = iframe?.contentDocument;
  if (!doc) return false;

  const top = doc.querySelector(".layout__area--top");
  const drawing = doc.querySelector(".drawing-toolbar");
  const center = doc.querySelector(".layout__area--center");

  [top, drawing].forEach((el) => {
    if (!el) return;
    el.style.pointerEvents = inactive ? "none" : "";
    el.style.opacity = inactive ? "0.55" : "";
  });

  let overlay = doc.querySelector(".chart-no-data-overlay");

  if (inactive && center) {
    if (!overlay) {
      overlay = doc.createElement("div");
      overlay.className = "chart-no-data-overlay";
      overlay.textContent = "No Data Here";
      Object.assign(overlay.style, {
        position: "absolute",
        inset: "0",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: darkMode ? "#101013" : "#ffffff",
        color: darkMode ? "rgba(137, 137, 139, 0.9)" : "rgba(21, 25, 36, 0.7)",
        fontSize: "14px",
        fontWeight: "500",
        zIndex: "10",
        pointerEvents: "auto",
      });

      if (getComputedStyle(center).position === "static") {
        center.style.position = "relative";
      }

      center.appendChild(overlay);
    }
  } else if (overlay) {
    overlay.remove();
  }

  return Boolean(top || drawing || center);
};

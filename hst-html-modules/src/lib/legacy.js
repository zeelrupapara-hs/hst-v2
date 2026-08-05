/** Load legacy IIFE scripts before the React app mounts. */
const LEGACY_SCRIPTS = ["/assets/js/mock-data.js", "/assets/js/trade-ui.js"];

export function loadLegacyScripts() {
  return Promise.all(
    LEGACY_SCRIPTS.map(
      (src) =>
        new Promise((resolve, reject) => {
          if (document.querySelector(`script[src="${src}"]`)) {
            resolve();
            return;
          }
          const script = document.createElement("script");
          script.src = src;
          script.onload = resolve;
          script.onerror = () => reject(new Error("Failed to load " + src));
          document.head.appendChild(script);
        })
    )
  );
}

export function getMock() {
  return window.HSTMock;
}

export function getTradeUi() {
  return window.HSTTradeUI;
}

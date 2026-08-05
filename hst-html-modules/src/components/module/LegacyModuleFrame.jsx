export function LegacyModuleFrame({ src, title }) {
  return (
    <iframe
      className="legacy-module-frame"
      src={src}
      title={title || "Module"}
      style={{ width: "100%", height: "100%", border: "none", display: "block" }}
    />
  );
}

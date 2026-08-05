/**
 * MT5 Administrator config dialog shell — fixed width/height.
 * @param {"inline"|"overlay"} mode inline = centered host; overlay = dialog only (inside overlay wrapper)
 */
export function Mt5ConfigDialog({
  title,
  tabs,
  children,
  footer,
  className = "",
  width = 656,
  mode = "inline",
  draggable = false,
  onTitlePointerDown,
  titleClassName = "",
}) {
  const dialog = (
    <div
      className={`config-window mt5-config-dialog ${className}`.trim()}
      style={{ width: `${width}px` }}
      role="dialog"
      aria-modal={mode === "overlay" ? "true" : undefined}
      aria-label={title}
    >
      <div
        className={`config-title${draggable ? " config-title-draggable" : ""} ${titleClassName}`.trim()}
        onMouseDown={draggable ? onTitlePointerDown : undefined}
      >
        {title}
      </div>
      {tabs}
      <div className="config-body">{children}</div>
      {footer}
    </div>
  );

  if (mode === "overlay") {
    return dialog;
  }

  return <div className="mt5-dialog-host">{dialog}</div>;
}

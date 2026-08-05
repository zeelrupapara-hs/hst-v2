/**
 * Settings dialog shell — fixed 656px, tab strip under the title, OK/Cancel footer.
 * @param {{title: string, tabs?: any, children: any, footer?: any, className?: string, width?: number, draggable?: boolean, onTitlePointerDown?: Function}} props
 */
export function SettingsDialog({
  title,
  tabs,
  children,
  footer,
  className = "",
  width = 656,
  draggable = false,
  onTitlePointerDown,
}) {
  return (
    <div
      className={`config-window settings-dialog ${className}`.trim()}
      style={{ width: `${width}px` }}
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <div
        className={`config-title${draggable ? " config-title-draggable" : ""}`}
        onMouseDown={draggable ? onTitlePointerDown : undefined}
      >
        {title}
      </div>
      {tabs}
      <div className="config-body">{children}</div>
      {footer}
    </div>
  );
}

/**
 * Settings dialog shell: title bar with the help and close glyphs, tab strip, body, footer.
 * Width and height are fixed per dialog: switching tabs must never resize the window.
 * @param {{title: string, tabs?: any, children: any, footer?: any, className?: string, width?: number, draggable?: boolean, onTitlePointerDown?: Function, onClose?: Function}} props
 */
export function SettingsDialog({
  title,
  tabs,
  children,
  footer,
  className = "",
  width = 613,
  height = 441,
  draggable = false,
  onTitlePointerDown,
  onClose,
}) {
  return (
    <div
      className={`config-window settings-dialog ${className}`.trim()}
      style={{ width: `${width}px`, height: `${height}px` }}
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <div
        className={`config-title${draggable ? " config-title-draggable" : ""}`}
        onMouseDown={draggable ? onTitlePointerDown : undefined}
      >
        <span className="config-title-text">{title}</span>
        <span className="config-title-btns">
          <button type="button" className="config-title-btn" title="Help" disabled>?</button>
          <button type="button" className="config-title-btn" title="Close" onClick={onClose}>
            ✕
          </button>
        </span>
      </div>
      {tabs}
      <div className="config-body">{children}</div>
      {footer}
    </div>
  );
}

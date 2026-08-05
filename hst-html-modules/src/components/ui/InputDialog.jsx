import { useEffect, useRef } from "react";
import { createPortal } from "react-dom";

/** Small text input dialog (replaces window.prompt). */
export function InputDialog({
  title,
  label,
  value,
  onChange,
  onOk,
  onCancel,
  error,
  okLabel = "OK",
  cancelLabel = "Cancel",
}) {
  const inputRef = useRef(null);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, []);

  function onKeyDown(e) {
    if (e.key === "Enter") {
      e.preventDefault();
      onOk?.();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel?.();
    }
  }

  return createPortal(
    <div
      className="dialog-overlay input-dialog-overlay"
      onClick={onCancel}
      role="presentation"
    >
      <div
        className="input-dialog"
        role="dialog"
        aria-modal="true"
        aria-label={title}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="config-title">{title}</div>
        <div className="input-dialog-body">
          <label htmlFor="input-dialog-field">{label}</label>
          <input
            id="input-dialog-field"
            ref={inputRef}
            type="text"
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            onKeyDown={onKeyDown}
          />
          {error ? <p className="input-dialog-error">{error}</p> : null}
        </div>
        <div className="config-actions">
          <button type="button" className="config-ok" onClick={onOk}>
            {okLabel}
          </button>
          <button type="button" onClick={onCancel}>
            {cancelLabel}
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

import { useEffect, useRef } from "react";
import { createPortal } from "react-dom";

/** Small MT5-style text input dialog (replaces window.prompt). */
export function Mt5InputDialog({
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
      className="mt5-dialog-overlay mt5-input-overlay"
      onClick={onCancel}
      role="presentation"
    >
      <div
        className="mt5-input-dialog"
        role="dialog"
        aria-modal="true"
        aria-label={title}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="config-title">{title}</div>
        <div className="mt5-input-body">
          <label htmlFor="mt5-input-field">{label}</label>
          <input
            id="mt5-input-field"
            ref={inputRef}
            type="text"
            value={value}
            onChange={(e) => onChange?.(e.target.value)}
            onKeyDown={onKeyDown}
          />
          {error ? <p className="mt5-input-error">{error}</p> : null}
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

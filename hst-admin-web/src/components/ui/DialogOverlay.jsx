import { createPortal } from "react-dom";

/**
 * Modal backdrop; does not dismiss on outside click. It renders at the document root so a
 * dialog opened from inside another dialog is not clipped by its parent's scroll box.
 */
export function DialogOverlay({ className = "dialog-overlay", children }) {
  return createPortal(
    <div className={className} role="presentation">
      {children}
    </div>,
    document.body,
  );
}

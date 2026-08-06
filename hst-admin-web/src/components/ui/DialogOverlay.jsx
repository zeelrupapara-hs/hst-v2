/** Modal backdrop; does not dismiss on outside click. */
export function DialogOverlay({ className = "dialog-overlay", children }) {
  return (
    <div className={className} role="presentation">
      {children}
    </div>
  );
}

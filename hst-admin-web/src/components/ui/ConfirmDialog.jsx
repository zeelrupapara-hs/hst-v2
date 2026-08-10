import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";

/** Simple Yes/No confirmation, styled like the administrator reference dialog. */
export function ConfirmDialog({
  title = "HST Administrator",
  message,
  prompt = "Are you sure you want to continue?",
  yesLabel = "Yes",
  noLabel = "No",
  onYes,
  onClose,
  showYes = true,
  variant = "question",
  overlayClassName = "dialog-overlay",
}) {
  const close = useDialogStack(onClose);
  const iconClass =
    variant === "warning" ? "admin-confirm-icon admin-confirm-icon-warning" : "admin-confirm-icon";

  return (
    <DialogOverlay className={overlayClassName}>
      <div className="config-window sym-confirm admin-confirm">
        <div className="config-title">
          <span className="config-title-text">{title}</span>
        </div>
        <div className="config-body admin-confirm-body">
          <span className={iconClass} aria-hidden="true">
            {variant === "warning" ? "!" : "?"}
          </span>
          <div className="admin-confirm-text">
            <p>{message}</p>
            {showYes && prompt ? <p>{prompt}</p> : null}
          </div>
        </div>
        <div className="config-actions admin-confirm-actions">
          {showYes ? (
            <>
              <button
                type="button"
                className="config-ok"
                onClick={() => {
                  onYes?.();
                  close();
                }}
              >
                {yesLabel}
              </button>
              <button type="button" onClick={close}>
                {noLabel}
              </button>
            </>
          ) : (
            <button type="button" className="config-ok" onClick={close}>
              OK
            </button>
          )}
        </div>
      </div>
    </DialogOverlay>
  );
}

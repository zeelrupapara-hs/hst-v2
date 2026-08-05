export function ModuleToolbar({ onReload, onOpen, editLabel, canOpen }) {
  return (
    <div className="module-toolbar">
      <button type="button" className="tb-btn" onClick={onReload}>
        ↻ Reload
      </button>
      {onOpen && (
        <button
          type="button"
          className="tb-btn"
          disabled={!canOpen}
          onClick={onOpen}
        >
          {editLabel || "✎ Open…"}
        </button>
      )}
    </div>
  );
}

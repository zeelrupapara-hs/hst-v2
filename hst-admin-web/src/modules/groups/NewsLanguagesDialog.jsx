import { useMemo, useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { NEWS_LANGUAGES, newsLangLabel } from "@/constants/newsLanguages.js";

/** MT5-style dual-list picker for group news_langs (Windows LANGID values). */
export function NewsLanguagesDialog({ selected = [], onClose, onChange }) {
  const [picked, setPicked] = useState(() => [...selected]);
  const [availSel, setAvailSel] = useState(null);
  const [chosenSel, setChosenSel] = useState(null);
  const { offset, onTitlePointerDown } = useDialogDrag("news-langs");
  const close = useDialogStack(onClose);

  const catalog = useMemo(() => {
    const map = new Map(NEWS_LANGUAGES.map((l) => [l.id, l.label]));
    for (const id of picked) {
      if (!map.has(id)) map.set(id, newsLangLabel(id));
    }
    return [...map.entries()]
      .map(([id, label]) => ({ id, label }))
      .sort((a, b) => a.label.localeCompare(b.label));
  }, [picked]);

  const available = catalog.filter((l) => !picked.includes(l.id));
  const chosen = picked
    .map((id) => catalog.find((l) => l.id === id) ?? { id, label: newsLangLabel(id) })
    .sort((a, b) => a.label.localeCompare(b.label));

  function insert() {
    if (availSel == null || picked.includes(availSel)) return;
    setPicked((prev) => [...prev, availSel]);
    setAvailSel(null);
  }

  function remove() {
    if (chosenSel == null) return;
    setPicked((prev) => prev.filter((id) => id !== chosenSel));
    setChosenSel(null);
  }

  function reset() {
    setPicked([]);
    setAvailSel(null);
    setChosenSel(null);
  }

  function handleClose() {
    onChange(picked);
    close();
  }

  return (
    <DialogOverlay>
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
      >
        <SettingsDialog
          draggable
          width={520}
          height={420}
          onClose={handleClose}
          onTitlePointerDown={onTitlePointerDown}
          title="Select news languages"
          footer={
            <div className="config-actions">
              <button type="button" className="config-ok" onClick={handleClose}>
                Close
              </button>
            </div>
          }
        >
          <div className="config-panel active dual-list-dialog">
            <div className="dual-list-layout">
              <div className="dual-list-pane">
                <span className="dual-list-caption">Available languages:</span>
                <ul className="dual-list-box" role="listbox">
                  {available.map((row) => (
                    <li
                      key={row.id}
                      role="option"
                      aria-selected={availSel === row.id}
                      className={availSel === row.id ? "selected" : ""}
                      onClick={() => setAvailSel(row.id)}
                      onDoubleClick={() => {
                        setPicked((prev) => (prev.includes(row.id) ? prev : [...prev, row.id]));
                        setAvailSel(null);
                      }}
                    >
                      {row.label}
                    </li>
                  ))}
                </ul>
              </div>

              <div className="dual-list-actions">
                <button type="button" disabled={availSel == null} onClick={insert}>
                  Insert -&gt;
                </button>
                <button type="button" disabled={chosenSel == null} onClick={remove}>
                  &lt;- Remove
                </button>
                <button type="button" disabled={!picked.length} onClick={reset}>
                  Reset
                </button>
              </div>

              <div className="dual-list-pane">
                <span className="dual-list-caption">Selected languages:</span>
                <ul className="dual-list-box" role="listbox">
                  {chosen.map((row) => (
                    <li
                      key={row.id}
                      role="option"
                      aria-selected={chosenSel === row.id}
                      className={chosenSel === row.id ? "selected" : ""}
                      onClick={() => setChosenSel(row.id)}
                      onDoubleClick={() => {
                        setPicked((prev) => prev.filter((id) => id !== row.id));
                        setChosenSel(null);
                      }}
                    >
                      {row.label}
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </SettingsDialog>
      </div>
    </DialogOverlay>
  );
}

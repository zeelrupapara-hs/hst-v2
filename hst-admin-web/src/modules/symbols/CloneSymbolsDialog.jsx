import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { useDialogDrag } from "@/hooks/useDialogDrag.js";
import { cloneSymbols } from "@/api/endpoints/symbols.js";
import { TabIntro } from "./SymbolTabs.jsx";

const suffix = (postfix) => (postfix ? `.${postfix}` : "");

function Pane({ title, folder, names }) {
  return (
    <div className="sym-clone-pane">
      <div className="sym-clone-pane-title">
        <span className="sym-clone-folder-icon" aria-hidden="true" />
        {title}
      </div>
      <div className="sym-clone-tree">
        <div className="sym-clone-folder">
          <span className="sym-clone-folder-icon" aria-hidden="true" />
          {folder || "\\"}
        </div>
        {names.map((n) => (
          <div key={n} className="sym-clone-leaf">
            <span className="sym-coin-icon" aria-hidden="true" />
            {n}
          </div>
        ))}
      </div>
    </div>
  );
}

/**
 * Clone Symbols dialog: one postfix applied to the selection, or to a whole folder.
 * @param {{symbols: string[], folder?: string, onClose: Function, onDone: Function}} props
 */
export function CloneSymbolsDialog({ symbols, folder = "", onClose, onDone }) {
  const [postfix, setPostfix] = useState("");
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag("clone-symbols");
  const target = `${folder}${suffix(postfix)}`;

  async function handleOk() {
    if (!postfix.trim()) {
      setError("A postfix is required");
      return;
    }
    // the selection wins; an empty one clones the folder the list is showing
    // the field is typed without the leading dot, so the wire carries the dotted form
    const dotted = suffix(postfix.trim());
    const res = await cloneSymbols(
      symbols.length
        ? { postfix: dotted, symbols: symbols.map((r) => r.symbol_id) }
        : { postfix: dotted, path: folder },
    );
    if (!res.ok) {
      setError(res.message || "clone failed");
      return;
    }
    onDone();
    onClose();
  }

  return (
    <div className="dialog-overlay" onClick={onClose} role="presentation">
      <div
        className="dialog-positioner"
        style={{ transform: `translate(${offset.x}px, ${offset.y}px)` }}
        onClick={(e) => e.stopPropagation()}
      >
        <SettingsDialog
          draggable
          onClose={onClose}
          onTitlePointerDown={onTitlePointerDown}
          className="sym-clone-window"
          width={651}
          height={420}
          title="Clone Symbols"
          footer={
            <div className="config-actions">
              {error && <span className="login-error">{error}</span>}
              <button type="button" className="config-ok" onClick={handleOk}>
                OK
              </button>
              <button type="button" onClick={onClose}>
                Cancel
              </button>
            </div>
          }
        >
          <TabIntro>Please specify new postfix to clone selected symbols.</TabIntro>
          <div className="form-grid sym-clone-form">
            <label>New postfix</label>
            <span className="sym-with-suffix">
              <input type="text" value={postfix} onChange={(e) => setPostfix(e.target.value)} />
              <span className="sym-suffix">without leading dot</span>
            </span>
            <label>Copy to</label>
            <input type="text" readOnly value={`${target}\\`} />
          </div>
          <p className="sym-clone-sentence">
            {symbols.length || "All"} symbols will be copied from &apos;{folder}\&apos; to &apos;
            {target}\&apos; with postfix {suffix(postfix)}
          </p>
          <div className="sym-clone-panes">
            <Pane title="Symbols - From" folder={folder} names={symbols.map((r) => r.symbol)} />
            <Pane
              title="Symbols - To"
              folder={target}
              names={symbols.map((r) => `${r.symbol}${suffix(postfix)}`)}
            />
          </div>
        </SettingsDialog>
      </div>
    </div>
  );
}

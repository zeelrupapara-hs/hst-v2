import { useState } from "react";
import { SettingsDialog } from "@/components/ui/SettingsDialog.jsx";
import { DialogOverlay } from "@/components/ui/DialogOverlay.jsx";
import { useDialogStack } from "@/hooks/useDialogStack.jsx";
import { SymbolFolderSelect } from "@/components/ui/SymbolFolderSelect.jsx";
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
          {folder || "Symbols"}
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
  const [copyTo, setCopyTo] = useState(folder);
  const [error, setError] = useState("");
  const { offset, onTitlePointerDown } = useDialogDrag("clone-symbols");
  const close = useDialogStack(onClose);
  const target = `${copyTo}${suffix(postfix)}`;

  async function handleOk() {
    if (!postfix.trim()) {
      setError("A postfix is required");
      return;
    }
    // the selection wins; an empty one clones the folder the list is showing
    // the field is typed without the leading dot, so the wire carries the dotted form
    const dotted = suffix(postfix.trim());
    const payload = { postfix: dotted, copy_to: copyTo };
    const res = await cloneSymbols(
      symbols.length
        ? { ...payload, symbols: symbols.map((r) => r.symbol_id) }
        : { ...payload, path: folder },
    );
    if (!res.ok) {
      setError(res.message || "clone failed");
      return;
    }
    onDone();
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
          onClose={close}
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
              <button type="button" onClick={close}>
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
            <SymbolFolderSelect value={copyTo} onChange={setCopyTo} />
          </div>
          <p className="sym-clone-sentence">
            {symbols.length || "All"} symbols will be copied from &apos;{folder || "Symbols"}\&apos; to &apos;
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
    </DialogOverlay>
  );
}

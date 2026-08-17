import { useEffect, useRef, useState } from "react";
import { PropSelect } from "@/components/ui/PropSelect.jsx";

const FONT_SIZES = [
  { value: 1, label: "8" },
  { value: 2, label: "10" },
  { value: 3, label: "12" },
  { value: 4, label: "14" },
  { value: 5, label: "18" },
  { value: 6, label: "24" },
  { value: 7, label: "36" },
];

/**
 * Compose editor: a contentEditable area with the mail toolbar. The value is an HTML string;
 * the server sanitizes on send, so this component only has to produce it, not police it.
 * @param {{value: string, onChange: (html: string) => void,
 *   extraTools?: (exec: (cmd: string, arg?: string) => void) => object}} props
 */
export function RichTextEdit({ value, onChange, extraTools }) {
  const areaRef = useRef(null);
  const savedRange = useRef(null);
  const [fontSize, setFontSize] = useState(2);
  const [urlMode, setUrlMode] = useState(null);
  const [url, setUrl] = useState("");

  // written back only when the outside value is not what the area already holds (template load)
  useEffect(() => {
    const el = areaRef.current;
    if (el && el.innerHTML !== (value || "")) el.innerHTML = value || "";
  }, [value]);

  const saveSelection = () => {
    const sel = window.getSelection();
    if (sel?.rangeCount && areaRef.current?.contains(sel.anchorNode)) {
      savedRange.current = sel.getRangeAt(0);
    }
  };

  const restoreSelection = () => {
    const range = savedRange.current;
    if (!range) return;
    const sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
  };

  const exec = (cmd, arg) => {
    areaRef.current?.focus();
    restoreSelection();
    document.execCommand(cmd, false, arg);
    saveSelection();
    onChange?.(areaRef.current?.innerHTML ?? "");
  };

  const applyUrl = () => {
    const trimmed = url.trim();
    if (trimmed) exec(urlMode === "image" ? "insertImage" : "createLink", trimmed);
    setUrlMode(null);
    setUrl("");
  };

  // mousedown is swallowed so the button never steals the selection the command applies to
  const tool = (label, title, cmd, arg, className = "") => (
    <button
      type="button"
      className={`rich-text-btn ${className}`}
      title={title}
      onMouseDown={(e) => e.preventDefault()}
      onClick={() => exec(cmd, arg)}
    >
      {label}
    </button>
  );

  return (
    <div className="rich-text">
      <div className="rich-text-toolbar" onMouseDown={saveSelection}>
        {tool("B", "Bold", "bold", undefined, "rich-text-b")}
        {tool("I", "Italic", "italic", undefined, "rich-text-i")}
        {tool("U", "Underline", "underline", undefined, "rich-text-u")}
        <PropSelect
          className="rich-text-size"
          value={fontSize}
          options={FONT_SIZES}
          onChange={(v) => {
            setFontSize(v);
            exec("fontSize", String(v));
          }}
        />
        <span className="rich-text-sep" />
        {tool("•≡", "Bulleted list", "insertUnorderedList")}
        {tool("1≡", "Numbered list", "insertOrderedList")}
        {tool("⇤", "Decrease indent", "outdent")}
        {tool("⇥", "Increase indent", "indent")}
        <span className="rich-text-sep" />
        {tool("⟝", "Align left", "justifyLeft")}
        {tool("⟠", "Align center", "justifyCenter")}
        {tool("⟞", "Align right", "justifyRight")}
        <span className="rich-text-sep" />
        <label className="rich-text-btn rich-text-color" title="Text color">
          A
          <input
            type="color"
            onMouseDown={saveSelection}
            onInput={(e) => exec("foreColor", e.target.value)}
          />
        </label>
        <button
          type="button"
          className="rich-text-btn"
          title="Insert link"
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => setUrlMode(urlMode === "link" ? null : "link")}
        >
          🔗
        </button>
        <button
          type="button"
          className="rich-text-btn"
          title="Insert image by URL"
          onMouseDown={(e) => e.preventDefault()}
          onClick={() => setUrlMode(urlMode === "image" ? null : "image")}
        >
          🖼
        </button>
        {extraTools?.(exec)}
      </div>
      {urlMode && (
        <div className="rich-text-urlbar">
          <span>{urlMode === "image" ? "Image URL" : "Link URL"}</span>
          <input
            type="text"
            value={url}
            autoFocus
            placeholder="https://"
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && applyUrl()}
          />
          <button type="button" onClick={applyUrl}>OK</button>
          <button type="button" onClick={() => { setUrlMode(null); setUrl(""); }}>Cancel</button>
        </div>
      )}
      <div
        ref={areaRef}
        className="rich-text-area"
        contentEditable
        suppressContentEditableWarning
        onInput={() => {
          saveSelection();
          onChange?.(areaRef.current?.innerHTML ?? "");
        }}
        onKeyUp={saveSelection}
        onMouseUp={saveSelection}
      />
    </div>
  );
}

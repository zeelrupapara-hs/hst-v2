import { CKEditor } from "@ckeditor/ckeditor5-react";
import ClassicEditor from "@ckeditor/ckeditor5-build-classic";

const Editor = ({ value = "", onChange, disabled = false }) => {
  return (
    <div className="custom-editor prose dark:prose-invert max-w-full">
      <CKEditor
        editor={ClassicEditor}
        data={value}
        onChange={(e, editor) => {
          const data = editor.getData();
          onChange?.(data);
        }}
        disabled={disabled}
      />
    </div>
  );
};

export default Editor;

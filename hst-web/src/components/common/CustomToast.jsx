import toast from "react-hot-toast";
import { LuCheck, LuX } from "react-icons/lu";
import Icon from "./Icon";

const toastStyles = {
  success: {
    border: "border-green",
    borderLeft: "border-l-green",
    icon: <Icon Icon={LuCheck} size={14} className="!text-green stroke-3" />,
  },
  error: {
    border: "border-red",
    borderLeft: "border-l-red",
    icon: <Icon Icon={LuX} size={14} className="!text-red stroke-3" />,
  },
};

const CustomToast = ({ type = "success", title, message, t, toast }) => {
  const style = toastStyles[type];

  return (
    <div
      className={`max-w-md flex items-center gap-3 p-3 bg-theme-bg border border-theme-border rounded-sm shadow-lg border-l-4 ${
        style.borderLeft
      } transform transition-all duration-300 ease-in-out ${
        t.visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2"
      }`}
    >
      <div className={`p-1 rounded-full border-2 ${style.border}`}>
        {style.icon}
      </div>

      <div className="flex-1">
        <p className="font-semibold text-theme-text">{title}</p>
        <p className="text-theme-text whitespace-pre-line">{message}</p>
      </div>

      <button onClick={() => toast.dismiss(t.id)}>
        <Icon Icon={LuX} size={16} />
      </button>
    </div>
  );
};

export const successToast = (message, title) => {
  toast.custom(
    (t) => (
      <CustomToast
        t={t}
        toast={toast}
        type="success"
        title={title}
        message={message}
      />
    ),
    { position: "bottom-right" }
  );
};

export const errorToast = (message, title) => {
  toast.custom(
    (t) => (
      <CustomToast
        t={t}
        toast={toast}
        type="error"
        title={title}
        message={message}
      />
    ),
    { position: "bottom-right" }
  );
};

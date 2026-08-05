import { Icon } from "./Icon.jsx";

export function ToolbarButton({ icon, children, ...props }) {
  return (
    <button type="button" className="tb-btn" {...props}>
      {icon && (
        <span className="ico">
          <Icon name={icon} />
        </span>
      )}{" "}
      {children}
    </button>
  );
}

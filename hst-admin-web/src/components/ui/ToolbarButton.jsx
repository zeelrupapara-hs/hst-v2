import { Icon } from "@/components/ui/Icon.jsx";

export function ToolbarButton({ icon, children, ...props }) {
  return (
    <button type="button" className="tb-btn" {...props}>
      {icon && (
        <span className="ico">
          <Icon id={icon} />
        </span>
      )}{" "}
      {children}
    </button>
  );
}

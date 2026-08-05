export function CenteredCard({ title, description, children, className = "picker-card" }) {
  return (
    <div className={className === "picker-card" ? "picker" : "login-page"}>
      <div className={className}>
        {title && <h1>{title}</h1>}
        {description && <p>{description}</p>}
        {children}
      </div>
    </div>
  );
}

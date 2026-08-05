export function SearchField({ value, onChange, onSubmit, placeholder = "Search…" }) {
  return (
    <div className="filter-bar">
      <label>Search</label>
      <input
        type="search"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && onSubmit?.()}
        placeholder={placeholder}
      />
    </div>
  );
}

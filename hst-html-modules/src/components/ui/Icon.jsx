export function Icon({ name, title }) {
  return (
    <img
      className="hst-icon"
      src={`/assets/icons/svg/${name}.svg?v=3`}
      alt=""
      title={title}
    />
  );
}

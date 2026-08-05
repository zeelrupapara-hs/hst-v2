import {
  iconUrl,
  normalizeIconId,
  spriteRef,
} from "../../lib/icons.js";

/**
 * generated icon (from assets/icons/manifest.json + build-icons.py).
 * Prefers the combined sprite; falls back to individual SVG files.
 */
export function Icon({ name, title, size = 16, className = "hst-icon", useSprite = true }) {
  const id = normalizeIconId(name, "");
  if (!id) return null;

  if (useSprite) {
    return (
      <svg
        className={className}
        width={size}
        height={size}
        aria-hidden={title ? undefined : true}
        role={title ? "img" : undefined}
      >
        {title ? <title>{title}</title> : null}
        <use href={spriteRef(id)} />
      </svg>
    );
  }

  return (
    <img
      className={className}
      src={iconUrl(id)}
      width={size}
      height={size}
      alt=""
      title={title}
      draggable={false}
    />
  );
}

/** Navigator row icon — wraps Icon in the standard .ico cell. */
export function NavIcon({ name, title }) {
  const id = normalizeIconId(name, "");
  if (!id) return <span className="ico" aria-hidden="true" />;
  return (
    <span className="ico">
      <Icon name={id} title={title} size={16} />
    </span>
  );
}

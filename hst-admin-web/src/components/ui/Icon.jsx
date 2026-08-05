import { spriteRef } from "@/lib/icons.js";

/** One sprite icon; null id renders nothing rather than a broken box. */
export function Icon({ id, size = 16, title }) {
  if (!id) return null;

  return (
    <svg
      className="hst-icon"
      width={size}
      height={size}
      role={title ? "img" : undefined}
      aria-hidden={title ? undefined : true}
    >
      {title ? <title>{title}</title> : null}
      <use href={spriteRef(id)} />
    </svg>
  );
}

#!/usr/bin/env python3
"""Build MT5 navigation SVG icons from help-doc raster sources."""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("Pillow required: pip install Pillow", file=sys.stderr)
    sys.exit(1)

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "assets" / "icons" / "manifest.json"
OUT_DIR = ROOT / "assets" / "icons" / "svg"
CACHE_DIR = Path(__file__).resolve().parent / ".icon-cache"
VTRACER = Path(__file__).resolve().parent / ".tools" / "vtracer"
SVGO = ROOT / "node_modules" / ".bin" / "svgo"
SPRITE = ROOT / "assets" / "icons" / "sprite.svg"
VIEW = 32


def expand(path: str) -> Path:
    return Path(os.path.expanduser(path))


def strip_black_bg(im: Image.Image, threshold: int = 30) -> Image.Image:
    im = im.convert("RGBA")
    px = im.load()
    for y in range(im.height):
        for x in range(im.width):
            r, g, b, a = px[x, y]
            if r < threshold and g < threshold and b < threshold:
                px[x, y] = (0, 0, 0, 0)
    return im


def fit_canvas(im: Image.Image, size: int = VIEW, nearest: bool = False) -> Image.Image:
    im = im.convert("RGBA")
    bbox = im.getbbox()
    if not bbox:
        return Image.new("RGBA", (size, size), (0, 0, 0, 0))
    im = im.crop(bbox)
    w, h = im.size
    pad = 4
    scale = min((size - pad) / w, (size - pad) / h)
    nw = max(1, int(w * scale))
    nh = max(1, int(h * scale))
    resample = Image.Resampling.NEAREST if nearest else Image.Resampling.LANCZOS
    im = im.resize((nw, nh), resample)
    canvas = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    canvas.paste(im, ((size - nw) // 2, (size - nh) // 2), im)
    return canvas


def prep_raster(entry: dict, help_dir: Path) -> Image.Image:
    src = help_dir / entry["source"]
    if not src.exists():
        raise FileNotFoundError(f"Missing source: {src}")

    nearest = entry.get("nearest", False)

    if "crop" in entry:
        im = Image.open(src).convert("RGBA")
        x0, y0, x1, y1 = entry["crop"]
        # Expand crop slightly for cleaner edges, then scale to VIEW
        pad = 2
        im = im.crop((max(0, x0 - pad), max(0, y0 - pad), x1 + pad, y1 + pad))
        nearest = True
    else:
        im = Image.open(src)

    if entry.get("stripBg"):
        im = strip_black_bg(im)
    elif im.mode == "P":
        im = im.convert("RGBA")

    upscale = entry.get("upscale", 0)
    if upscale and max(im.size) < VIEW:
        factor = upscale
        im = im.resize((im.width * factor, im.height * factor), Image.Resampling.NEAREST)
        nearest = True

    if entry.get("flipH"):
        im = im.transpose(Image.Transpose.FLIP_LEFT_RIGHT)

    return fit_canvas(im, VIEW, nearest=nearest)


def run_vtracer(png: Path, svg: Path) -> None:
    if not VTRACER.exists():
        raise FileNotFoundError(f"vtracer not found at {VTRACER}")
    cmd = [
        str(VTRACER),
        "-i", str(png),
        "-o", str(svg),
        "--colormode", "color",
        "--mode", "polygon",
        "--filter_speckle", "2",
        "--color_precision", "8",
    ]
    subprocess.run(cmd, check=True, capture_output=True)


def optimize_svg(svg: Path) -> None:
    if SVGO.exists():
        subprocess.run([str(SVGO), str(svg), "-o", str(svg)], check=True, capture_output=True)


def normalize_svg(svg: Path) -> None:
    text = svg.read_text(encoding="utf-8")
    text = re.sub(r'\swidth="[^"]*"', "", text, count=1)
    text = re.sub(r'\sheight="[^"]*"', "", text, count=1)
    if "viewBox=" not in text:
        text = text.replace("<svg", f'<svg viewBox="0 0 {VIEW} {VIEW}"', 1)
    else:
        text = re.sub(
            r'viewBox="[^"]*"',
            f'viewBox="0 0 {VIEW} {VIEW}"',
            text,
            count=1,
        )
    title = svg.stem.replace("-", " ").title()
    if "<title>" not in text:
        text = text.replace(
            "<svg",
            f'<svg role="img" aria-label="{title}"',
            1,
        )
        text = text.replace(">", f"><title>{title}</title>", 1)
    svg.write_text(text, encoding="utf-8")


def build_sprite(svg_dir: Path, icon_ids: list[str]) -> None:
    parts = [
        '<?xml version="1.0" encoding="UTF-8"?>',
        '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="display:none">',
    ]
    for icon_id in icon_ids:
        path = svg_dir / f"{icon_id}.svg"
        if not path.exists():
            continue
        inner = path.read_text(encoding="utf-8")
        m = re.search(r"<svg[^>]*>(.*)</svg>", inner, re.DOTALL)
        if not m:
            continue
        parts.append(f'  <symbol id="icon-{icon_id}" viewBox="0 0 {VIEW} {VIEW}">')
        parts.append(m.group(1).strip())
        parts.append("  </symbol>")
    parts.append("</svg>")
    SPRITE.write_text("\n".join(parts) + "\n", encoding="utf-8")


def main() -> int:
    manifest = json.loads(MANIFEST.read_text(encoding="utf-8"))
    help_dir = expand(manifest.get("helpDir", "~/Documents/mt5administrator-help"))
    if not help_dir.exists():
        print(f"MT5 help dir not found: {help_dir}", file=sys.stderr)
        return 1

    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    OUT_DIR.mkdir(parents=True, exist_ok=True)

    icons = manifest["icons"]
    report: list[tuple[str, int, str]] = []

    for entry in icons:
        icon_id = entry["id"]
        png = CACHE_DIR / f"{icon_id}.png"
        svg = OUT_DIR / f"{icon_id}.svg"
        try:
            im = prep_raster(entry, help_dir)
            im.save(png)
            run_vtracer(png, svg)
            optimize_svg(svg)
            normalize_svg(svg)
            size = svg.stat().st_size
            flag = "WARN" if size > 4096 else "OK"
            report.append((icon_id, size, flag))
            print(f"  [{flag}] {icon_id}.svg ({size} bytes)")
        except Exception as exc:
            print(f"  [FAIL] {icon_id}: {exc}", file=sys.stderr)
            report.append((icon_id, 0, "FAIL"))

    build_sprite(OUT_DIR, [e["id"] for e in icons])
    print(f"\nSprite: {SPRITE} ({SPRITE.stat().st_size} bytes)")

    failed = [r for r in report if r[2] == "FAIL"]
    large = [r for r in report if r[2] == "WARN"]
    total = sum(r[1] for r in report)
    print(f"Total: {len(report)} icons, {total // 1024} KB (viewBox {VIEW}x{VIEW})")
    if large:
        print(f"Large icons ({len(large)}): {', '.join(r[0] for r in large)}")
    if failed:
        print(f"Failed ({len(failed)}): {', '.join(r[0] for r in failed)}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""Convert a PNG image to a Go []string sprite array.

Transparent and white pixels become '.', all others become 'X'.

Usage:
    python scripts/png2sprite.py <input.png> [--name varName] [--threshold 240]

Examples:
    python scripts/png2sprite.py assets/sectoid_head.png
    python scripts/png2sprite.py head.png --name headDark --threshold 200
    python scripts/png2sprite.py torso.png --name torsoWings --format bool
"""

import argparse
import sys
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("Requires Pillow: pip install Pillow", file=sys.stderr)
    sys.exit(1)


def load_pixels(path: str):
    img = Image.open(path).convert("RGBA")
    return img.getdata(), img.width, img.height


def is_empty(r, g, b, a, threshold: int) -> bool:
    if a < 128:
        return True
    if r >= threshold and g >= threshold and b >= threshold:
        return True
    return False


def to_strings(pixels, w, h, threshold: int) -> list[str]:
    rows = []
    for y in range(h):
        row = []
        for x in range(w):
            r, g, b, a = pixels[y * w + x]
            row.append("." if is_empty(r, g, b, a, threshold) else "X")
        rows.append("".join(row))
    return rows


def to_bool_grid(pixels, w, h, threshold: int) -> list[str]:
    rows = []
    for y in range(h):
        row = []
        for x in range(w):
            r, g, b, a = pixels[y * w + x]
            row.append("false" if is_empty(r, g, b, a, threshold) else "true")
        rows.append("    {" + ", ".join(row) + "},\n")
    return rows


def format_go_strings(name: str, rows: list[str]) -> str:
    lines = [f"var {name} = []string{{\n"]
    for row in rows:
        lines.append(f'\t"{row}",\n')
    lines.append("}\n")
    return "".join(lines)


def format_go_bool(name: str, rows: list[str]) -> str:
    lines = [f"var {name} = [][20]bool{{\n"]
    lines.extend(rows)
    lines.append("}\n")
    return "".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Convert PNG to Go sprite array")
    parser.add_argument("input", help="Input PNG file")
    parser.add_argument("--name", default="sprite", help="Go variable name (default: sprite)")
    parser.add_argument("--threshold", type=int, default=240,
                        help="White threshold 0-255 (default: 240)")
    parser.add_argument("--format", choices=["string", "bool"], default="string",
                        help="Output format: string or bool (default: string)")
    parser.add_argument("--width", type=int, default=0,
                        help="Resize to this width before conversion (preserves pixels)")
    args = parser.parse_args()

    path = Path(args.input)
    if not path.exists():
        print(f"File not found: {path}", file=sys.stderr)
        sys.exit(1)

    img = Image.open(path).convert("RGBA")
    if args.width > 0 and args.width != img.width:
        ratio = args.width / img.width
        new_h = int(img.height * ratio)
        img = img.resize((args.width, new_h), Image.NEAREST)

    pixels = list(img.getdata())
    w, h = img.size

    rows = to_strings(pixels, w, h, args.threshold)

    if args.format == "bool":
        bool_rows = to_bool_grid(pixels, w, h, args.threshold)
        output = format_go_bool(args.name, bool_rows)
    else:
        output = format_go_strings(args.name, rows)

    print(f"// {w}x{h} pixels from {path.name}")
    print(output)


if __name__ == "__main__":
    main()

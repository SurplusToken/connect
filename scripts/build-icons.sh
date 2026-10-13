#!/bin/sh
# macOS asset conversion only; original artwork remains unchanged.
set -eu
cd "$(dirname "$0")/.."
source_image=assets/branding/susu-white.png
iconset_dir=$(mktemp -d "${TMPDIR:-/tmp}/surplustoken-icons.XXXXXX")
trap 'rm -rf "$iconset_dir"' EXIT
mkdir -p "$iconset_dir/AppIcon.iconset" frontend/public/brand build/darwin
swift scripts/render-app-icon.swift "$source_image" build/darwin/AppIcon.png
for size in 16 32 128 256 512; do
  sips -z "$size" "$size" build/darwin/AppIcon.png --out "$iconset_dir/AppIcon.iconset/icon_${size}x${size}.png" >/dev/null
  retina_size=$((size * 2))
  sips -z "$retina_size" "$retina_size" build/darwin/AppIcon.png --out "$iconset_dir/AppIcon.iconset/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil -c icns "$iconset_dir/AppIcon.iconset" -o build/darwin/AppIcon.icns
sips -z 128 128 "$source_image" --out frontend/public/brand/susu.png >/dev/null
sips -z 32 32 "$source_image" --out frontend/public/brand/favicon.png >/dev/null

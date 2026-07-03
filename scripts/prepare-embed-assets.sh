#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target="${root_dir}/internal/server/static/dist"

require_dir() {
  local path="$1"
  if [ ! -d "${path}" ]; then
    echo "missing frontend build output: ${path}" >&2
    exit 1
  fi
}

require_dir "${root_dir}/web/admin/dist"
require_dir "${root_dir}/web/user/dist"
require_dir "${root_dir}/docs-site/.vitepress/dist"

rm -rf "${target}"
mkdir -p "${target}"

# Go embed can only read files below the package directory, so release builds
# copy frontend outputs into this package-local dist tree before compilation.
cp -R "${root_dir}/web/admin/dist" "${target}/admin"
cp -R "${root_dir}/web/user/dist" "${target}/user"
cp -R "${root_dir}/docs-site/.vitepress/dist" "${target}/docs"

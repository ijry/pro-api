#!/usr/bin/env bash
set -euo pipefail

goos="${1:?goos required}"
goarch="${2:?goarch required}"
version="${3:?version required}"
binary_path="${4:?binary path required}"
output_dir="${5:?output dir required}"

mkdir -p "${output_dir}"

base="proapi_${goos}_${goarch}"
stage_dir="${output_dir}/${base}"
rm -rf "${stage_dir}"
mkdir -p "${stage_dir}"

binary_name="proapi"
if [ "${goos}" = "windows" ]; then
  binary_name="proapi.exe"
fi

cp "${binary_path}" "${stage_dir}/${binary_name}"
cp LICENSE "${stage_dir}/LICENSE"
cp README.md "${stage_dir}/README.md"

archive_path=""
if [ "${goos}" = "windows" ]; then
  archive_path="${output_dir}/${base}.zip"
  rm -f "${archive_path}"
  (
    cd "${stage_dir}"
    zip -q -r "../$(basename "${archive_path}")" .
  )
else
  archive_path="${output_dir}/${base}.tar.gz"
  tar -C "${stage_dir}" -czf "${archive_path}" .
fi

archive_name="$(basename "${archive_path}")"
if command -v sha256sum >/dev/null 2>&1; then
  (
    cd "${output_dir}"
    sha256sum "${archive_name}" > "${archive_name}.sha256"
  )
else
  (
    cd "${output_dir}"
    shasum -a 256 "${archive_name}" > "${archive_name}.sha256"
  )
fi

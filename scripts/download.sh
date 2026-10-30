#!/bin/bash
set -euo pipefail

goos="$(printf '%s' "${OS}" | tr '[:upper:]' '[:lower:]')"
[[ "${goos}" == "macos" ]] && goos="darwin"

goarch="$(printf '%s' "${ARCH}" | tr '[:upper:]' '[:lower:]' | sed 's/x64/amd64/')"

target="${goos}_${goarch}"
[[ "${target}" =~ ^(linux|darwin)_(amd64|arm64)$ ]] || {
  echo "::error::synac publishes no binary for ${OS}/${ARCH}."
  exit 1
}

ver="${VERSION:-latest}"

if [[ "${ver}" == "latest" ]]; then
  ver="$(curl -fsSL --retry 3 \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    ${GH_TOKEN:+-H "Authorization: Bearer ${GH_TOKEN}"} \
    "https://api.github.com/repos/${REPO}/releases/latest" | grep -oE '"tag_name": *"[^"]+"' | head -n 1 | sed -E 's/.*\"([^\"]+)\"/\1/' || true)"
  if [[ -z "${ver}" ]]; then
    echo "::error::Cannot resolve the latest release for ${REPO}."
    exit 1
  fi
  echo "Resolved latest to ${ver}."
elif [[ "${ver}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  :
else
  echo "::error::Unsupported version ${ver}. Use a version tag like v0.1.0 or \"latest\"."
  exit 1
fi

dir="${RUNNER_TEMP}/synac"
rm -rf "${dir}"
mkdir -p "${dir}"

base="https://github.com/${REPO}/releases/download/${ver}"
asset="synac_${ver#v}_${target}.tar.gz"
curl -fsSL --retry 3 -o "${dir}/${asset}" "${base}/${asset}" || {
  echo "::error::Cannot download ${asset}. Does tag ${ver} have release assets?"
  exit 1
}
curl -fsSL --retry 3 -o "${dir}/checksums.txt" "${base}/synac_${ver#v}_checksums.txt" || {
  echo "::error::Cannot download the checksums file. Does tag ${ver} have release assets?"
  exit 1
}

expected="$(awk -v a="${asset}" '$2 == a { print $1 }' "${dir}/checksums.txt")"
actual="$(openssl dgst -sha256 "${dir}/${asset}" | awk '{ print $NF }')"
if [[ -z "${expected}" || "${actual}" != "${expected}" ]]; then
  echo "::error::synac ${asset} is missing from the release or its checksum mismatched. Is ${ver} released?"
  exit 1
fi

tar -xzf "${dir}/${asset}" -C "${dir}"
echo "binary=${dir}/synac" >> "${GITHUB_OUTPUT}"

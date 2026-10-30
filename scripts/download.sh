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

ref="${REF#refs/tags/}"
ver="${ref}"

if [[ "${ver}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  :

# Floating tags (v0, v0.1) have no release assets, so resolve them to the highest matching stable version first.
elif [[ "${ver}" =~ ^v[0-9]+(\.[0-9]+)?$ ]]; then
  prefix="${ver}."
  tags="$(git ls-remote --tags --sort=v:refname "https://github.com/${REPO}.git" "refs/tags/${prefix}*" || true)"

  ver="$(printf '%s\n' "${tags}" | grep -v '\^{}' | grep -oE 'refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$' | grep -F "refs/tags/${prefix}" | tail -n 1 | sed 's|.*/||' || true)"
  if [[ -z "${ver}" ]]; then
    echo "::error::No stable release found for ${ref}."
    exit 1
  fi

  echo "Resolved ${ref} to ${ver}."

# A pinned commit SHA is resolved back to its tag via ls-remote, so pinning to a full-length SHA keeps working.
elif [[ "${ver}" =~ ^[0-9a-f]{40}$ ]]; then
  sha="${ver}"
  tags="$(git ls-remote --tags "https://github.com/${REPO}.git" || true)"

  ver="$(printf '%s\n' "${tags}" | grep "^${sha}[[:space:]]" | grep -oE 'refs/tags/v[0-9]+\.[0-9]+\.[0-9]+(\^\{\})?$' | sed 's/\^{}$//' | sed 's|.*/||' | sort -V | tail -n 1 || true)"
  if [[ -z "${ver}" ]]; then
    echo "::error::No release tag found for commit ${sha}."
    exit 1
  fi

  echo "Resolved ${sha} to ${ver}."

else
  echo "::error::Unsupported ref ${ref}. Use a version tag like v0.1.0, a floating tag like v0, or a full-length commit SHA."
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

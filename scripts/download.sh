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

if [[ "${ver}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  :

# Floating versions (v0, v0.1) have no release assets, so resolve them to the highest matching stable version first.
elif [[ "${ver}" =~ ^v[0-9]+(\.[0-9]+)?$ ]]; then
  prefix="${ver}."
  tags="$(git ls-remote --tags --sort=v:refname "https://github.com/${REPO}.git" "refs/tags/${prefix}*" || true)"

  candidates="$(printf '%s\n' "${tags}" | grep -v '\^{}' | grep -oE 'refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$' | grep -F "refs/tags/${prefix}" | sed 's|.*/||' | sort -Vr || true)"
  ver=""
  while IFS= read -r candidate; do
    [[ -n "${candidate}" ]] || continue
    candidate_base="https://github.com/${REPO}/releases/download/${candidate}"
    candidate_asset="synac_${candidate#v}_${target}.tar.gz"
    asset_code="$(curl -fsSLI --retry 3 -o /dev/null -w "%{http_code}" "${candidate_base}/${candidate_asset}" || true)"
    if [[ "${asset_code}" == "404" ]]; then
      continue
    elif [[ "${asset_code}" != "200" ]]; then
      echo "::error::Cannot probe ${candidate} (asset HTTP ${asset_code})."
      exit 1
    fi
    checksum_code="$(curl -fsSLI --retry 3 -o /dev/null -w "%{http_code}" "${candidate_base}/synac_${candidate#v}_checksums.txt" || true)"
    if [[ "${checksum_code}" == "200" ]]; then
      ver="${candidate}"
      break
    elif [[ "${checksum_code}" == "404" ]]; then
      continue
    else
      echo "::error::Cannot probe ${candidate} (checksums HTTP ${checksum_code})."
      exit 1
    fi
  done <<< "${candidates}"

  if [[ -z "${ver}" ]]; then
    echo "::error::No stable release found for ${VERSION}."
    exit 1
  fi

  echo "Resolved ${VERSION} to ${ver}."

elif [[ "${ver}" == "latest" ]]; then
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

else
  echo "::error::Unsupported version ${ver}. Use a version tag like v0.1.0, a floating version like v0, or \"latest\"."
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

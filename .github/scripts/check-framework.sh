#!/usr/bin/env bash

set -euo pipefail

framework_dir="${1:-}"

if [[ -z "${framework_dir}" ]]; then
  echo "Usage: $0 <framework-dir>"
  exit 1
fi

mapfile -t modules < <(
  {
    if [[ -f "${framework_dir}/go.mod" ]]; then
      echo "${framework_dir}"
    fi
    if [[ -d "${framework_dir}" ]]; then
      find "${framework_dir}" -mindepth 1 -type f -name go.mod -printf '%h\n'
    fi
  } | sort -u
)

if [[ "${#modules[@]}" -eq 0 ]]; then
  echo "No Go modules found under ${framework_dir}, skipping."
  exit 0
fi

printf 'Discovered modules under %s:\n' "${framework_dir}"
printf ' - %s\n' "${modules[@]}"

for module in "${modules[@]}"; do
  echo "::group::Checking ${module}"

  mapfile -d '' -t go_files < <(find "${module}" -type f -name '*.go' -not -path '*/vendor/*' -print0)
  if [[ "${#go_files[@]}" -eq 0 ]]; then
    echo "No Go files found in ${module}, skipping."
    echo "::endgroup::"
    continue
  fi

  fmt_output="$(gofmt -l "${go_files[@]}")"
  if [[ -n "${fmt_output}" ]]; then
    echo "The following files need gofmt:"
    echo "${fmt_output}"
    exit 1
  fi

  if ! package_output="$(cd "${module}" && go list ./... 2>&1)"; then
    if grep -q "matched no packages" <<<"${package_output}"; then
      echo "No Go packages found in ${module}, skipping go vet and go test."
      echo "::endgroup::"
      continue
    fi

    echo "${package_output}"
    exit 1
  fi

  package_lines="$(printf '%s\n' "${package_output}" | grep -v '^go:' | sed '/^[[:space:]]*$/d' || true)"
  if [[ -z "${package_lines}" ]]; then
    echo "No Go packages found in ${module}, skipping go vet and go test."
    echo "::endgroup::"
    continue
  fi

  (
    cd "${module}"
    go vet ./...
    go test ./...
  )

  echo "::endgroup::"
done

#!/usr/bin/env bash

set -euo pipefail

required_docs=(
  "docs/README.md"
  "docs/architecture/template-boundaries.md"
  "docs/architecture/addon-design-rules.md"
  "docs/architecture/repository-rules.md"
  "docs/guides/template-selection.md"
  "docs/roadmap/roadmap.md"
)

for file in "${required_docs[@]}"; do
  if [[ ! -f "${file}" ]]; then
    echo "Missing required docs file: ${file}"
    exit 1
  fi
done

templates=(heavy medium light extra-light)

for template in "${templates[@]}"; do
  module_dir="v3/${template}"
  config_path="${module_dir}/config/server.yaml"
  readme_en="${module_dir}/README.md"
  readme_zh="${module_dir}/README_zh.md"

  echo "::group::Contract checks for ${template}"

  for file in "${config_path}" "${readme_en}" "${readme_zh}"; do
    if [[ ! -f "${file}" ]]; then
      echo "Missing required template file: ${file}"
      exit 1
    fi
  done

  case "${template}" in
    heavy)
      required_keys=(server database redis prometheus log middleware waf schedule)
      embed_go="${module_dir}/internal/transport/http/webui/embed.go"
      embed_dist="${module_dir}/internal/transport/http/webui/dist"
      ;;
    medium)
      required_keys=(server database redis log middleware waf)
      embed_go="${module_dir}/internal/transport/http/webui/embed.go"
      embed_dist="${module_dir}/internal/transport/http/webui/dist"
      ;;
    light)
      required_keys=(server database log middleware)
      embed_go="${module_dir}/internal/transport/http/webui/embed.go"
      embed_dist="${module_dir}/internal/transport/http/webui/dist"
      ;;
    extra-light)
      required_keys=(server database log)
      embed_go="${module_dir}/internal/http/webui/embed.go"
      embed_dist="${module_dir}/internal/http/webui/dist"
      ;;
    *)
      echo "Unknown template: ${template}"
      exit 1
      ;;
  esac

  for key in "${required_keys[@]}"; do
    if ! grep -Eq "^${key}:" "${config_path}"; then
      echo "Missing config section '${key}' in ${config_path}"
      exit 1
    fi
    if ! grep -Fq "\`${key}\`" "${readme_en}"; then
      echo "README.md for ${template} does not mention \`${key}\`"
      exit 1
    fi
    if ! grep -Fq "\`${key}\`" "${readme_zh}"; then
      echo "README_zh.md for ${template} does not mention \`${key}\`"
      exit 1
    fi
  done

  if [[ ! -f "${embed_go}" ]]; then
    echo "Missing embed entry file: ${embed_go}"
    exit 1
  fi
  if [[ ! -d "${embed_dist}" ]]; then
    echo "Missing embedded UI dist directory: ${embed_dist}"
    exit 1
  fi

  (
    cd "${module_dir}"
    go run . version --config config/server.yaml > /dev/null
  )

  echo "::endgroup::"
done

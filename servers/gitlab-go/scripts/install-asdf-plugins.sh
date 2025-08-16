#!/usr/bin/env bash
# Vendored from https://gitlab.com/gitlab-com/gl-infra/common-template-copier
# Consider contributing upstream when updating this file

# This file is deprecated: going forward running `mise install` should be sufficient.

set -euo pipefail
IFS=$'\n\t'

echo >&2 -e "2024-08-07: this file is deprecated: going forward, simply run 'mise install' to install plugins."
echo >&2 -e "Recommended reading: https://gitlab.com/gitlab-com/gl-infra/common-ci-tasks/-/blob/main/docs/developer-setup.md"

mise install

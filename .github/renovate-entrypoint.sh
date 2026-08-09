#!/usr/bin/env bash
set -euo pipefail


# renovate: datasource=github-releases depName=norwoodj/helm-docs
HELM_DOCS_VERSION=1.14.2
curl -fsSL \
    "https://github.com/norwoodj/helm-docs/releases/download/v${HELM_DOCS_VERSION#v}/helm-docs_${HELM_DOCS_VERSION#v}_Linux_x86_64.tar.gz" \
    | tar -xz -C /usr/local/bin helm-docs
helm-docs --version

runuser -u ubuntu renovate

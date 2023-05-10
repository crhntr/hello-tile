#!/usr/bin/env bash

set -euxo pipefail

: "${TILE_VERSION:?}"
: "${KILN_VERSION:?}"
: "${GITHUB_TOKEN:?}"

cd "${GITHUB_WORKSPACE}" || exit 1

# Install Kiln
curl --fail -L "https://github.com/pivotal-cf/kiln/releases/download/v${KILN_VERSION}/kiln-linux-amd64-${KILN_VERSION}" --output kiln
chmod +x kiln
mv kiln /usr/local/bin/

echo "Kiln Version: $(kiln version)"
echo "Tile Version: ${TILE_VERSION}"

printf "%s" "${TILE_VERSION/#v}" > version

export CREDENTIALS_FILE=/tmp/credentials.yml
echo "github_token: '${GITHUB_TOKEN}'" > "${CREDENTIALS_FILE}"

kiln fetch --variables-file "${CREDENTIALS_FILE}"
kiln validate --variables-file "${CREDENTIALS_FILE}"
kiln bake --variables-file variables/hello.yml --variables-file "${CREDENTIALS_FILE}"
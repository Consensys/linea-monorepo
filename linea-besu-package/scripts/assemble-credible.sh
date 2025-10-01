#!/bin/bash

source ./scripts/assemble-packages.sh

cd ..

# Required parameters
OWNER="phylaxsystems"
REPO="credible-layer-besu-plugin"
GROUP_ID="net.phylax.credible"
ARTIFACT_ID="credible-layer-besu-plugin"
VERSION="0.1.0-b497f885"

OUTPUT_LOC="./tmp/besu/plugins/$ARTIFACT_ID-$VERSION.jar"

# Download using curl
response=$(curl -s -w "%{http_code}" -L -H "Authorization: token $GH_TOKEN" \
     -H "Accept: application/octet-stream" \
     "https://maven.pkg.github.com/$OWNER/$REPO/$GROUP_ID/$ARTIFACT_ID/$VERSION/$ARTIFACT_ID-$VERSION.jar" \
     -o "$OUTPUT_LOC")

http_code=${response: -3}

if [ "$http_code" -eq 200 ]; then
    echo "✅ Download successful!"
else
    echo "❌ Download failed with HTTP code: $http_code"
    echo "💡 Check:"
    echo "   - Token has 'read:packages' scope"
    echo "   - You have access to the repository"
    echo "   - Package coordinates are correct"
    exit 1
fi
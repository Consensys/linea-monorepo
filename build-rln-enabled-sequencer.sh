#!/bin/bash
set -e

echo "🚀 Building Simple RLN-Enabled Linea Sequencer"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build paths - Updated for correct directory structure
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LINEA_SEQUENCER_DIR="${SCRIPT_DIR}/besu-plugins/linea-sequencer"
STATUS_RLN_PROVER_DIR="/Users/nadeem/dev/status/linea/status-rln-prover"
CUSTOM_BESU_DIR="${SCRIPT_DIR}/custom-besu-package"
LINEA_BESU_DIR="/Users/nadeem/dev/status/linea/linea-besu"

echo -e "${BLUE}📁 Working directories:${NC}"
echo -e "  Linea Sequencer: ${LINEA_SEQUENCER_DIR}"
echo -e "  Status RLN Prover: ${STATUS_RLN_PROVER_DIR}"
echo -e "  Custom Besu Package: ${CUSTOM_BESU_DIR}"
echo -e "  Linea Besu Repository: ${LINEA_BESU_DIR}"

# Check if directories exist
for dir in "$LINEA_SEQUENCER_DIR" "$STATUS_RLN_PROVER_DIR" "$LINEA_BESU_DIR"; do
    if [[ ! -d "$dir" ]]; then
        echo -e "${RED}❌ Error: Directory not found: $dir${NC}"
        exit 1
    fi
done

echo -e "${BLUE}🦀 Building RLN Bridge Rust Library for Linux...${NC}"
cd "${LINEA_SEQUENCER_DIR}/sequencer/src/main/rust/rln_bridge"

# Check if Linux library already exists
RLN_LIB_FILE="${LINEA_SEQUENCER_DIR}/sequencer/src/main/rust/rln_bridge/target/x86_64-unknown-linux-gnu/release/librln_bridge.so"

if [[ ! -f "$RLN_LIB_FILE" ]]; then
    echo -e "${YELLOW}🐳 Cross-compiling Rust library for Linux x86-64...${NC}"
    # Use Docker to cross-compile for Linux
    docker run --rm --platform linux/amd64 \
        -v "$(pwd)":/workspace \
        -w /workspace \
        rust:1.85-bookworm bash -c "
            set -e
            apt-get update -qq
            apt-get install -y -qq pkg-config libssl-dev build-essential
            rustup target add x86_64-unknown-linux-gnu
            cargo build --release --target x86_64-unknown-linux-gnu
        "
fi
if [[ ! -f "$RLN_LIB_FILE" ]]; then
    echo -e "${RED}❌ Error: Linux RLN library not found at: $RLN_LIB_FILE${NC}"
    exit 1
fi

# Verify it's actually a Linux .so file
LIB_INFO=$(file "$RLN_LIB_FILE")
if echo "$LIB_INFO" | grep -q "ELF.*x86-64"; then
    echo -e "${GREEN}✅ RLN Bridge library built for Linux: $RLN_LIB_FILE${NC}"
else
    echo -e "${RED}❌ Error: Library is not Linux x86-64 format: $LIB_INFO${NC}"
    exit 1
fi

echo -e "${BLUE}☕ Building Custom Sequencer JAR...${NC}"
cd "$SCRIPT_DIR"
./gradlew clean :besu-plugins:linea-sequencer:sequencer:build -x test -x checkSpdxHeader -x spotlessJavaCheck -x spotlessGroovyGradleCheck --no-daemon

# Find the built JAR
SEQUENCER_JAR=$(find "${LINEA_SEQUENCER_DIR}/sequencer/build/libs" -name "linea-sequencer-*.jar" | head -1)
if [[ ! -f "$SEQUENCER_JAR" ]]; then
    echo -e "${RED}❌ Error: Sequencer JAR not found!${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Custom Sequencer JAR built: $SEQUENCER_JAR${NC}"

echo -e "${BLUE}🦀 Building RLN Prover Service...${NC}"
cd "$STATUS_RLN_PROVER_DIR"
cargo build --release

PROVER_BINARY="${STATUS_RLN_PROVER_DIR}/target/release/status_rln_prover"
if [[ ! -f "$PROVER_BINARY" ]]; then
    echo -e "${RED}❌ Error: RLN Prover binary not found!${NC}"
    exit 1
fi
echo -e "${GREEN}✅ RLN Prover service built: $PROVER_BINARY${NC}"

echo -e "${BLUE}🐳 Building Custom Besu Docker Image...${NC}"
mkdir -p "$CUSTOM_BESU_DIR"
cd "$CUSTOM_BESU_DIR"

# Get the working official image and extract its structure
BESU_PACKAGE_TAG="beta-v2.1-rc16.2-20250521124830-4d89458"
echo -e "${YELLOW}📥 Extracting base Besu from official image...${NC}"
docker create --name temp-besu-extract "consensys/linea-besu-package:${BESU_PACKAGE_TAG}"
docker cp temp-besu-extract:/opt/besu/ ./besu/
docker rm temp-besu-extract

# Verify all required Linea plugins are present
echo -e "${YELLOW}🔍 Verifying Linea plugins...${NC}"
REQUIRED_PLUGINS=(
    "linea-staterecovery-besu-plugin"
    "linea-tracer"
    "linea-finalized-tag-updater"
    "besu-shomei-plugin"
)

echo -e "  Current plugins in extracted image:"
ls -la ./besu/plugins/ || echo "  No plugins directory found!"

for plugin in "${REQUIRED_PLUGINS[@]}"; do
    if ls ./besu/plugins/*${plugin}*.jar 1> /dev/null 2>&1; then
        echo -e "  ✅ Found: $plugin"
    else
        echo -e "  ⚠️  Missing: $plugin (will be included from base image)"
    fi
done

# Replace the sequencer plugin with our custom one
echo -e "${YELLOW}🔄 Installing custom sequencer...${NC}"
rm -f ./besu/plugins/linea-sequencer-*.jar
cp "$SEQUENCER_JAR" ./besu/plugins/
echo -e "  ✅ Installed: $(basename "$SEQUENCER_JAR")"

# Copy RLN native library
echo -e "${YELLOW}📚 Installing RLN native library...${NC}"
mkdir -p ./besu/lib/native
cp "$RLN_LIB_FILE" ./besu/lib/native/
echo -e "  ✅ Installed: librln_bridge.so"

# Update Besu startup scripts to include plugins in classpath
echo -e "${YELLOW}⚙️ Updating Besu startup scripts...${NC}"
for script in besu besu.bat besu-untuned besu-untuned.bat; do
    if [[ -f "./besu/bin/$script" ]]; then
        # Create backup
        cp "./besu/bin/$script" "./besu/bin/$script.backup"
        
        if [[ "$script" == *.bat ]]; then
            # Windows batch files
            sed -i.tmp 's|CLASSPATH=%APP_HOME%\\lib\\*|CLASSPATH=%APP_HOME%\\lib\\*;%APP_HOME%\\plugins\\*|g' "./besu/bin/$script"
        else
            # Unix shell scripts
            sed -i.tmp 's|CLASSPATH=\$APP_HOME/lib/\*|CLASSPATH=\$APP_HOME/lib/\*:\$APP_HOME/plugins/\*|g' "./besu/bin/$script"
        fi
        rm -f "./besu/bin/$script.tmp"
        echo -e "  ✅ Updated classpath in $script"
    else
        echo -e "  ⚠️  Script not found: $script"
    fi
done

# Verify final plugin structure
echo -e "${YELLOW}📋 Final plugin inventory:${NC}"
ls -la ./besu/plugins/ | grep -E "\.(jar|JAR)$" | while read -r line; do
    echo -e "  📦 $line"
done

# Create simple Dockerfile
cat > Dockerfile << 'EOF'
FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y openjdk-21-jre-headless libjemalloc-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    (groupadd -g 1000 besu || true) && \
    useradd -u 1000 -g 1000 -m -s /bin/bash besu || \
    (userdel -r besu 2>/dev/null || true && groupdel besu 2>/dev/null || true && \
     groupadd -g 1001 besu && useradd -u 1001 -g besu -m -s /bin/bash besu)

USER besu
WORKDIR /opt/besu

# Copy entire Besu distribution with custom sequencer
COPY --chown=besu:besu besu/ /opt/besu/

# Set library paths for RLN
ENV LD_LIBRARY_PATH="/opt/besu/lib/native:/usr/local/lib:/usr/lib"
ENV JAVA_LIBRARY_PATH="/opt/besu/lib/native"
ENV PATH="/opt/besu/bin:${PATH}"

EXPOSE 8545 8546 8547 8550 8551 30303

ENTRYPOINT ["besu"]
HEALTHCHECK --start-period=5s --interval=5s --timeout=1s --retries=10 CMD bash -c "[ -f /tmp/pid ]"
EOF

# Build custom Docker images
TIMESTAMP=$(date +%Y%m%d%H%M%S)
BESU_IMAGE_TAG="linea-besu-custom-sequencer:${TIMESTAMP}"
docker build -t "$BESU_IMAGE_TAG" .

echo -e "${GREEN}✅ Custom Besu image built: $BESU_IMAGE_TAG${NC}"

echo -e "${BLUE}🐳 Building RLN Prover Docker image...${NC}"
cd "$STATUS_RLN_PROVER_DIR"
RLN_PROVER_TAG="status-rln-prover:${TIMESTAMP}"
docker build -t "$RLN_PROVER_TAG" .

echo -e "${GREEN}✅ RLN Prover image built: $RLN_PROVER_TAG${NC}"

echo -e "${BLUE}📝 Updating Docker Compose...${NC}"
COMPOSE_FILE="${SCRIPT_DIR}/docker/compose-spec-l2-services-rln.yml"
if [[ -f "$COMPOSE_FILE" ]]; then
    # Create backup
    cp "$COMPOSE_FILE" "${COMPOSE_FILE}.backup.$(date +%Y%m%d%H%M%S)"
    
    # Update image tags
    sed -i.tmp "s|image: linea-besu-custom-sequencer:.*|image: ${BESU_IMAGE_TAG}|g" "$COMPOSE_FILE"
    sed -i.tmp "s|image: status-rln-prover:.*|image: ${RLN_PROVER_TAG}|g" "$COMPOSE_FILE"
    rm -f "${COMPOSE_FILE}.tmp"
    
    echo -e "${GREEN}✅ Updated Docker Compose with new images:${NC}"
    echo -e "  Besu: $BESU_IMAGE_TAG"
    echo -e "  RLN Prover: $RLN_PROVER_TAG"
fi

echo -e "${GREEN}🎉 Build Complete!${NC}"
echo -e "${BLUE}📋 Built Components:${NC}"
echo -e "  Custom Sequencer JAR: $(basename "$SEQUENCER_JAR")"
echo -e "  RLN Library: librln_bridge.so (Linux x86-64)"
echo -e "  Besu Image: $BESU_IMAGE_TAG"
echo -e "  RLN Prover Image: $RLN_PROVER_TAG"
echo
echo -e "${YELLOW}🚀 Next Steps:${NC}"
echo -e "  1. Run: ${GREEN}make start-env-with-rln${NC}"
echo -e "  2. Test gasless transactions with your custom sequencer"
echo -e "  3. Check logs: ${GREEN}docker logs sequencer${NC}"
echo
echo -e "${BLUE}🔧 Environment Variables:${NC}"
echo -e "  export BESU_IMAGE_TAG=${BESU_IMAGE_TAG}"
echo -e "  export RLN_PROVER_IMAGE_TAG=${RLN_PROVER_TAG}"
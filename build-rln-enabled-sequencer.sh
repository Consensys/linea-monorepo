#!/bin/bash
set -e

echo "🚀 Building RLN-Enabled Sequencer"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LINEA_SEQUENCER_DIR="${SCRIPT_DIR}/besu-plugins/linea-sequencer"
STATUS_RLN_PROVER_DIR="/Users/nadeem/dev/status/linea/status-rln-prover"
CUSTOM_BESU_DIR="${SCRIPT_DIR}/custom-besu-minimal"

echo -e "${BLUE}📁 Working directories:${NC}"
echo -e "  Script: ${SCRIPT_DIR}"
echo -e "  Sequencer: ${LINEA_SEQUENCER_DIR}"
echo -e "  RLN Prover: ${STATUS_RLN_PROVER_DIR}"
echo -e "  Custom Besu: ${CUSTOM_BESU_DIR}"

# Use the exact same image version as the official Linea setup
BESU_PACKAGE_TAG="beta-v2.1-rc16.2-20250521134911-f6cb0f2"
BESU_BASE_IMAGE="consensys/linea-besu-package:${BESU_PACKAGE_TAG}"

echo -e "${BLUE}🦀 Building RLN Bridge Rust Library for Linux...${NC}"
cd "${LINEA_SEQUENCER_DIR}/sequencer/src/main/rust/rln_bridge"

RLN_LIB_FILE="${LINEA_SEQUENCER_DIR}/sequencer/src/main/rust/rln_bridge/target/x86_64-unknown-linux-gnu/release/librln_bridge.so"

if [[ ! -f "$RLN_LIB_FILE" ]]; then
    echo -e "${YELLOW}🐳 Cross-compiling Rust library for Linux x86-64...${NC}"
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
    echo -e "${RED}❌ Error: Linux RLN library not found: $RLN_LIB_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}✅ RLN Bridge library ready: $RLN_LIB_FILE${NC}"

echo -e "${BLUE}☕ Building Custom Sequencer JAR with Dependencies...${NC}"
cd "$SCRIPT_DIR"
# Build with distPlugin to include dependencies not provided by Besu
./gradlew clean :besu-plugins:linea-sequencer:sequencer:distPlugin -x test -x checkSpdxHeader -x spotlessJavaCheck -x spotlessGroovyGradleCheck --no-daemon

# Look for both JAR and ZIP files from distPlugin
SEQUENCER_JAR=$(find "${LINEA_SEQUENCER_DIR}/sequencer/build/libs" -name "linea-sequencer-*.jar" | head -1)
SEQUENCER_DIST=$(find "${LINEA_SEQUENCER_DIR}/sequencer/build/distributions" -name "linea-sequencer-*.zip" | head -1)

echo -e "${YELLOW}  Found JAR: $SEQUENCER_JAR${NC}"
echo -e "${YELLOW}  Found Distribution: $SEQUENCER_DIST${NC}"
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

echo -e "${BLUE}🐳 Building Minimal Custom Besu Image...${NC}"
mkdir -p "$CUSTOM_BESU_DIR"
cd "$CUSTOM_BESU_DIR"

# Copy the files we need to the build directory first
cp "$SEQUENCER_JAR" .
cp "$RLN_LIB_FILE" .

# Extract dependencies to the current build directory (MOVED HERE!)
if [[ -f "$SEQUENCER_DIST" ]]; then
    echo -e "${YELLOW}  📦 Extracting dependencies from distribution ZIP...${NC}"
    unzip -q "$SEQUENCER_DIST" -d extracted-deps/
    # Copy dependency JARs (excluding the main sequencer JAR) to current directory
    find extracted-deps/ -name "*.jar" -not -name "linea-sequencer-*" -exec cp {} . \;
    DEPS_COUNT=$(find extracted-deps/ -name "*.jar" -not -name "linea-sequencer-*" | wc -l)
    echo -e "${YELLOW}  ✅ Extracted $DEPS_COUNT dependency JARs to build directory${NC}"
    rm -rf extracted-deps/
    
    # List what we extracted for debugging
    echo -e "${YELLOW}  Dependencies extracted:${NC}"
    ls -la *.jar | grep -v "$(basename "$SEQUENCER_JAR")" | head -5
else
    echo -e "${YELLOW}  ⚠️  No distribution ZIP found - dependencies may be missing${NC}"
fi

# Extract the entire Besu distribution from official image (like your working script)
echo -e "${YELLOW}📥 Extracting base Besu from official image...${NC}"
docker rm temp-besu-extract 2>/dev/null || true
docker create --name temp-besu-extract "${BESU_BASE_IMAGE}"
docker cp temp-besu-extract:/opt/besu/ ./besu/
docker rm temp-besu-extract

# Verify all required Linea plugins are present
echo -e "${YELLOW}🔍 Verifying Linea plugins...${NC}"
echo -e "  Current plugins in extracted image:"
ls -la ./besu/plugins/

# Replace the sequencer plugin with our custom one
echo -e "${YELLOW}🔄 Installing custom sequencer...${NC}"
rm -f ./besu/plugins/linea-sequencer-*.jar
cp "$SEQUENCER_JAR" ./besu/plugins/
echo -e "  ✅ Installed: $(basename "$SEQUENCER_JAR")"

# Install missing dependency JARs in lib directory
if ls *.jar 1> /dev/null 2>&1; then
    echo -e "${YELLOW}📚 Installing dependency JARs...${NC}"
    for jar in *.jar; do
        if [[ "$jar" != "$(basename "$SEQUENCER_JAR")" ]]; then
            cp "$jar" ./besu/lib/
            echo -e "  ✅ Installed dependency: $jar"
        fi
    done
fi

# Copy RLN native library  
echo -e "${YELLOW}📚 Installing RLN native library...${NC}"
mkdir -p ./besu/lib/native
cp "$RLN_LIB_FILE" ./besu/lib/native/
echo -e "  ✅ Installed: librln_bridge.so"

# Update Besu startup scripts to include plugins in classpath (critical!)
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

# Create Dockerfile like your working script
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

# Remove old extract script - not needed with this approach
rm -f extract-deps.sh

# Build the minimal custom image
TIMESTAMP=$(date +%Y%m%d%H%M%S)
BESU_IMAGE_TAG="linea-besu-minimal-rln:${TIMESTAMP}"

echo -e "${YELLOW}🔨 Building Docker image...${NC}"
docker build --platform linux/amd64 -t "$BESU_IMAGE_TAG" .

echo -e "${GREEN}✅ Minimal custom Besu image built: $BESU_IMAGE_TAG${NC}"

echo -e "${BLUE}🐳 Building RLN Prover Docker image...${NC}"
cd "$STATUS_RLN_PROVER_DIR"
RLN_PROVER_TAG="status-rln-prover:${TIMESTAMP}"
docker build --platform linux/amd64 -t "$RLN_PROVER_TAG" .

echo -e "${GREEN}✅ RLN Prover image built: $RLN_PROVER_TAG${NC}"

echo -e "${BLUE}📝 Updating Docker Compose...${NC}"
COMPOSE_FILE="${SCRIPT_DIR}/docker/compose-spec-l2-services-rln.yml"
if [[ -f "$COMPOSE_FILE" ]]; then
    # Create backup
    cp "$COMPOSE_FILE" "${COMPOSE_FILE}.backup.$(date +%Y%m%d%H%M%S)"
    
    # Update only the sequencer and l2-node-besu images
    sed -i.tmp "s|image: linea-besu.*:.*|image: ${BESU_IMAGE_TAG}|g" "$COMPOSE_FILE"
    sed -i.tmp "s|image: status-rln-prover:.*|image: ${RLN_PROVER_TAG}|g" "$COMPOSE_FILE"
    rm -f "${COMPOSE_FILE}.tmp"
    
    echo -e "${GREEN}✅ Updated Docker Compose with minimal images:${NC}"
    echo -e "  Besu: $BESU_IMAGE_TAG"
    echo -e "  RLN Prover: $RLN_PROVER_TAG"
fi

# Clean up build directory
cd "$SCRIPT_DIR"
rm -rf "$CUSTOM_BESU_DIR"

echo -e "${GREEN}🎉 Minimal Build Complete!${NC}"
echo -e "${BLUE}📋 Built Components:${NC}"
echo -e "  Custom Sequencer JAR: $(basename "$SEQUENCER_JAR")"
echo -e "  RLN Library: librln_bridge.so (Linux x86-64)"
echo -e "  Minimal Besu Image: $BESU_IMAGE_TAG"
echo -e "  RLN Prover Image: $RLN_PROVER_TAG"
echo
echo -e "${YELLOW}🚀 Next Steps:${NC}"
echo -e "  1. Run: ${GREEN}make start-env-with-rln${NC}"
echo -e "  2. Test gasless transactions"
echo -e "  3. Check logs: ${GREEN}docker logs sequencer${NC}"
echo
echo -e "${BLUE}🔧 Environment Variables:${NC}"
echo -e "  export BESU_IMAGE_TAG=${BESU_IMAGE_TAG}"
echo -e "  export RLN_PROVER_IMAGE_TAG=${RLN_PROVER_TAG}"
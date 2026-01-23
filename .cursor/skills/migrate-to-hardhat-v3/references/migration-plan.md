---
name: Hardhat V2 to V3 Migration
overview: A step-by-step guide to migrate an existing Hardhat V2 project to Hardhat V3, covering environment setup, configuration migration, and test migration.
todos:
  - id: phase1-prep
    content: "Phase 1: Preparation - Check Node.js version, clean artifacts, remove V2 dependencies, backup config, make project ESM"
    status: pending
  - id: phase2-install
    content: "Phase 2: Install Hardhat 3, create new config, migrate solidity settings, verify compilation"
    status: pending
  - id: phase3-toolbox
    content: "Phase 3: Install and configure toolbox plugin (Mocha+Ethers or Viem)"
    status: pending
  - id: phase4-tests
    content: "Phase 4: Migrate tests - convert to ESM, update network connections, update Chai matchers, update network helpers"
    status: pending
  - id: phase5-networks
    content: "Phase 5: Migrate network configuration and secrets"
    status: pending
isProject: false
---

# Hardhat V2 to V3 Migration Plan

## Phase 1: Preparation (Before Installing V3)

### 1.1 Check Node.js Version

Hardhat 3 requires **Node.js v22.10.0 or later**.

```bash
node --version
```

If needed, upgrade Node.js before proceeding.

### 1.2 Clean Up Hardhat 2 Artifacts

Run the clean task to remove stale artifacts and caches:

```bash
npx hardhat clean
```

### 1.3 Remove Hardhat 2 Dependencies

Remove these packages from `package.json`:

- `hardhat`
- Any packages starting with `hardhat-`, `@nomicfoundation/`, or `@nomiclabs/`
- `solidity-coverage`

Then reinstall and verify no Hardhat dependencies remain:

```bash
npm install
npm why hardhat
```

Repeat until no Hardhat-related dependencies remain.

### 1.4 Backup Your Old Config

Rename your old config file for reference:

```bash
mv hardhat.config.js hardhat.config.old.js
# or for TypeScript:
mv hardhat.config.ts hardhat.config.old.ts
```

### 1.5 Make Your Project ESM

Add `"type": "module"` to your `package.json`:

```json
{
  "type": "module"
}
```

### 1.6 Update tsconfig.json (TypeScript projects)

If using TypeScript, update `compilerOptions.module` to an ESM-compatible value:

```json
{
  "compilerOptions": {
    "module": "node16"
  }
}
```

---

## Phase 2: Install and Configure Hardhat 3

### 2.1 Install Hardhat 3

```bash
npm install --save-dev hardhat
```

### 2.2 Create New Config File

Create `hardhat.config.ts` with minimal content:

```ts
import { defineConfig } from "hardhat/config";

export default defineConfig({});
```

### 2.3 Verify Installation

```bash
npx hardhat --help
```

### 2.4 Migrate Solidity Config

Copy the `solidity` section from your old config. The format is backwards-compatible:

```ts
import { defineConfig } from "hardhat/config";

export default defineConfig({
  solidity: {
    version: "0.8.24",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
});
```

### 2.5 Compile Contracts

Verify compilation works:

```bash
npx hardhat build
```

---

## Phase 3: Install Toolbox Plugin

### 3.1 Choose Your Toolbox

**Option A: Mocha + Ethers (familiar to V2 users)**

```bash
npm install --save-dev @nomicfoundation/hardhat-toolbox-mocha-ethers
```

**Option B: Node.js test runner + Viem (recommended for new projects)**

```bash
npm install --save-dev @nomicfoundation/hardhat-toolbox-viem
```

### 3.2 Add Plugin to Config (Declarative Config)

Unlike V2, you must explicitly add plugins to the config:

```ts
import { defineConfig } from "hardhat/config";
import hardhatToolboxMochaEthers from "@nomicfoundation/hardhat-toolbox-mocha-ethers";

export default defineConfig({
  plugins: [hardhatToolboxMochaEthers],
  solidity: { /* ... */ },
});
```

---

## Phase 4: Migrate Tests

### 4.1 Convert to ESM Syntax

Change CommonJS imports to ESM:

**Before (V2):**

```js
const { expect } = require("chai");
const hre = require("hardhat");
```

**After (V3):**

```ts
import { expect } from "chai";
import hre from "hardhat";
```

Note: Relative imports must include file extensions (e.g., `./helper.js` not `./helper`).

### 4.2 Update Network Connections

V3 requires explicit network connections. No more global `hre.network.provider`.

**Before (V2):**

```ts
const contract = await hre.ethers.deployContract("MyContract");
```

**After (V3):**

```ts
const { ethers } = await hre.network.connect();
const contract = await ethers.deployContract("MyContract");
```

For shared connections across tests, use top-level await:

```ts
import hre from "hardhat";

const { ethers, provider, networkHelpers } = await hre.network.connect();

describe("MyContract", function () {
  it("should work", async function () {
    const contract = await ethers.deployContract("MyContract");
    // ...
  });
});
```

### 4.3 Update Chai Matchers

Some matchers now require the `ethers` object:

**Before (V2):**

```ts
await expect(tx).to.be.reverted;
await expect(tx).to.changeEtherBalance(addr, amount);
```

**After (V3):**

```ts
await expect(tx).to.be.revert(ethers);
await expect(tx).to.changeEtherBalance(ethers, addr, amount);
```

### 4.4 Update Network Helpers

Network helpers are now part of the connection:

**Before (V2):**

```ts
import { mine } from "@nomicfoundation/hardhat-network-helpers";
await mine(5);
```

**After (V3):**

```ts
const { networkHelpers } = await hre.network.connect();
await networkHelpers.mine(5);
```

### 4.5 Run Tests Incrementally

Migrate one test file at a time and verify:

```bash
npx hardhat test test/some-test.ts
```

---

## Phase 5: Migrate Network Configuration

### 5.1 Update Network Config Format

```ts
import { defineConfig, configVariable } from "hardhat/config";

export default defineConfig({
  networks: {
    sepolia: {
      type: "http",
      chainType: "l1",
      url: configVariable("SEPOLIA_RPC_URL"),
      accounts: [configVariable("SEPOLIA_PRIVATE_KEY")],
    },
  },
});
```

### 5.2 Store Secrets with Keystore Plugin

```bash
npx hardhat keystore set SEPOLIA_RPC_URL
npx hardhat keystore set SEPOLIA_PRIVATE_KEY
```

---

## Key Documentation References

- Main migration guide: [docs/migrate-from-hardhat2/index.mdx](src/content/docs/docs/migrate-from-hardhat2/index.mdx)
- Mocha test migration: [docs/migrate-from-hardhat2/guides/mocha-tests.mdx](src/content/docs/docs/migrate-from-hardhat2/guides/mocha-tests.mdx)
- What's new in V3: [docs/hardhat3/whats-new.mdx](src/content/docs/docs/hardhat3/whats-new.mdx)
- Configuration reference: [docs/reference/configuration.mdx](src/content/docs/docs/reference/configuration.mdx)

---

## Common Gotchas

1. **ESM file extensions**: Relative imports must include `.js` extension
2. **No global network**: Always use `await hre.network.connect()`
3. **Declarative plugins**: Must add plugins to `plugins: []` array, not just import them
4. **Chai matchers**: Some require `ethers` as first argument
5. **`__dirname`/`__filename`**: Use `import.meta.dirname` and `import.meta.filename` instead
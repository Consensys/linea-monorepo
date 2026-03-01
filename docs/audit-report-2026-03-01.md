# Documentation Audit Report - 2026-03-01

## Audit Results: 14 issues found across 170 files

### [HIGH] Terminology conflict: Node.js version requirement
- **Location A:** `AGENTS.md`:53 - "Node.js | >= 22.22.0"
- **Location B:** `CONTRIBUTING.md`:7 - "Node.js | >= 22.22.0 (see `.nvmrc`)"
- **Location C:** `docs/get-started.md`:5 - "Node.js v20 or higher"
- **Location D:** `docs/local-development-guide.md`:9 - "Node.js v20 or higher"
- **Suggested fix:** Update `docs/get-started.md` and `docs/local-development-guide.md` to match the authoritative `AGENTS.md` value of ">= 22.22.0". The `.nvmrc` file is the single source of truth.

### [HIGH] Terminology conflict: pnpm version requirement
- **Location A:** `AGENTS.md`:54 - "pnpm | >= 10.28.0"
- **Location B:** `CONTRIBUTING.md`:8 - "pnpm | >= 10.28.0"
- **Location C:** `docs/get-started.md`:10 - "Pnpm v10.18.3"
- **Location D:** `docs/local-development-guide.md`:14 - "Pnpm v10.18.3"
- **Suggested fix:** Update `docs/get-started.md` and `docs/local-development-guide.md` to ">= 10.28.0" to match `AGENTS.md`.

### [HIGH] Terminology conflict: Solidity pragma version
- **Location A:** `contracts/docs/contract-style-guide.md`:52,77,99 - `pragma solidity >=0.8.19 <=0.8.26;`
- **Location B:** `contracts/AGENTS.md`:50 - "Solidity version: `0.8.33` (exact for contracts, caret `^0.8.33` for interfaces/abstract/libraries)"
- **Suggested fix:** The style guide predates the current Solidity version. Update `contracts/docs/contract-style-guide.md` pragma examples to `0.8.33` (exact) for contracts and `^0.8.33` for interfaces/abstract/libraries, matching `contracts/AGENTS.md` which is the current canonical source.

### [HIGH] Protocol contradiction: unpausing execution path shows wrong function
- **Location A:** `contracts/docs/workflows/administration/unpausing.md`:25 - "calls pauseByType(type)"
- **Location B:** `contracts/docs/workflows/administration/unpausing.md`:37 - Function signature table correctly lists "unPauseByType(uint8)"
- **Suggested fix:** Change line 25 from `pauseByType(type)` to `unPauseByType(type)`. The execution path was likely copy-pasted from `pausing.md` without updating the function name.

### [HIGH] Terminology conflict: Java version for tracer
- **Location A:** `tracer/SETUP.md`:8 - "Install Java 17"
- **Location B:** `tracer/docs/get-started.md`:5 - "Install Java 21"
- **Location C:** `AGENTS.md`:55 - "JDK | 21"
- **Location D:** `docs/tech/components/README.md`:122 - "Java 21 required"
- **Suggested fix:** Update `tracer/SETUP.md` to Java 21. Java 17 is stale; all other docs agree on JDK 21.

### [HIGH] Enumerated value mismatch: L2 Token Bridge proxy address differs across workflow docs
- **Location A:** `contracts/docs/workflows/administration/pausing.md`:77 - `0x353012d04a9A6cF5C941bADC267f82004A8ceB9`
- **Location B:** `contracts/docs/workflows/administration/upgradeContract.md`:106 - `0x353012d04a9A6cF5C541BADC267f82004A8ceB9`
- **Location C:** `contracts/docs/workflows/administration/upgradeAndCallContract.md`:108 - `0x353012d04a9A6cF5C541BADC267f82004A8ceB9`
- **Suggested fix:** Verify the correct address on-chain and update all files to match. The difference is `C941b` vs `C541B` at character position ~25. Cross-check against `contracts/docs/mainnet-address-book.csv`.

### [HIGH] Enumerated value mismatch: Deployer Linea address differs between upgrade docs
- **Location A:** `contracts/docs/workflows/administration/upgradeContract.md`:97 - `0x49ee40140522561744c1C2878c76eE9f28028d33`
- **Location B:** `contracts/docs/workflows/administration/upgradeAndCallContract.md`:99 - `0x49ee40140E522651744c1C2878c76eE9f28028d33`
- **Suggested fix:** Verify the correct address on-chain and reconcile. Key differences: `0522561` vs `0E522651`. Cross-check against `contracts/docs/mainnet-address-book.csv`.

### [MEDIUM] Terminology conflict: Go version
- **Location A:** `AGENTS.md`:57 and `prover/AGENTS.md`:28 - "Go | 1.24.6"
- **Location B:** `docs/tech/components/README.md`:130 - "Go 1.21+ required"
- **Suggested fix:** Update `docs/tech/components/README.md` to "Go 1.24.6" to match the authoritative `AGENTS.md` and the prover's `go.mod`.

### [MEDIUM] Enumerated value mismatch: PauseType count divergence
- **Location A:** `docs/features/pause-and-security.md` - Lists 14 PauseTypes (including NATIVE_YIELD_STAKING, NATIVE_YIELD_UNSTAKING, NATIVE_YIELD_PERMISSIONLESS_ACTIONS, NATIVE_YIELD_REPORTING, STATE_DATA_SUBMISSION, and BLOB_SUBMISSION/CALLDATA_SUBMISSION as deprecated)
- **Location B:** `contracts/docs/workflows/administration/pausing.md`:47-56 - Lists only 8 PauseTypes (GENERAL through COMPLETE_TOKEN_BRIDGING, with PROVING_SYSTEM instead of BLOB_SUBMISSION)
- **Suggested fix:** Update `pausing.md` and `unpausing.md` to include the full current set of PauseTypes. The workflow docs appear to predate the native yield and state data submission additions.

### [MEDIUM] Terminology conflict: E2E test command inconsistency
- **Location A:** `AGENTS.md`:109 - `pnpm -F e2e run test:local`
- **Location B:** `docs/get-started.md`:25 - `pnpm run test:e2e:local`
- **Location C:** `docs/tech/README.md`:123 - `cd e2e && pnpm run test:e2e:local`
- **Location D:** `e2e/README.md`:58 - `pnpm run -F e2e test:local`
- **Suggested fix:** Standardize on one canonical form. `AGENTS.md` uses `pnpm -F e2e run test:local` while `get-started.md` uses `pnpm run test:e2e:local` (potentially a different script name). Verify which script name is correct in `e2e/package.json` and update all docs to match.

### [MEDIUM] Protocol contradiction: package manager command in token bridge docs
- **Location A:** `contracts/docs/linea-token-bridge.md`:78 - `npm run test`
- **Location B:** `contracts/docs/linea-token-bridge.md`:93 - `npm run coverage`
- **Location C:** `AGENTS.md`:69 - "Only `pnpm` is allowed (enforced by `preinstall`)"
- **Suggested fix:** Replace all `npm run` and `npx hardhat` commands in `contracts/docs/linea-token-bridge.md` with their `pnpm` equivalents. npm is blocked by the preinstall script.

### [LOW] Stale cross-reference: Solidity version in components README
- **Location:** `docs/tech/components/README.md`:146 - "Solidity 0.8.x"
- **Problem:** Vague version reference; actual version is `0.8.33` per `contracts/AGENTS.md`:50
- **Suggested fix:** Update to "Solidity 0.8.33" for consistency.

### [LOW] Stale cross-reference: TypeScript version requirements vary in precision
- **Location A:** `docs/tech/components/README.md`:137 - "Node.js 22+"
- **Location B:** `docs/tech/components/README.md`:138 - "pnpm 10+"
- **Problem:** These are less precise than the canonical values in `AGENTS.md` (Node.js >= 22.22.0, pnpm >= 10.28.0) but not technically wrong. `[UNCERTAIN]` May be stylistic rather than substantive.
- **Suggested fix:** Consider updating to match exact minimum versions from `AGENTS.md` for consistency.

### [LOW] Stale cross-reference: PauseType column header
- **Location A:** `contracts/docs/workflows/administration/pausing.md`:47 - Column header says "Address" for PauseType name column
- **Location B:** `contracts/docs/workflows/administration/unpausing.md`:47 - Same issue
- **Problem:** The column header "Address" should be "Name" or "Pause Type" - PauseType names are not addresses.
- **Suggested fix:** Change column header from "Address" to "Name" in both files.

---

### [OK] Permission inconsistencies
- Checked role definitions across `contracts/docs/workflows/administration/roleManagement.md`, `contracts/docs/security-council-charter.md`, `docs/features/pause-and-security.md`, `docs/features/yield-management.md`, and `docs/features/token-bridge.md`. Role names and permissions are consistent across all references.

### [OK] Lifecycle mismatches
- Checked messaging workflows across `contracts/docs/workflows/messaging/`, `docs/features/messaging.md`, `docs/features/postman.md`, `docs/tech/architecture/OVERVIEW.md`. Message lifecycle (SENT -> ANCHORED -> CLAIMED) is consistently described. The L1->L2 and L2->L1 asymmetry (rolling hash vs Merkle tree) is correctly documented in all files.

### [OK] Stale cross-references (internal links)
- Checked 45+ internal Markdown links and file path references across all docs. All referenced files exist: workflow diagrams (`contracts/docs/workflows/diagrams/*.png`), security council charter, mainnet address book CSV, architecture overview, feature docs. No broken internal links found beyond the issues already noted above.

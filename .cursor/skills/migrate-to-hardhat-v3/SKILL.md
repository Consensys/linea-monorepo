---
name: migrate-to-hardhat-v3
description: Guide for migrating Hardhat V2 projects to Hardhat V3. Use when asked to upgrade Hardhat, migrate to V3, or fix V2/V3 compatibility issues.
---

# Hardhat V2 to V3 Migration

This skill provides guidance for migrating existing Hardhat V2 projects to Hardhat V3.

## When to Use

- User asks to upgrade/migrate Hardhat to V3
- User encounters Hardhat V2/V3 compatibility issues
- User wants to modernize their Hardhat setup
- User asks about ESM migration for Hardhat projects
- User has issues with network connections, plugins, or test runners after a Hardhat upgrade

## Primary Reference

**Always start with the migration plan:**
[references/migration-plan.md](references/migration-plan.md)

This contains a phased approach:
1. **Phase 1**: Preparation (Node.js version, cleanup, ESM setup)
2. **Phase 2**: Install and configure Hardhat 3
3. **Phase 3**: Install toolbox plugin (Mocha+Ethers or Viem)
4. **Phase 4**: Migrate tests (ESM syntax, network connections, Chai matchers)
5. **Phase 5**: Migrate network configuration and secrets

## Documentation Reference

When you need more details or encounter issues not covered in the migration plan, explore the official Hardhat V3 docs:
[references/hardhat-v3-docs/](references/hardhat-v3-docs/)

Key documentation locations:
- **Migration overview**: `migrate-from-hardhat2/index.mdx`
- **Mocha test migration**: `migrate-from-hardhat2/guides/mocha-tests.mdx`
- **What's new in V3**: `hardhat3/whats-new.mdx`
- **Configuration reference**: `reference/configuration.mdx`
- **Network manager**: `reference/network-manager.mdx`
- **Guides**: `guides/` (testing, deploying, verifying, etc.)
- **Explanations**: `explanations/` (deep dives on concepts)

## Instructions

1. **Assess the current state**: Check the existing `hardhat.config.js/ts`, `package.json`, and test files to understand what needs to be migrated.

2. **Follow the phased approach**: Work through the migration plan phases sequentially. Do not skip phases as each builds on the previous.

3. **Migrate incrementally**: Migrate one test file at a time and verify it passes before moving to the next.

4. **Key breaking changes to remember**:
   - ESM is required (`"type": "module"` in package.json)
   - Plugins must be added to `plugins: []` array (not just imported)
   - Network connections are explicit: `await hre.network.connect()`
   - Some Chai matchers require `ethers` as first argument
   - Relative imports must include file extensions (`.js`)
   - Use `import.meta.dirname` instead of `__dirname`

5. **If stuck**: Search the `hardhat-v3-docs/` directory for relevant documentation. Use glob patterns to find specific topics.

6. **Ask for clarification**: If the user's project has unusual configurations or custom plugins, use the ask questions tool to clarify requirements before proceeding.

## Common Issues and Solutions

| Issue | Solution |
|-------|----------|
| `require is not defined` | Project needs ESM: add `"type": "module"` to package.json |
| `hre.ethers is undefined` | Use `const { ethers } = await hre.network.connect()` |
| Plugin not working | Add plugin to `plugins: []` array in config |
| `Cannot find module` with relative import | Add `.js` extension to relative imports |
| `__dirname is not defined` | Use `import.meta.dirname` instead |
| Tests timeout or hang | Check network connection is awaited properly |

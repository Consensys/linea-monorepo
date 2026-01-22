# LineaRollup V7 Upgrade Verification - Sepolia

## Transaction Details

- **TX Hash**: [0xc3d0a93b353c78cfbe1d9e450c8683087fa8835723a40690a4e694a269f51e9c](https://sepolia.etherscan.io/tx/0xc3d0a93b353c78cfbe1d9e450c8683087fa8835723a40690a4e694a269f51e9c/advanced)
- **Date**: Jan 20, 2026 11:37:24 AM UTC
- **Block**: 8,008,468 (1,008,468,616,582 confirmations at time of analysis)

## Contract Addresses

| Contract | Address |
|----------|---------|
| Proxy | `0x2A5CDCfc38856e2590E9Bd32F54Fa348e5De5f48` |
| New Implementation | `0xCaAa421FfCF701bEFd676a2F5d0A161CCFA5a07E` |
| Safe (Multisig) | `0xe6Ec44e651B6d961c15f1A8df9eA7DFaDb986eA1` |
| TimelockController | `0xB8817B2d36B368Fa9Ef4D9c5fF9658858037b024` |
| ProxyAdmin | `0x10b7ef80D4bA8df6b4Daed7B7638cd88C6d52F02` |
| YieldManager | `0xafeB487DD3E3Cb0342e8CF0215987FfDc9b72c9b` |

## Transaction Flow

```
Safe.execTransaction()
  └─> TimelockController.execute()
        └─> ProxyAdmin.upgradeAndCall()
              ├─> Proxy.upgradeTo(implementation)
              └─> LineaRollup.reinitializeLineaRollupV7(...)
```

## Decoded Reinitializer Call

### Function Signature

```solidity
function reinitializeLineaRollupV7(
    RoleAddress[] calldata _roleAddresses,
    PauseTypeRole[] calldata _pauseTypeRoles,
    PauseTypeRole[] calldata _unpauseTypeRoles,
    address _yieldManager
) external reinitializer(7)
```

### Parameters

#### _roleAddresses (5 roles granted)

| Role Name | Role Hash | Recipient |
|-----------|-----------|-----------|
| `OPERATOR_ROLE` | `0x76ef52a5344b10ed112c1d48c7c06f51e919518ea6fb005f9b25b359b955e3be` | Safe (`0xe6Ec...eA1`) |
| `VERIFIER_SETTER_ROLE` | `0x220bd22ef7c53d75fe3eac0a09e90815a0c5ba4f9e8da8b039542cd3db347258` | Safe (`0xe6Ec...eA1`) |
| `PAUSE_NATIVE_YIELD_STAKING_ROLE` | `0xcc10d6eec3c757d645e27b3f3001a3ba52f692da0bce25fabf58c6ecaf376450` | Safe (`0xe6Ec...eA1`) |
| `UNPAUSE_NATIVE_YIELD_STAKING_ROLE` | `0x4b4665d8754e6ea0608430ef3e91c1b45c72aafe8800e289cd35f38d85361858` | Safe (`0xe6Ec...eA1`) |
| `VERIFIER_SETTER_ROLE` | `0x220bd22ef7c53d75fe3eac0a09e90815a0c5ba4f9e8da8b039542cd3db347258` | Dead address (`0x...dEaD`) |

#### _pauseTypeRoles (1 mapping)

| PauseType | Role Hash |
|-----------|-----------|
| 9 (NATIVE_YIELD_STAKING) | `PAUSE_NATIVE_YIELD_STAKING_ROLE` (`0xcc10d6ee...`) |

#### _unpauseTypeRoles (1 mapping)

| UnPauseType | Role Hash |
|-------------|-----------|
| 9 (NATIVE_YIELD_STAKING) | `UNPAUSE_NATIVE_YIELD_STAKING_ROLE` (`0x4b4665d8...`) |

#### _yieldManager

`0xafeB487DD3E3Cb0342e8CF0215987FfDc9b72c9b`

## Emitted Events (Log Order)

1. **Upgraded** - Implementation changed to `0xCaAa421F...`
2. **RoleGranted** (x5) - Roles assigned to Safe and dead address
3. **PauseTypeRoleSet** - PauseType 9 mapped to role
4. **UnPauseTypeRoleSet** - UnPauseType 9 mapped to role
5. **YieldManagerChanged** - `0x0` → `0xafeB487D...`
6. **LineaRollupVersionChanged** - "6.0" → "7.0"
7. **Initialized** - version = 7

## Verification Commands

```bash
# Set RPC URL
export ETHEREUM_SEPOLIA_RPC_URL="https://sepolia.infura.io/v3/YOUR_KEY"

# Run verification
cd contracts
npx ts-node scripts/operational/contract-integrity-verifier/index.ts \
  -c scripts/operational/contract-integrity-verifier/configs/sepolia-linea-rollup-v7.json \
  -v
```

## Expected Verification Results

- **Bytecode**: PASS (implementation matches artifact)
- **ABI**: PASS (all selectors present)
- **State Verification**:
  - `contractVersion()` = "7.0" ✓
  - `yieldManager()` = `0xafeB487D...` ✓
  - All `hasRole()` checks = true ✓

---
title: Comprehensive Testing Strategy for EIP-7702 (Set EOA Account Code) Support
labels: testing, eip-7702, prague, priority:high
assignees: ''
---

🔐 🔐 `Should this be a security advisory instead?` 🔐 🔐

## Description

Linea currently has basic EIP-7702 transaction **denial** tests but lacks comprehensive testing for when the feature is **enabled**. We need a full testing strategy covering transaction lifecycle, zkEVM constraints, proof generation, and security boundaries to safely support [EIP-7702: Set EOA account code](https://eips.ethereum.org/EIPS/eip-7702).

**Current state:**
- ✅ Transaction denial works (`EIP7702TransactionDenialTest.kt`)
- ✅ CLI flag `--plugin-linea-delegate-code-tx-enabled` implemented
- ✅ Test infrastructure exists (contracts, scripts)
- ❌ No end-to-end tests with delegation enabled
- ❌ Limited zkEVM constraint validation
- ❌ No RPC endpoint coverage for delegated accounts
- ❌ Missing security/edge case tests

## Motivation

EIP-7702 allows EOAs to temporarily delegate execution to contract code, fundamentally changing Ethereum's execution model. For Linea's zkEVM, this introduces unique challenges:

1. **Constraint System**: Every delegation operation must be provable
2. **Security**: Code delegation creates new attack vectors  
3. **State Management**: Delegation affects zkEVM state representation
4. **Performance**: Impact on proof generation must be acceptable
5. **L1/L2 Bridge**: Delegated accounts may interact with cross-chain messages

Without comprehensive testing, enabling EIP-7702 risks security vulnerabilities, proof generation failures, or incompatibility with Ethereum.

## Tasks

### 1. Core Transaction Lifecycle
- [ ] Authorization list validation (signatures, nonces, chain IDs)
- [ ] Transaction pool acceptance when feature enabled
- [ ] Block production with Type 4 transactions
- [ ] Block import and P2P gossip
- [ ] Transaction ordering and gas accounting

### 2. Execution & State Management  
- [ ] Delegation code marker (`0xef0100`) verification
- [ ] Delegated code execution (DELEGATECALL behavior)
- [ ] State transitions (before/during/after delegation)
- [ ] Storage access and balance changes
- [ ] CREATE/CREATE2 from delegated EOA

### 3. zkEVM Constraint System
- [ ] ROMLEX module handling of delegation code
- [ ] HUB module state transitions for Type 4 txs
- [ ] RLPTXN module authorization list parsing  
- [ ] Proof generation for delegation operations
- [ ] Performance benchmarks (proof time overhead)

### 4. RPC Endpoints
- [ ] `eth_sendRawTransaction` with Type 4 transactions
- [ ] `eth_getCode` for delegated EOA (returns `0xef0100...`)
- [ ] `eth_call` and `eth_estimateGas` with delegations
- [ ] `linea_estimateGas` gas/profitability calculations

### 5. Security & Edge Cases
- [ ] Invalid authorizations (wrong sig, chain ID, nonce)
- [ ] Empty/maximum authorization list sizes
- [ ] Reentrancy in delegated code
- [ ] Out-of-gas scenarios
- [ ] Authorization replay protection
- [ ] Transaction-scoped delegation (not persistent)

### 6. Integration Testing  
- [ ] L1→L2 messages with delegated recipients
- [ ] Token bridge with delegated accounts
- [ ] End-to-end: tx → block → proof → verification
- [ ] Multi-node synchronization

### 7. Performance & Stress Testing
- [ ] High volume of Type 4 transactions
- [ ] Proof generation at scale
- [ ] Gas cost benchmarks (predicted vs actual)

### 8. Documentation & Tooling
- [ ] Update developer docs with EIP-7702 examples
- [ ] Document CLI flags and usage
- [ ] Verify ethers.js/wallet compatibility

## Acceptance Criteria

### Must Have ✅
- [ ] All transaction lifecycle tests pass (pool → block → import)
- [ ] zkEVM constraint system handles delegation correctly
- [ ] Proof generation succeeds for all delegation scenarios  
- [ ] RPC endpoints return correct Type 4 transaction data
- [ ] Security tests show no unauthorized state changes
- [ ] Transaction denial works when `--plugin-linea-delegate-code-tx-enabled=false`
- [ ] Code coverage ≥ 80% for EIP-7702 code paths

### Should Have 🎯
- [ ] Performance overhead < 15% vs regular transactions
- [ ] Stress tests demonstrate stability
- [ ] Integration tests validate cross-component behavior
- [ ] Documentation updated and reviewed

## Risks

### Technical Risks 🔴
- [ ] **Constraint complexity**: Delegation may require difficult-to-verify constraints
  - *Mitigation*: Incremental testing, start simple
- [ ] **Proof performance**: Type 4 txs may slow proof generation significantly
  - *Mitigation*: Early benchmarking, constraint optimization
- [ ] **State edge cases**: New transition paths may have gaps
  - *Mitigation*: Comprehensive fuzzing

### Security Risks 🔒  
- [ ] **Authorization bypass**: Improper validation → unauthorized delegations
  - *Mitigation*: Multi-layer validation (plugin + constraints)
- [ ] **Reentrancy**: Delegated code may introduce new vectors
  - *Mitigation*: Specific reentrancy tests, security audit
- [ ] **Replay attacks**: Authorization signatures replayable?
  - *Mitigation*: Verify chain ID and nonce validation

### Operational Risks ⚠️
- [ ] **Network coordination**: Prague upgrade requires synchronized rollout
  - *Mitigation*: Testnet validation, clear communication
- [ ] **Backward compatibility**: Pre/post Prague block handling
  - *Mitigation*: Compatibility testing

## Testing Phases

**Phase 1 (Weeks 1-2):** Core lifecycle + transaction pool  
**Phase 2 (Weeks 3-4):** zkEVM integration + RPC endpoints  
**Phase 3 (Weeks 5-6):** Security + edge cases (CRITICAL)  
**Phase 4 (Weeks 7-8):** Performance + integration  
**Phase 5 (Week 9):** Documentation + CI/CD  

## Test Scenarios

1. **Simple self-delegation**: EOA delegates to event-logging contract
2. **Cross-account delegation**: A delegates to contract B, initiated by C  
3. **Multiple authorizations**: 3+ entries in single transaction
4. **Failed delegation**: Invalid sig, wrong chain ID, bad nonce
5. **Complex execution**: Delegation to contract with storage/events/calls
6. **Gas edge cases**: Exactly hit limit, OOG scenarios

## Remember to
- [ ] Add labels: `testing`, `eip-7702`, `prague`, `priority:high`
- [ ] Add team labels: `team:zktracer`, `team:sequencer`  
- [ ] Link to [EIP-7702 spec](https://eips.ethereum.org/EIPS/eip-7702)
- [ ] Update metrics/alerts for Type 4 transaction monitoring
- [ ] Consider security audit before production
- [ ] Update runbook with troubleshooting steps

## References

**Implementation:**
- `besu-plugins/linea-sequencer/acceptance-tests/.../EIP7702TransactionDenialTest.kt`
- `contracts/src/_testing/mocks/base/TestEIP7702Delegation.sol`  
- `contracts/scripts/testEIP7702/sendType4Tx.ts`
- `besu-plugins/linea-sequencer/docs/plugins.md`

**Specification:**
- [EIP-7702: Set EOA account code](https://eips.ethereum.org/EIPS/eip-7702)
- [Prague Network Upgrade](https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/prague.md)

**Components:**
- ROMLEX module (delegation code)
- HUB module (transaction processing)
- RLPTXN module (authorization parsing)
- LineaTransactionValidatorPlugin
- LineaTransactionSelectorPlugin

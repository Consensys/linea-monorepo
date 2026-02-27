# Comprehensive Testing Strategy for EIP-7702 (Set EOA Account Code) Support

## Description

As Linea implements support for [EIP-7702: Set EOA account code](https://eips.ethereum.org/EIPS/eip-7702), we need a comprehensive testing strategy to ensure the feature works correctly across all components of the zkEVM stack while maintaining security and compatibility with the constraint system.

EIP-7702 introduces a new transaction type (Type 4) that allows Externally Owned Accounts (EOAs) to temporarily delegate their code execution to a contract address. This is a significant change to Ethereum's execution model and requires thorough testing across:

- Transaction validation and pool management
- Block production and import
- zkEVM constraint system
- Arithmetization and proving
- RPC endpoints
- Cross-layer message passing
- Security boundaries

## Current Implementation Status

### ✅ Completed
- Basic transaction denial tests (`EIP7702TransactionDenialTest.kt`)
- Transaction validator plugin with CLI option `--plugin-linea-delegate-code-tx-enabled`
- Test contract (`TestEIP7702Delegation.sol`)
- Type 4 transaction sending scripts (`sendType4Tx.ts`, `deploy_TestEIP7702Delegation.ts`)
- Transaction type recognition in SDK and prover components

### 🚧 Gaps Identified
1. **No end-to-end acceptance tests** when EIP-7702 is enabled
2. **Limited zkEVM constraint validation** for delegation opcodes
3. **No stress/fuzzing tests** for authorization lists
4. **No cross-chain bridge interaction tests**
5. **Limited RPC method coverage** for delegated accounts
6. **No performance benchmarking** for proof generation with delegation
7. **Insufficient edge case coverage** (nonce handling, reentrancy, gas limits)

## Motivation

EIP-7702 support is critical for Linea to maintain compatibility with the latest Ethereum features. However, as a zkEVM, we face unique challenges:

1. **Constraint System Complexity**: Every EIP-7702 operation must be provable within our constraint system
2. **Security Implications**: Temporary code delegation introduces new attack vectors
3. **State Management**: Delegation affects account state representation in the zkEVM
4. **Gas Accounting**: Delegation has specific gas costs that must be accurately tracked
5. **Rollup Implications**: L1/L2 message passing may interact with delegated accounts

A robust testing strategy ensures we can safely enable EIP-7702 without compromising Linea's security guarantees or proof generation capabilities.

## Tasks

### 1. Core Transaction Lifecycle Testing
- [ ] **Transaction Pool Validation**
  - [ ] Test authorization list parsing and validation
  - [ ] Test authorization signature verification
  - [ ] Test nonce handling (especially `authNonce = currentNonce + 1` for self-delegation)
  - [ ] Test chain ID validation in authorizations
  - [ ] Test invalid authorization rejection (wrong signature, chain ID mismatch)
  - [ ] Test transaction pool acceptance when `--plugin-linea-delegate-code-tx-enabled=true`
  - [ ] Test gas limit validation for Type 4 transactions

- [ ] **Block Production**
  - [ ] Test Type 4 transaction inclusion in blocks
  - [ ] Test multiple Type 4 transactions in same block
  - [ ] Test interaction between Type 4 and other transaction types
  - [ ] Test block gas limit accounting with delegations
  - [ ] Test transaction ordering with delegations

- [ ] **Block Import**
  - [ ] Test importing blocks with Type 4 transactions
  - [ ] Test block validation with multiple authorizations
  - [ ] Test block rejection for invalid delegations
  - [ ] Test P2P gossip of blocks containing Type 4 transactions

### 2. Execution & State Management
- [ ] **Delegation Execution**
  - [ ] Test successful delegation to valid contract
  - [ ] Test delegation code marker (`0xef0100`) in account code
  - [ ] Test execution of delegated contract code
  - [ ] Test DELEGATECALL behavior from delegated EOA
  - [ ] Test SELFDESTRUCT in delegated context
  - [ ] Test CREATE/CREATE2 from delegated EOA
  - [ ] Test storage access from delegated code
  - [ ] Test balance changes during delegation

- [ ] **State Transitions**
  - [ ] Test account state before/after delegation
  - [ ] Test delegation persistence within transaction
  - [ ] Test delegation cleanup after transaction completion
  - [ ] Test multiple delegations to same address
  - [ ] Test delegation to non-existent contract address
  - [ ] Test delegation to EOA address (should fail or be handled gracefully)

### 3. zkEVM Constraint System
- [ ] **Arithmetization Coverage**
  - [ ] Test ROMLEX module handling of delegation code (`0xef0100`)
  - [ ] Test HUB module state transitions for Type 4 transactions
  - [ ] Test RLPTXN module parsing of authorization lists
  - [ ] Test transaction metadata with authorization data
  - [ ] Test constraint generation for delegation operations

- [ ] **Proof Generation**
  - [ ] Test proof generation for blocks with Type 4 transactions
  - [ ] Benchmark proof generation time vs. regular transactions
  - [ ] Test proof verification for delegated execution
  - [ ] Test prover with maximum authorization list size
  - [ ] Test prover stability with various delegation patterns

### 4. RPC Endpoints
- [ ] **Standard Ethereum RPC**
  - [ ] Test `eth_sendRawTransaction` with Type 4 transactions
  - [ ] Test `eth_getTransactionByHash` response format
  - [ ] Test `eth_getTransactionReceipt` with delegation data
  - [ ] Test `eth_getCode` for delegated EOA (should return `0xef0100...`)
  - [ ] Test `eth_call` with delegated account as sender
  - [ ] Test `eth_estimateGas` for Type 4 transactions

- [ ] **Linea-Specific RPC**
  - [ ] Test `linea_estimateGas` with Type 4 transactions
  - [ ] Verify gas estimation includes authorization processing costs
  - [ ] Test profitability calculations for delegated transactions
  - [ ] Test module line count limits with delegation operations

### 5. Security & Edge Cases
- [ ] **Authorization Edge Cases**
  - [ ] Test empty authorization list
  - [ ] Test maximum authorization list size
  - [ ] Test duplicate authorizations in same transaction
  - [ ] Test authorization with zero address
  - [ ] Test authorization with invalid nonce (too low, too high)
  - [ ] Test authorization reuse attempts

- [ ] **Reentrancy & Complex Flows**
  - [ ] Test reentrancy in delegated code
  - [ ] Test nested delegations (if possible)
  - [ ] Test delegation + DELEGATECALL chains
  - [ ] Test delegation with value transfer
  - [ ] Test delegation interaction with precompiles

- [ ] **Gas & Limits**
  - [ ] Test gas consumption for authorization processing
  - [ ] Test out-of-gas scenarios during delegation
  - [ ] Test max calldata size with authorization lists
  - [ ] Test transaction with excessive authorization list

- [ ] **Security Boundaries**
  - [ ] Test that delegation is transaction-scoped (not persistent)
  - [ ] Test authorization signature verification edge cases
  - [ ] Test replay protection across different chains
  - [ ] Verify no unauthorized state changes

### 6. Integration Testing
- [ ] **Bridge Interactions**
  - [ ] Test L1→L2 message with delegated recipient
  - [ ] Test L2→L1 message from delegated EOA
  - [ ] Test token bridge operations with delegated accounts

- [ ] **Multi-Component Testing**
  - [ ] End-to-end test: Type 4 tx → block production → proof generation → verification
  - [ ] Test sequencer behavior with delegations enabled
  - [ ] Test validator/RPC node synchronization with Type 4 transactions
  - [ ] Test transaction rejection flow across all components

### 7. Performance & Stress Testing
- [ ] **Load Testing**
  - [ ] Stress test with high volume of Type 4 transactions
  - [ ] Test block production under delegation load
  - [ ] Test proof generation performance at scale

- [ ] **Benchmarking**
  - [ ] Measure delegation overhead vs. regular transactions
  - [ ] Compare gas costs: predicted vs. actual
  - [ ] Measure constraint system impact

### 8. Developer Experience
- [ ] **Documentation**
  - [ ] Update developer documentation with EIP-7702 usage examples
  - [ ] Document CLI flags for delegation control
  - [ ] Add code comments explaining delegation flow
  - [ ] Update SDK documentation for Type 4 transaction creation

- [ ] **Tooling**
  - [ ] Verify ethers.js integration works correctly
  - [ ] Test with popular wallet libraries
  - [ ] Validate transaction encoding/decoding tools

### 9. Compatibility Testing
- [ ] **Ethereum Compatibility**
  - [ ] Cross-reference tests with Ethereum reference tests
  - [ ] Verify behavior matches latest EIP-7702 specification
  - [ ] Test with official Prague testnet scenarios

- [ ] **Backward Compatibility**
  - [ ] Ensure nodes without EIP-7702 can sync pre-Prague blocks
  - [ ] Test graceful degradation when feature is disabled

## Acceptance Criteria

### Must Have ✅
- [ ] All transaction lifecycle tests pass (pool → block → import)
- [ ] zkEVM constraint system correctly handles delegation opcodes
- [ ] Proof generation succeeds for all delegation scenarios
- [ ] RPC endpoints return correct data for Type 4 transactions
- [ ] Security tests demonstrate no unauthorized state changes
- [ ] Transaction denial mechanism works correctly when feature disabled
- [ ] End-to-end tests cover happy path and common error cases
- [ ] Code coverage ≥ 80% for new EIP-7702 related code

### Should Have 🎯
- [ ] Performance tests show acceptable overhead (< 15% vs regular transactions)
- [ ] Stress tests demonstrate stability under load
- [ ] Integration tests validate cross-component behavior
- [ ] Fuzzing tests identify no critical vulnerabilities
- [ ] Documentation updated and reviewed

### Nice to Have 💡
- [ ] Automated regression tests in CI/CD pipeline
- [ ] Comparison benchmarks with other zkEVMs
- [ ] Advanced delegation patterns tested (e.g., delegation chains)
- [ ] Interactive testing dashboard

## Risks

### Technical Risks 🔴
- [ ] **Constraint System Complexity**: Delegation may require new constraints that are complex to implement and verify
  - *Mitigation*: Start with simple delegation scenarios, gradually increase complexity
  
- [ ] **Proof Generation Performance**: Type 4 transactions may significantly impact proof generation time
  - *Mitigation*: Benchmark early, optimize constraint system if needed
  
- [ ] **State Transition Edge Cases**: Delegation introduces new state transition paths that may not be fully covered
  - *Mitigation*: Comprehensive fuzzing and edge case testing

### Security Risks 🔒
- [ ] **Authorization Bypass**: Improper validation could allow unauthorized delegations
  - *Mitigation*: Multi-layer validation in validator plugin + constraint system
  
- [ ] **Reentrancy Vulnerabilities**: Delegated code may introduce new reentrancy vectors
  - *Mitigation*: Specific reentrancy tests, security audit

- [ ] **Replay Attacks**: Authorization signatures might be replayable across chains/networks
  - *Mitigation*: Verify chain ID and nonce validation

### Operational Risks ⚠️
- [ ] **Network Split**: Enabling EIP-7702 requires coordinated upgrade
  - *Mitigation*: Clear communication, testnet validation before mainnet
  
- [ ] **Backward Compatibility**: Nodes must handle pre/post Prague blocks
  - *Mitigation*: Thorough compatibility testing

## Testing Strategy & Priorities

### Phase 1: Foundation (Week 1-2)
**Priority: HIGH**
- Implement core transaction lifecycle tests
- Validate transaction pool behavior
- Test basic delegation execution
- Verify transaction denial when disabled

### Phase 2: zkEVM Integration (Week 3-4)
**Priority: HIGH**
- Constraint system validation
- Proof generation tests
- Arithmetization coverage
- RPC endpoint testing

### Phase 3: Security & Edge Cases (Week 5-6)
**Priority: CRITICAL**
- Authorization edge cases
- Reentrancy testing
- Gas limit validation
- Security boundary verification

### Phase 4: Performance & Integration (Week 7-8)
**Priority: MEDIUM**
- Performance benchmarking
- Stress testing
- Bridge integration tests
- Cross-component validation

### Phase 5: Documentation & Polish (Week 9)
**Priority: MEDIUM**
- Developer documentation
- Code review and cleanup
- CI/CD integration
- Final validation

## Test Data & Scenarios

### Representative Test Vectors
1. **Simple Self-Delegation**: EOA delegates to simple contract with event logging
2. **Cross-Account Delegation**: Account A delegates to contract B, initiated by account C
3. **Multiple Authorizations**: Single transaction with 3+ authorization entries
4. **Failed Delegation**: Invalid signature, wrong chain ID, incorrect nonce
5. **Complex Execution**: Delegation to contract that performs multiple operations (storage, events, calls)
6. **Gas Edge Cases**: Delegation that exactly hits gas limit, OOG scenarios

### Test Environments
- Local development network (Hardhat/Ganache with Prague fork)
- Linea devnet with EIP-7702 enabled
- Public testnet (Sepolia/Goerli with Prague)
- Mainnet shadow testing (replay blocks with delegation support)

## Remember to
- [ ] Add the `testing` label
- [ ] Add the `eip-7702` label  
- [ ] Add the `prague` label
- [ ] Add `priority:high` label
- [ ] Add `team:zktracer` and `team:sequencer` labels
- [ ] Link to EIP-7702 specification: https://eips.ethereum.org/EIPS/eip-7702
- [ ] Update metrics/alerts for Type 4 transaction monitoring
- [ ] Consider security audit for delegation implementation
- [ ] Update runbook with EIP-7702 troubleshooting steps

## References

### Specification & Standards
- [EIP-7702: Set EOA account code](https://eips.ethereum.org/EIPS/eip-7702)
- [Prague Network Upgrade](https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/prague.md)
- Besu EIP-7702 implementation
- Linea constraint system specification

### Existing Implementation
- `besu-plugins/linea-sequencer/acceptance-tests/src/test/kotlin/linea/plugin/acc/test/EIP7702TransactionDenialTest.kt`
- `contracts/src/_testing/mocks/base/TestEIP7702Delegation.sol`
- `contracts/scripts/testEIP7702/sendType4Tx.ts`
- `besu-plugins/linea-sequencer/docs/plugins.md` (LineaTransactionValidatorPlugin)
- `sdk/sdk-core/src/types/transaction.ts` (Type 4 transaction support)

### Related Components
- ROMLEX module (delegation code handling)
- HUB module (transaction processing)
- RLPTXN module (authorization list parsing)
- Transaction validator plugin
- Transaction selector plugin

---

## Additional Context

This testing strategy is designed to be:
- **Comprehensive**: Covers all aspects of EIP-7702 from transaction submission to proof verification
- **Incremental**: Can be implemented in phases based on priority
- **Risk-Aware**: Identifies and mitigates key technical and security risks
- **Measurable**: Clear acceptance criteria and metrics

The phased approach allows us to validate core functionality early while progressively adding complexity and edge case coverage. Security testing is prioritized throughout all phases given the sensitive nature of code delegation.

### Success Metrics
- **Correctness**: All tests pass, no bugs found in production
- **Coverage**: ≥80% code coverage for EIP-7702 related code
- **Performance**: Delegation overhead < 15% vs. regular transactions
- **Security**: Zero critical vulnerabilities identified post-deployment
- **Compatibility**: 100% compliance with EIP-7702 specification

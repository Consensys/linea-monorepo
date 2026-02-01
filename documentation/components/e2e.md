# E2E Tests

> End-to-end test suite validating the complete Linea stack.

## Overview

The E2E tests validate:
- Cross-chain bridging (L1 ↔ L2)
- Message passing and claiming
- Rollup submission and finalization
- Coordinator restart recovery
- Node fleet consistency
- Liveness detection

## Test Coverage

```
┌────────────────────────────────────────────────────────────────────────┐
│                         E2E TEST COVERAGE                              │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       Bridging Tests                             │  │
│  │                                                                  │  │
│  │  bridge-tokens.spec.ts                                           │  │
│  │  ├── L1 → L2: mint, approve, bridge, anchor, claim               │  │
│  │  └── L2 → L1: mint, approve, bridge, finalize, claim             │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                      Messaging Tests                             │  │
│  │                                                                  │  │
│  │  messaging.spec.ts                                               │  │
│  │  ├── L1 → L2: with/without fees, with/without calldata           │  │
│  │  ├── L2 → L1: with/without fees, with/without calldata           │  │
│  │  └── Postman sponsoring                                          │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       Rollup Tests                               │  │
│  │                                                                  │  │
│  │  submission-finalization.spec.ts                                 │  │
│  │  ├── L2 anchoring via RollingHashUpdated                         │  │
│  │  ├── L1 data submission (DataSubmittedV3)                        │  │
│  │  ├── Finalization with proofs (DataFinalizedV3)                  │  │
│  │  └── Safe/finalized tag updates                                  │  │
│  │                                                                  │  │
│  │  restart.spec.ts                                                 │  │
│  │  ├── Coordinator restart recovery                                │  │
│  │  └── Anchoring resume after restart                              │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                   Infrastructure Tests                           │  │
│  │                                                                  │  │
│  │  linea-besu-fleet.spec.ts                                        │  │
│  │  ├── Leader/follower consistency                                 │  │
│  │  ├── linea_estimateGas matching                                  │  │
│  │  └── Block sync verification                                     │  │
│  │                                                                  │  │
│  │  liveness.spec.ts                                                │  │
│  │  └── Sequencer downtime detection                                │  │
│  │                                                                  │  │
│  │  l2.spec.ts                                                      │  │
│  │  ├── Transaction type support                                    │  │
│  │  └── Calldata size limits                                        │  │
│  │                                                                  │  │
│  │  opcodes.spec.ts                                                 │  │
│  │  └── EVM opcode execution                                        │  │
│  │                                                                  │  │
│  │  send-bundle.spec.ts                                             │  │
│  │  └── Transaction bundling                                        │  │
│  │                                                                  │  │
│  │  transaction-exclusion.spec.ts                                   │  │
│  │  └── Rejection tracking                                          │  │
│  │                                                                  │  │
│  │  shomei-get-proof.spec.ts                                        │  │
│  │  └── Merkle proof generation                                     │  │
│  │                                                                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
e2e/
├── src/
│   ├── bridge-tokens.spec.ts
│   ├── messaging.spec.ts
│   ├── submission-finalization.spec.ts
│   ├── restart.spec.ts
│   ├── linea-besu-fleet.spec.ts
│   ├── liveness.spec.ts
│   ├── l2.spec.ts
│   ├── opcodes.spec.ts
│   ├── send-bundle.spec.ts
│   ├── transaction-exclusion.spec.ts
│   ├── shomei-get-proof.spec.ts
│   │
│   ├── common/
│   │   ├── utils.ts              # Shared utilities
│   │   ├── constants.ts          # Test constants
│   │   ├── deployments.ts        # Contract deployments
│   │   ├── generateL2Traffic.ts  # Traffic generation
│   │   └── types.ts              # Type definitions
│   │
│   └── config/
│       ├── jest/
│       │   ├── global-setup.ts   # Test setup
│       │   ├── global-teardown.ts # Test cleanup
│       │   └── setup.ts
│       ├── logger/               # Logging utilities
│       └── tests-config/
│           ├── environments/
│           │   ├── local.ts      # Local stack config
│           │   ├── dev.ts        # Dev environment
│           │   └── sepolia.ts    # Testnet
│           └── accounts/         # Account management
│
├── jest.config.ts
└── package.json
```

## Test Infrastructure

### Prerequisites

```bash
# Start full stack
make start-env-with-tracing-v2-ci

# Verify services
docker compose -f docker/compose-tracing-v2-ci-extension.yml ps
```

### Configuration

```typescript
// config/local.ts
export const config = {
  l1: {
    rpcUrl: 'http://localhost:8445',
    chainId: 31648428,
  },
  l2: {
    rpcUrl: 'http://localhost:8545',
    chainId: 1337,
  },
  contracts: {
    lineaRollup: '0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9',
    l2MessageService: '0xe537D669CA013d86EBeF1D64e40fC74CADC91987',
    tokenBridge: {
      l1: '0x...',
      l2: '0x...',
    },
  },
  accounts: {
    deployer: {
      privateKey: '0xac0974bec...',
    },
  },
};
```

### Global Setup

```typescript
// global-setup.ts
export default async function globalSetup() {
  // Deploy test contracts
  await deployTestContracts();
  
  // Start background L2 traffic
  await startTrafficGeneration();
  
  console.log('E2E setup complete');
}

async function deployTestContracts() {
  const deployer = new Wallet(config.accounts.deployer.privateKey);
  
  // Deploy DummyContract
  const DummyContract = await ethers.getContractFactory('DummyContract');
  await DummyContract.connect(deployer).deploy();
  
  // Deploy TestContract
  const TestContract = await ethers.getContractFactory('TestContract');
  await TestContract.connect(deployer).deploy();
  
  // Deploy LineaSequencerUptimeFeed
  // ...
}
```

## Test Examples

### Bridge Tokens Test

```typescript
// bridge-tokens.spec.ts
describe('Bridge Tokens', () => {
  describe('L1 → L2', () => {
    it('should bridge ERC20 from L1 to L2', async () => {
      const { l1Wallet, l2Wallet } = await generateAccounts();
      
      // Mint test tokens on L1
      await testToken.connect(l1Wallet).mint(
        l1Wallet.address,
        parseEther('100')
      );
      
      // Approve token bridge
      await testToken.connect(l1Wallet).approve(
        tokenBridge.address,
        parseEther('10')
      );
      
      // Bridge tokens
      const tx = await tokenBridge.connect(l1Wallet).bridgeToken(
        testToken.address,
        parseEther('10'),
        l2Wallet.address
      );
      
      // Wait for anchoring
      await waitForEvents(
        l2MessageService,
        'RollingHashUpdated',
        1
      );
      
      // Claim on L2
      const message = await getMessageFromTx(tx);
      await l2MessageService.connect(l2Wallet).claimMessage(
        message.from,
        message.to,
        message.fee,
        message.value,
        l2Wallet.address,
        message.calldata,
        message.nonce
      );
      
      // Verify balance
      const bridgedToken = await getBridgedTokenAddress(testToken.address);
      const balance = await IERC20(bridgedToken).balanceOf(l2Wallet.address);
      expect(balance).to.equal(parseEther('10'));
    });
  });
});
```

### Submission & Finalization Test

```typescript
// submission-finalization.spec.ts
describe('Submission and Finalization', () => {
  it('should submit and finalize blocks', async () => {
    // Generate L2 transactions
    await generateL2Transactions(10);
    
    // Wait for data submission
    const submissionEvent = await waitForEvents(
      lineaRollup,
      'DataSubmittedV3',
      1,
      { timeout: 120000 }
    );
    
    expect(submissionEvent.shnarf).to.not.be.null;
    
    // Wait for finalization
    const finalizationEvent = await waitForEvents(
      lineaRollup,
      'DataFinalizedV3',
      1,
      { timeout: 300000 }
    );
    
    expect(finalizationEvent.endBlockNumber).to.be.gt(0);
    
    // Verify safe block updated
    const safeBlock = await l2Provider.getBlock('safe');
    expect(safeBlock.number).to.be.gte(finalizationEvent.endBlockNumber);
  });
});
```

### Coordinator Restart Test

```typescript
// restart.spec.ts
describe('Coordinator Restart', () => {
  it('should resume blob submission after restart', async () => {
    // Get initial finalized block
    const initialFinalized = await lineaRollup.currentL2BlockNumber();
    
    // Restart coordinator
    await execDockerCommand('docker restart coordinator');
    
    // Wait for coordinator to be healthy
    await waitForHealthy('coordinator');
    
    // Generate more transactions
    await generateL2Transactions(10);
    
    // Verify blob submission resumes
    const newFinalized = await awaitUntil(
      async () => lineaRollup.currentL2BlockNumber(),
      (blockNum) => blockNum > initialFinalized,
      { timeout: 300000, interval: 5000 }
    );
    
    expect(newFinalized).to.be.gt(initialFinalized);
  });
});
```

### Fleet Consistency Test

```typescript
// linea-besu-fleet.spec.ts
describe('Fleet Consistency', () => {
  it('should have consistent linea_estimateGas across nodes', async () => {
    const tx = {
      from: account.address,
      to: testContract.address,
      data: testContract.interface.encodeFunctionData('doSomething'),
    };
    
    // Get estimate from leader
    const leaderEstimate = await leaderClient.estimateGas(tx);
    
    // Get estimate from follower
    const followerEstimate = await followerClient.estimateGas(tx);
    
    expect(leaderEstimate).to.equal(followerEstimate);
  });
  
  it('should have matching finalized blocks', async () => {
    const leaderBlock = await leaderProvider.getBlock('finalized');
    const followerBlock = await followerProvider.getBlock('finalized');
    
    expect(leaderBlock.hash).to.equal(followerBlock.hash);
  });
});
```

## Running Tests

```bash
cd e2e

# Install dependencies (generates TypeChain types)
pnpm install

# Run all tests (except fleet and liveness)
pnpm run test:e2e:local

# Run specific test file
pnpm run test:e2e:local -- bridge-tokens.spec.ts

# Run fleet tests
pnpm run test:e2e:fleet:local

# Run liveness tests
pnpm run test:e2e:liveness:local
```

## Jest Configuration

```typescript
// jest.config.ts
export default {
  preset: 'ts-jest',
  testEnvironment: 'node',
  testTimeout: 180000,          // 3 minutes
  maxConcurrency: 7,
  maxWorkers: '75%',
  globalSetup: './src/global-setup.ts',
  globalTeardown: './src/global-teardown.ts',
  testMatch: ['**/*.spec.ts'],
  testPathIgnorePatterns: [
    'linea-besu-fleet.spec.ts',
    'liveness.spec.ts',
  ],
};
```

## Utility Functions

### Wait for Events

```typescript
// utils/events.ts
export async function waitForEvents<T>(
  contract: Contract,
  eventName: string,
  count: number,
  options: { timeout?: number } = {}
): Promise<T[]> {
  const { timeout = 60000 } = options;
  const events: T[] = [];
  
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`Timeout waiting for ${count} ${eventName} events`));
    }, timeout);
    
    contract.on(eventName, (...args) => {
      events.push(args as unknown as T);
      if (events.length >= count) {
        clearTimeout(timer);
        contract.removeAllListeners(eventName);
        resolve(events);
      }
    });
  });
}
```

### Await Until

```typescript
// utils/common.ts
export async function awaitUntil<T>(
  fn: () => Promise<T>,
  predicate: (value: T) => boolean,
  options: { timeout?: number; interval?: number } = {}
): Promise<T> {
  const { timeout = 60000, interval = 1000 } = options;
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    const value = await fn();
    if (predicate(value)) {
      return value;
    }
    await sleep(interval);
  }
  
  throw new Error('Timeout in awaitUntil');
}
```

## Dependencies

- **Jest**: Test framework
- **ethers**: Ethereum interactions
- **TypeChain**: Contract type generation
- **Docker**: Container management (restart, exec)

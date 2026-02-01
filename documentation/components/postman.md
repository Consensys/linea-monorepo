# Postman

> TypeScript service for automated cross-chain message claiming.

> **Diagrams:** [Postman Architecture](../diagrams/postman-architecture.mmd) | [Message Lifecycle](../diagrams/message-lifecycle.mmd)

## Overview

The Postman service:
- Monitors L1 and L2 for `MessageSent` events
- Checks when messages become claimable
- Automatically claims messages on destination chains
- Supports fee-based and sponsored claiming

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           POSTMAN SERVICE                                │
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │                     PostmanServiceClient                           │  │
│  │                    (Main Orchestrator)                             │  │
│  └──────────────────────────────┬─────────────────────────────────────┘  │
│                                 │                                        │
│        ┌────────────────────────┴────────────────────────┐               │
│        │                                                 │               │
│        ▼                                                 ▼               │
│  ┌─────────────────────────────────┐  ┌─────────────────────────────────┐│
│  │       L1 → L2 Pipeline          │  │       L2 → L1 Pipeline          ││
│  │                                 │  │                                 ││
│  │ ┌─────────────────────────────┐ │  │ ┌─────────────────────────────┐ ││
│  │ │ MessageSentEventPoller      │ │  │ │ MessageSentEventPoller      │ ││
│  │ │ (Watch L1 LineaRollup)      │ │  │ │ (Watch L2 MessageService)   │ ││
│  │ └──────────────┬──────────────┘ │  │ └──────────────┬──────────────┘ ││
│  │                ▼                │  │                ▼                ││
│  │ ┌─────────────────────────────┐ │  │ ┌─────────────────────────────┐ ││
│  │ │ MessageAnchoringPoller      │ │  │ │ MessageAnchoringPoller      │ ││
│  │ │ (Check L2 claimability)     │ │  │ │ (Check L1 claimability)     │ ││
│  │ └──────────────┬──────────────┘ │  │ └──────────────┬──────────────┘ ││
│  │                ▼                │  │                ▼                ││
│  │ ┌─────────────────────────────┐ │  │ ┌─────────────────────────────┐ ││
│  │ │ MessageClaimingPoller       │ │  │ │ MessageClaimingPoller       │ ││
│  │ │ (Claim on L2)               │ │  │ │ (Claim on L1 + proof)       │ ││
│  │ └──────────────┬──────────────┘ │  │ └──────────────┬──────────────┘ ││
│  │                ▼                │  │                ▼                ││
│  │ ┌─────────────────────────────┐ │  │ ┌─────────────────────────────┐ ││
│  │ │ MessagePersistingPoller     │ │  │ │ MessagePersistingPoller     │ ││
│  │ │ (Track claim receipts)      │ │  │ │ (Track claim receipts)      │ ││
│  │ └─────────────────────────────┘ │  │ └─────────────────────────────┘ ││
│  │                                 │  │                                 ││
│  └─────────────────────────────────┘  └─────────────────────────────────┘│
│                                                                          │
│  ┌────────────────────────────────────────────────────────────────────┐  │
│  │                        Shared Services                             │  │
│  │                                                                    │  │
│  │ ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐   │  │
│  │ │  PostgreSQL  │  │   Metrics    │  │  Transaction Validation  │   │  │
│  │ │  (Messages)  │  │  (Prometheus)│  │  (Gas, Fees, Limits)     │   │  │
│  │ └──────────────┘  └──────────────┘  └──────────────────────────┘   │  │
│  │                                                                    │  │
│  └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
postman/
├── scripts/
│   └── runPostman.ts                    # Entry point
│
├── src/
│   ├── application/
│   │   └── postman/
│   │       ├── app/
│   │       │   ├── PostmanServiceClient.ts  # Main orchestrator
│   │       │   └── config/
│   │       │       └── config.ts
│   │       │
│   │       ├── persistence/
│   │       │   ├── entities/
│   │       │   │   └── Message.entity.ts
│   │       │   └── repositories/
│   │       │       └── TypeOrmMessageRepository.ts
│   │       │
│   │       └── api/
│   │           └── metrics/             # Metrics endpoints
│   │
│   ├── core/                            # Core domain logic
│   │
│   └── services/
│       ├── pollers/
│       │   ├── MessageSentEventPoller.ts
│       │   ├── MessageAnchoringPoller.ts
│       │   ├── MessageClaimingPoller.ts
│       │   ├── MessagePersistingPoller.ts
│       │   └── L2ClaimMessageTransactionSizePoller.ts
│       │
│       ├── processors/
│       │   ├── MessageClaimingProcessor.ts
│       │   └── MessageClaimingPersister.ts
│       │
│       └── TransactionValidationService.ts
│
└── package.json
```

## Message Lifecycle

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         MESSAGE LIFECYCLE                                │
│                                                                          │
│  SENT         ANCHORED        CLAIMING        CLAIMED_SUCCESS            │
│   │               │               │                  │                   │
│   │  Detected     │  Claimable    │  TX submitted    │  Complete         │
│   │  on source    │  on dest      │                  │                   │
│   ▼               ▼               ▼                  ▼                   │
│  ┌─┐             ┌─┐             ┌─┐                ┌─┐                  │
│  │●│────────────▶│●│────────────▶│●│───────────────▶│●│                  │
│  └─┘             └─┘             └─┘                └─┘                  │
│                                   │                                      │
│                                   │  On failure                          │
│                                   ▼                                      │
│                                  ┌─┐                                     │
│                              ┌──▶│●│ CLAIMED_REVERTED                    │
│                              │   └─┘                                     │
│                              │    │                                      │
│                              │    │  Retry                               │
│                              │    ▼                                      │
│                              │   ┌─┐                                     │
│                              └───│●│ Back to SENT                        │
│                                  └─┘                                     │
│                                                                          │
│  Alternative States:                                                     │
│  - EXCLUDED: Filtered out by rules                                       │
│  - CLAIMED_ALREADY: Already claimed by someone else                      │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

## Polling Services

### 1. MessageSentEventPoller

Monitors chains for new messages.

```typescript
class MessageSentEventPoller {
  async poll(): Promise<void> {
    // Get last processed block
    const fromBlock = await this.getLastProcessedBlock();
    
    // Fetch MessageSent events
    const events = await this.contract.queryFilter(
      'MessageSent',
      fromBlock,
      toBlock
    );
    
    // Apply filters (address, calldata)
    const filtered = events.filter(this.applyFilters);
    
    // Save to database
    for (const event of filtered) {
      await this.messageRepository.save({
        messageHash: event.messageHash,
        from: event.from,
        to: event.to,
        fee: event.fee,
        value: event.value,
        nonce: event.nonce,
        calldata: event.calldata,
        status: 'SENT',
      });
    }
  }
}
```

### 2. MessageAnchoringPoller

Checks if messages are claimable.

```typescript
class MessageAnchoringPoller {
  async poll(): Promise<void> {
    // Get messages in SENT status
    const messages = await this.messageRepository.findByStatus('SENT');
    
    for (const message of messages) {
      // Check on-chain status
      const status = await this.contract.getMessageStatus(
        message.messageHash
      );
      
      if (status === OnChainMessageStatus.CLAIMABLE) {
        await this.messageRepository.updateStatus(
          message.id,
          'ANCHORED'
        );
      } else if (status === OnChainMessageStatus.CLAIMED) {
        await this.messageRepository.updateStatus(
          message.id,
          'CLAIMED_SUCCESS'
        );
      }
    }
  }
}
```

### 3. MessageClaimingPoller

Claims messages on destination chain.

```typescript
class MessageClaimingPoller {
  async poll(): Promise<void> {
    // Get next claimable message
    const message = await this.getNextClaimableMessage();
    if (!message) return;
    
    // Validate transaction
    const validation = await this.validationService.validate(message);
    if (!validation.valid) {
      await this.handleValidationFailure(message, validation);
      return;
    }
    
    // Execute claim
    const tx = await this.contract.claim({
      from: message.from,
      to: message.to,
      fee: message.fee,
      value: message.value,
      nonce: message.nonce,
      calldata: message.calldata,
      feeRecipient: this.feeRecipient,
    });
    
    // Update status
    await this.messageRepository.updateWithTx(message.id, {
      status: 'CLAIMING',
      claimTxHash: tx.hash,
    });
  }
}
```

### 4. MessagePersistingPoller

Tracks claim transaction receipts.

```typescript
class MessagePersistingPoller {
  async poll(): Promise<void> {
    // Get messages in CLAIMING status
    const messages = await this.messageRepository.findByStatus('CLAIMING');
    
    for (const message of messages) {
      const receipt = await this.provider.getTransactionReceipt(
        message.claimTxHash
      );
      
      if (!receipt) {
        // Check timeout, retry if needed
        await this.handlePendingTx(message);
        continue;
      }
      
      if (receipt.status === 1) {
        await this.messageRepository.updateStatus(
          message.id,
          'CLAIMED_SUCCESS'
        );
      } else {
        await this.messageRepository.updateStatus(
          message.id,
          'CLAIMED_REVERTED'
        );
      }
    }
  }
}
```

## Transaction Validation

```typescript
interface TransactionValidation {
  valid: boolean;
  reason?: string;
}

class TransactionValidationService {
  async validate(message: Message): Promise<TransactionValidation> {
    // 1. Check fee is non-zero
    if (message.fee === 0n) {
      return { valid: false, reason: 'ZERO_FEE' };
    }
    
    // 2. Check gas limit
    const gasEstimate = await this.estimateGas(message);
    if (gasEstimate > this.maxClaimGasLimit) {
      return { valid: false, reason: 'GAS_LIMIT_EXCEEDED' };
    }
    
    // 3. Check profitability
    const gasCost = gasEstimate * await this.getGasPrice();
    if (message.fee < gasCost * this.profitMargin) {
      return { valid: false, reason: 'UNDERPRICED' };
    }
    
    // 4. Check rate limit
    if (await this.isRateLimited(message)) {
      return { valid: false, reason: 'RATE_LIMITED' };
    }
    
    return { valid: true };
  }
}
```

## Configuration

### Environment Variables

```bash
# RPC Endpoints
L1_RPC_URL=https://mainnet.infura.io/v3/...
L2_RPC_URL=https://rpc.linea.build

# Contract Addresses
L1_CONTRACT_ADDRESS=0xd19d4B5d358258f05D7B411E21A1460D11B0876F
L2_CONTRACT_ADDRESS=0x508Ca82Df566dCD1B0DE8296e70a96332cD644ec

# Signer
PRIVATE_KEY=0x...

# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postman

# Polling Intervals
L1_LISTENER_INTERVAL=10000  # 10 seconds
L2_LISTENER_INTERVAL=5000   # 5 seconds

# Claiming Configuration
MAX_CLAIM_GAS_LIMIT=500000
PROFIT_MARGIN=1.1  # 10% profit margin

# Feature Flags
L1_L2_AUTO_CLAIM_ENABLED=true
L2_L1_AUTO_CLAIM_ENABLED=true
L1_L2_EOA_MESSAGES_ENABLED=true
L2_L1_EOA_MESSAGES_ENABLED=true

# Filtering
L1_L2_FROM_ADDRESS_FILTER=0x...
L1_L2_TO_ADDRESS_FILTER=0x...
L1_L2_CALLDATA_FILTER=startsWith(calldata, "0x1234")
```

## Database Schema

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    message_hash VARCHAR(66) UNIQUE NOT NULL,
    direction VARCHAR(10) NOT NULL,  -- 'L1_L2' or 'L2_L1'
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    fee NUMERIC(78, 0) NOT NULL,
    value NUMERIC(78, 0) NOT NULL,
    nonce NUMERIC(78, 0) NOT NULL,
    calldata TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    source_block_number BIGINT NOT NULL,
    source_tx_hash VARCHAR(66) NOT NULL,
    claim_tx_hash VARCHAR(66),
    claim_gas_price NUMERIC(78, 0),
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_messages_status ON messages(status);
CREATE INDEX idx_messages_direction ON messages(direction);
CREATE INDEX idx_messages_message_hash ON messages(message_hash);
```

## Building

```bash
cd postman

# Install dependencies
pnpm install

# Build
pnpm run build

# Run
pnpm run start

# Development mode
pnpm run dev
```

## Docker

```bash
docker build -t consensys/linea-postman .

docker run \
  -e L1_RPC_URL=... \
  -e L2_RPC_URL=... \
  -e PRIVATE_KEY=... \
  -e DATABASE_URL=... \
  consensys/linea-postman
```

## Metrics

Exposed at `/metrics`:

- `postman_messages_sent_total{direction}`
- `postman_messages_claimed_total{direction,status}`
- `postman_claim_gas_used{direction}`
- `postman_sponsorship_fees_paid{direction}`
- `postman_processing_time_seconds{stage}`

## Dependencies

- **@consensys/linea-sdk**: Contract interactions
- **@consensys/linea-shared-utils**: Logging, metrics
- **@consensys/linea-native-libs**: Transaction compression
- **TypeORM**: Database ORM
- **PostgreSQL**: Message persistence

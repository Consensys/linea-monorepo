# Yield Boost — ETH Fund Flow

## Overview

The Linea Yield Boost system stakes LineaRollup funds on Ethereum L1 via Lido V3 stVaults to generate beacon chain rewards, which are reported to L2 for distribution. This document maps **where ETH flows** and **which roles control each movement**.

## High-Level Architecture

```mermaid
flowchart TD
    User([User])
    LR[LineaRollup]
    YM[YieldManager]
    DB[Dashboard]
    SV[StakingVault]
    NO([Node Operator])
    BC{{Beacon Chain}}

    User -->|deposit ETH| LR
    LR -->|excess ETH| YM
    YM -->|fund| DB
    DB -->|fund| SV
    NO -->|deposit to validators| SV
    SV -->|stake| BC
    BC -->|withdrawal| SV
    SV -->|withdraw| DB
    DB -->|withdraw ETH| LR
    LR -->|withdraw ETH| User

    classDef admin fill:#f8d7da,stroke:#721c24,stroke-width:1px,color:#721c24
    classDef permissionless fill:#cce5ff,stroke:#004085,stroke-width:1px,color:#004085
    classDef l2 fill:#d4edda,stroke:#155724,stroke-width:1px,color:#155724

    class NO admin
    class User permissionless
```


```mermaid
flowchart TD
    User([User])
    LR[LineaRollup]
    YM[YieldManager]
    DB[Dashboard]
    SV[StakingVault]
    NO([Node Operator])
    BC{{Beacon Chain}}
    L2MS[L2MessageService]
    L2YD[L2YieldDistributor]

    User -->|deposit ETH| LR
    LR -->|excess ETH| YM
    YM -->|fund| DB
    DB -->|fund| SV
    NO -->|deposit to validators| SV
    SV -->|stake| BC
    BC -->|withdrawal| SV
    SV -->|withdraw| DB
    DB -->|withdraw ETH| LR
    LR -->|withdraw ETH| User

    YM -.->|reportNativeYield| LR
    LR -.->|MessageSent| L2MS
    L2MS -.->|distribute yield| L2YD

    classDef admin fill:#f8d7da,stroke:#721c24,stroke-width:1px,color:#721c24
    classDef permissionless fill:#cce5ff,stroke:#004085,stroke-width:1px,color:#004085
    classDef l2 fill:#d4edda,stroke:#155724,stroke-width:1px,color:#155724

    class NO admin
    class User permissionless
    class L2MS,L2YD l2
```

**Legend:** Red = privileged operator | Blue = permissionless | Green = L2 | Dashed = synthetic message (no L1 ETH moves)

## Roles & Fund Movement Permissions

| Role | Held By | ETH Movement Authorized |
|------|---------|------------------------|
| `YIELD_PROVIDER_STAKING_ROLE` | Automation Service | LineaRollup → YieldManager → StakingVault |
| `YIELD_PROVIDER_UNSTAKER_ROLE` | Automation Service | StakingVault → YieldManager → LineaRollup |
| `YIELD_REPORTER_ROLE` | Automation Service | None (synthetic message only) |
| `STAKING_PAUSE_CONTROLLER_ROLE` | Security Council | None (pauses/unpauses beacon deposits) |
| `OSSIFICATION_INITIATOR_ROLE` | Security Council | Initiates full withdrawal of all staked funds |
| `OSSIFICATION_PROCESSOR_ROLE` | Automation Service | Progressive withdrawal during ossification |
| `SET_YIELD_MANAGER_ROLE` | Security Council | None (configuration) |
| Permissionless | Anyone | `fund()` donations; `unstakePermissionless()` + `replenishWithdrawalReserve()` during deficit |

## Fund Flow Scenarios

### 1. Staking

Surplus ETH in LineaRollup (above minimum reserve) is routed to the StakingVault for beacon chain staking.

```mermaid
sequenceDiagram
    actor Auto as Automation Service
    participant LR as LineaRollup
    participant YM as YieldManager
    participant DB as Dashboard
    participant SV as StakingVault
    actor NO as Node Operator
    participant BC as Beacon Chain

    Auto->>LR: transferFundsForNativeYield()
    LR-->>YM: send ETH
    Auto->>YM: fundYieldProvider()
    YM-->>DB: fund()
    DB-->>SV: fund()
    Note over NO: Decides when/how much to deposit
    NO->>SV: depositToBeaconChain()
    SV-->>BC: stake ETH
```

### 2. Yield Reporting

No L1 ETH moves — a synthetic `MessageSent` event relays net yield to L2 for distribution.

```mermaid
flowchart LR
    Auto[Automation Service] -->|reportYield| YM[YieldManager]
    YM -->|reportNativeYield| LR[LineaRollup]
    LR -.->|MessageSent| L2[L2MessageService]
    L2 -.->|distribute| YD[L2YieldDistributor]

    classDef l2 fill:#d4edda,stroke:#155724,stroke-width:1px,color:#155724
    class L2,YD l2
```

Yield is reported **after** deducting LST liabilities, Lido protocol fees, and node operator fees.

### 3. Reserve Replenishment — Operator

Two-phase process: first trigger beacon chain withdrawal, then route funds to LineaRollup once they arrive in the vault.

```mermaid
sequenceDiagram
    actor Auto as Automation Service
    participant LR as LineaRollup
    participant YM as YieldManager
    participant SV as StakingVault
    participant BC as Beacon Chain

    Note over Auto: Phase 1 — Unstake
    Auto->>YM: unstake()
    YM-->>SV: triggerValidatorWithdrawals()
    SV-->>BC: request withdrawal
    Note over BC: Withdrawal delay
    BC->>SV: automatic withdrawal

    Note over Auto: Phase 2 — Replenish reserve
    Auto->>YM: safeAddToWithdrawalReserve()
    YM-->>SV: withdraw()
    SV-->>LR: send ETH
```

### 4. Permissionless Flows

When LineaRollup balance drops below the minimum reserve, anyone can trigger unstaking and reserve replenishment.

```mermaid
sequenceDiagram
    actor U as Anyone
    participant LR as LineaRollup
    participant YM as YieldManager
    participant SV as StakingVault
    participant BC as Beacon Chain

    Note over LR: Reserve in deficit
    U->>YM: unstakePermissionless()
    Note over YM: Validate against EIP-4788 beacon root
    YM-->>SV: triggerValidatorWithdrawals()
    SV-->>BC: request withdrawal
    Note over BC: Withdrawal delay
    BC->>SV: automatic withdrawal

    U->>YM: replenishWithdrawalReserve()
    YM-->>SV: withdraw()
    SV-->>LR: send ETH
```

`unstakePermissionless()` is capped to the remaining deficit minus available liquidity in the YieldManager and provider.

### 5. LST Withdrawal — Last Resort

When LineaRollup lacks sufficient ETH for a user withdrawal, stETH is minted against StakingVault collateral and sent directly to the user.

```mermaid
sequenceDiagram
    actor U as User
    participant LR as LineaRollup
    participant YM as YieldManager
    participant DB as Dashboard

    Note over LR: Insufficient ETH for withdrawal
    U->>LR: claimMessageWithProofAndWithdrawLST()
    LR->>YM: withdrawLST()
    YM->>DB: mintStETH()
    DB-->>U: send stETH
    Note over YM: Pauses beacon deposits,<br/>creates interest-bearing liability
```

This creates an LST liability that accrues interest; the system prioritizes repaying it from subsequent deposits and yield.

### 6. Ossification Withdrawal

Security Council initiates permanent vault freeze; Automation Service progressively withdraws all funds.

```mermaid
sequenceDiagram
    actor SC as Security Council
    actor Auto as Automation Service
    participant YM as YieldManager
    participant DB as Dashboard
    participant SV as StakingVault
    participant LR as LineaRollup

    SC->>YM: initiateOssification()
    YM-->>DB: voluntaryDisconnect()
    Note over DB: Pending disconnection

    Auto->>YM: progressPendingOssification()
    YM-->>SV: ossify()
    Note over SV: Vault permanently frozen

    loop Until all funds withdrawn
        Auto->>YM: unstake()
        Note over SV: Beacon chain withdrawal
        Auto->>YM: safeAddToWithdrawalReserve()
        YM-->>SV: withdraw()
        SV-->>LR: send ETH
    end
```

An accounting report must be submitted between `initiateOssification()` and `progressPendingOssification()`. Partial validator withdrawals are preferred to bypass the Ethereum exit queue.

## Quick Reference

| Fund Movement | Source | Destination | Trigger | Role Required |
|--------------|--------|-------------|---------|---------------|
| Stake excess reserve | LineaRollup | StakingVault | Automation | `YIELD_PROVIDER_STAKING_ROLE` |
| Beacon chain deposit | StakingVault | Validators | Node Operator decision | Node Operator |
| Report yield to L2 | — (synthetic) | L2YieldDistributor | Automation | `YIELD_REPORTER_ROLE` |
| Operator replenish reserve | StakingVault | LineaRollup | Reserve below target | `YIELD_PROVIDER_UNSTAKER_ROLE` |
| Permissionless unstake | StakingVault | LineaRollup | Reserve below minimum | Permissionless |
| LST withdrawal | StakingVault (mint) | User | Insufficient ETH | Permissionless (user) |
| Ossification withdrawal | StakingVault | LineaRollup | Security Council initiates | `OSSIFICATION_INITIATOR_ROLE` + `OSSIFICATION_PROCESSOR_ROLE` |
| Donation | External | LineaRollup | Voluntary | Permissionless |
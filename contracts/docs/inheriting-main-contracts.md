# Customizing Linea Rollup and Bridging Components

This guide explains how to customize core rollup, messaging, and bridging functionality by overriding virtual functions in the Linea protocol contracts. Several examples are provided to help you get started.

## 🧱 Customizing Linea Rollup Logic

To modify the behavior of the rollup mechanism itself (e.g. state roots, batch handling, fraud proofs), you can override virtual functions in `LineaRollup.sol`.

An example implementation is provided in [`InheritingLineaRollup.sol`](../src/_testing/unit/rollup/InheritingLineaRollup.sol), where selected virtual functions are overridden to customize rollup logic.

## ✉️ Customizing Message Sending

The `MessageService` contract handles L1 ↔ L2 message passing. If you want to adjust how messages are sent or verified (e.g., adding metadata, changing fee logic), you can override its virtual functions.

See [`InheritingLineaRollup.sol`](../src/_testing/unit/rollup/InheritingLineaRollup.sol) for an example where `MessageService` functionality is customized within the rollup context.

## 🔁 Overriding Token Bridge Logic

To implement custom behavior for token bridging (e.g., additional validation, whitelisting, alternative transfer mechanisms), override functions in the `TokenBridge` contract.

A reference implementation is available in [`InheritingTokenBridge.sol`](./InheritingTokenBridge.sol), demonstrating how to override core bridge functions for ERC20 tokens.

## 📨 Overriding L2 Message Service Logic

If you need to modify how messages are processed or validated on the L2 side, override functions in the `L2MessageService` contract.

Check out [`InheritingL2MessageService.sol`](../src/_testing/unit/messaging/InheritingL2MessageService.sol) for a practical example of extending L2 message processing behavior.

## 🪙 Custom Ether Bridge Examples

For more advanced use cases involving Ether transfers—such as delayed withdrawals or multi-phase bridging—you can extend and override the send/receive logic in:

- [`L1GenericBridge.sol`](../src/_testing/unit/bridging/l1/L1GenericBridge.sol)
- [`L2GenericBridge.sol`](../src/_testing/unit/bridging/l2/L2GenericBridge.sol)

These sample files serve as a foundation for creating fully custom Ether bridging flows tailored to your application needs.

---

By leveraging inheritance and Solidity's `virtual`/`override` keywords, you can safely extend and adapt Linea's modular bridge and rollup system.

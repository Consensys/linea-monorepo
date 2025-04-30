# 🧩 Blob Submission & Finalization

This document outlines the core data and finalization flows involved in LineaRollup's lifecycle, including blob commitment and zk-proof-based finality.

---

## 📦 Blob Submission

This flow is used by the **Data Submission Operator** to submit blobs to the LineaRollup system.

### 🔄 Steps

1. **Data Submission Operator** calls `submitBlobs()` on the `LineaRollup` contract with **1 to N blobs**, where **N = network maximum**..
2. For each submitted blob:
   - The contract verifies data integrity.
   - Performs evaluation checks via the point evaluation precompile.
   - Computes a `shnarf` for internal tracking.
3. The final computed `shnarf` is stored.
4. A corresponding event is emitted to reflect successful blob(s) storage.

---

## 🧮 Finalization Submission

This flow finalizes 1 or more aggregated blob transaction submissions by verifying correct execution proven via zero-knowledge proofs.

### 🔄 Steps

1. **Finalization Submission Operator** calls `finalizeBlocks()`.
2. `LineaRollup` contract:
   - Verifies submitted data matches the last finalized state and is non-empty to verify proper continuity.
   - Validates the messaging rolling hash feedback loop preventing manipulation or censorship.
   - Emits events for L2 blocks containing L2 → L1 messages.
   - Stores **Merkle roots** of L2 messages for proof-based claiming.
3. Computes the **public input** to be verified.
4. Calls the Plonk-based **Verifier** to validate the provided zk-proof.
5. Upon success, updates finalized state:
   - Includes latest 
   - Stores the `currentL2BlockNumber`, `finalStateRootHash` and other related finalization state metadata.

---

<img src="../diagrams/blobSubmissionAndFinalization.png">


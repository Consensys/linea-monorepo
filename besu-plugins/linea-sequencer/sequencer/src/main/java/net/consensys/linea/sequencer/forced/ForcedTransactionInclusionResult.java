/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.forced;

/** Represents the outcome of a forced transaction inclusion attempt. */
public enum ForcedTransactionInclusionResult {
  /** Transaction was successfully included in the block. */
  INCLUDED,

  /** Transaction failed due to nonce mismatch. */
  BAD_NONCE,

  /** Transaction failed due to insufficient balance. */
  BAD_BALANCE,

  /** Transaction failed due to precompile call failure. */
  BAD_PRECOMPILE,

  /** Transaction failed due to exceeding log limits. */
  TOO_MANY_LOGS,

  /** Transaction failed because sender or recipient is on deny list. */
  FILTERED_ADDRESSES,

  /**
   * Transaction failed for an unrecognized reason. This is a transient status - the transaction
   * will be retried in the next block rather than being finalized with this status.
   */
  OTHER;

  /**
   * Returns true if this result is transient and the transaction should be retried.
   *
   * @return true if the transaction should be retried
   */
  public boolean shouldRetry() {
    return this == OTHER;
  }
}

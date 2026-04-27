/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.config;

public final class TransactionGasLimitCap {
  public static final int EIP_7825_MAX_TRANSACTION_GAS_LIMIT = 16_777_216;

  private TransactionGasLimitCap() {}

  public static int effectiveMaxTxGasLimit(final int configuredMaxTxGasLimit) {
    return Math.min(configuredMaxTxGasLimit, EIP_7825_MAX_TRANSACTION_GAS_LIMIT);
  }

  public static String gasLimitExceededMessage(
      final long gasLimit, final int configuredMaxTxGasLimit) {
    final int effectiveLimit = effectiveMaxTxGasLimit(configuredMaxTxGasLimit);

    if (configuredMaxTxGasLimit < EIP_7825_MAX_TRANSACTION_GAS_LIMIT) {
      return "Gas limit "
          + gasLimit
          + " exceeds configured maximum transaction gas limit of "
          + effectiveLimit
          + " (EIP-7825 maximum transaction gas limit is "
          + EIP_7825_MAX_TRANSACTION_GAS_LIMIT
          + ")";
    }

    return "Gas limit "
        + gasLimit
        + " exceeds EIP-7825 maximum transaction gas limit of "
        + effectiveLimit;
  }
}

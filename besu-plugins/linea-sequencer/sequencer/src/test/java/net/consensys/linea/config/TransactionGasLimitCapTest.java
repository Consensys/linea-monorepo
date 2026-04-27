/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.config;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.Test;

class TransactionGasLimitCapTest {

  @Test
  void effectiveMaxTxGasLimitUsesConfiguredLimitWhenLowerThanEip7825Cap() {
    final int configuredMaxTxGasLimit = 9_000_000;

    final int effectiveMaxTxGasLimit =
        TransactionGasLimitCap.effectiveMaxTxGasLimit(configuredMaxTxGasLimit);

    assertThat(effectiveMaxTxGasLimit).isEqualTo(configuredMaxTxGasLimit);
  }

  @Test
  void effectiveMaxTxGasLimitUsesEip7825CapWhenConfiguredLimitIsHigher() {
    final int configuredMaxTxGasLimit = 24_000_000;

    final int effectiveMaxTxGasLimit =
        TransactionGasLimitCap.effectiveMaxTxGasLimit(configuredMaxTxGasLimit);

    assertThat(effectiveMaxTxGasLimit)
        .isEqualTo(TransactionGasLimitCap.EIP_7825_MAX_TRANSACTION_GAS_LIMIT);
  }

  @Test
  void gasLimitExceededMessageUsesEip7825CapMessageWhenConfiguredLimitIsHigher() {
    final long gasLimit = 16_777_217L;
    final int configuredMaxTxGasLimit = 24_000_000;

    final String message =
        TransactionGasLimitCap.gasLimitExceededMessage(gasLimit, configuredMaxTxGasLimit);

    assertThat(message)
        .isEqualTo(
            "Gas limit 16777217 exceeds EIP-7825 maximum transaction gas limit of 16777216");
  }

  @Test
  void gasLimitExceededMessageUsesConfiguredLimitMessageWhenConfiguredLimitIsLower() {
    final long gasLimit = 9_000_001L;
    final int configuredMaxTxGasLimit = 9_000_000;

    final String message =
        TransactionGasLimitCap.gasLimitExceededMessage(gasLimit, configuredMaxTxGasLimit);

    assertThat(message)
        .isEqualTo(
            "Gas limit 9000001 exceeds configured maximum transaction gas limit of 9000000 "
                + "(EIP-7825 maximum transaction gas limit is 16777216)");
  }
}

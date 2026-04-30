/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.sequencer.txpoolvalidation.validators;

import static net.consensys.linea.config.TransactionGasLimitCap.EIP_7825_MAX_TRANSACTION_GAS_LIMIT;
import static org.assertj.core.api.Assertions.assertThat;

import java.util.Optional;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.Test;

public class GasLimitValidatorTest {

  @Test
  public void acceptsTransactionAtEip7825MaxTransactionGasLimit() {
    final GasLimitValidator validator = new GasLimitValidator(EIP_7825_MAX_TRANSACTION_GAS_LIMIT);
    final Transaction transaction = transactionWithGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT);

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result).isEmpty();
  }

  @Test
  public void rejectsTransactionAboveEip7825MaxWhenConfiguredAtEip7825Max() {
    final GasLimitValidator validator = new GasLimitValidator(EIP_7825_MAX_TRANSACTION_GAS_LIMIT);
    final Transaction transaction =
        transactionWithGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT + 1L);

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result)
        .contains("Gas limit 16777217 exceeds EIP-7825 maximum transaction gas limit of 16777216");
  }

  @Test
  public void rejectsTransactionAboveEip7825MaxWhenConfiguredAboveEip7825Max() {
    final GasLimitValidator validator = new GasLimitValidator(24_000_000);
    final Transaction transaction =
        transactionWithGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT + 1L);

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result)
        .contains("Gas limit 16777217 exceeds EIP-7825 maximum transaction gas limit of 16777216");
  }

  @Test
  public void rejectsTransactionAboveConfiguredMaxWhenConfiguredBelowEip7825Max() {
    final GasLimitValidator validator = new GasLimitValidator(9_000_000);
    final Transaction transaction = transactionWithGasLimit(9_000_001L);

    final Optional<String> result = validator.validateTransaction(transaction, false, false);

    assertThat(result)
        .contains(
            "Gas limit 9000001 exceeds configured maximum transaction gas limit of 9000000 "
                + "(EIP-7825 maximum transaction gas limit is 16777216)");
  }

  private static Transaction transactionWithGasLimit(final long gasLimit) {
    return org.hyperledger.besu.ethereum.core.Transaction.builder()
        .gasLimit(gasLimit)
        .gasPrice(Wei.ZERO)
        .payload(Bytes.EMPTY)
        .build();
  }
}

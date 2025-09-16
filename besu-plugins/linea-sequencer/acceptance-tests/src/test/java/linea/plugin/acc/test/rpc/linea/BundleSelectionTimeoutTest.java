/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test.rpc.linea;

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.util.Arrays;
import java.util.List;
import java.util.stream.IntStream;
import linea.plugin.acc.test.TestCommandLineOptionsBuilder;
import org.junit.jupiter.api.Test;
import org.web3j.tx.gas.DefaultGasProvider;

public class BundleSelectionTimeoutTest extends AbstractSendBundleTest {
  private static final long MAX_BUNDLE_SELECTION_TIME_MILLIS = 1000L;
  private static final BigInteger GAS_LIMIT = DefaultGasProvider.GAS_LIMIT;
  private static final BigInteger VALUE = BigInteger.ZERO;
  private static final BigInteger GAS_PRICE = BigInteger.TEN.pow(11);

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set(
            "--plugin-linea-max-bundle-selection-time-millis=",
            Long.toString(MAX_BUNDLE_SELECTION_TIME_MILLIS))
        .build();
  }

  @Test
  public void transactionOverModuleLineCountRemoved() throws Exception {
    final var mulmodExecutor = deployMulmodExecutor();

    final var calls =
        IntStream.rangeClosed(1, 2)
            .mapToObj(
                nonce ->
                    mulmodOperation(
                        mulmodExecutor,
                        accounts.getPrimaryBenefactor(),
                        nonce,
                        1_000,
                        BigInteger.valueOf(MAX_TX_GAS_LIMIT / 10)))
            .toArray(MulmodCall[]::new);

    final var rawTxs = Arrays.stream(calls).map(MulmodCall::rawTx).toArray(String[]::new);

    final var sendBundleRequest =
        new SendBundleRequest(new BundleParams(rawTxs, Integer.toHexString(2)));
    final var sendBundleResponse = sendBundleRequest.execute(minerNode.nodeRequests());

    final var transferTxHash =
        accountTransactions
            .createTransfer(accounts.getSecondaryBenefactor(), accounts.getPrimaryBenefactor(), 1)
            .execute(minerNode.nodeRequests());

    assertThat(sendBundleResponse.hasError()).isFalse();
    assertThat(sendBundleResponse.getResult().bundleHash()).isNotBlank();

    minerNode.verify(eth.expectSuccessfulTransactionReceipt(transferTxHash.toHexString()));
  }
}

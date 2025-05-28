/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package linea.plugin.acc.test;

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.util.List;
import net.consensys.linea.metrics.LineaMetricCategory;
import org.hyperledger.besu.tests.acceptance.dsl.WaitUtils;
import org.junit.jupiter.api.Test;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.DefaultBlockParameterName;
import org.web3j.protocol.core.methods.response.EthBlock;

public class LivenessPluginTest extends LineaPluginTestBase {
  private static final String TEST_REJECTED_TX_ENDPOINT = "http://localhost:8080";
  private static final String TEST_NODE_TYPE = "SEQUENCER";
  private static final String TEST_CONTRACT_ADDRESS = "0x0000000000000000000000000000000000000001";
  private static final String TEST_SIGNER_URL = "http://localhost:9000";
  private static final String TEST_SIGNER_KEY_ID = "test-key-id";
  private static final String TEST_SIGNER_ADDRESS = "0x1234567890123456789012345678901234567890";
  private static final long TEST_MAX_BLOCK_AGE_MILLISECONDS = 5000;
  private static final long TEST_CHECK_INTERVAL_MILLISECONDS = 2000;
  private static final boolean TEST_METRIC_ENABLED = true;
  private static final long TEST_GAS_LIMIT = 100000;
  private static final int TEST_MAX_RETRY_ATTEMPTS = 3;
  private static final long TEST_RETRY_DELAY_MILLISECONDS = 1000;

  @Override
  public List<String> getTestCliOptions() {
    return getTestCommandLineOptionsBuilder().build();
  }

  protected TestCommandLineOptionsBuilder getTestCommandLineOptionsBuilder() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-rejected-tx-endpoint=", TEST_REJECTED_TX_ENDPOINT)
        .set("--plugin-linea-node-type=", TEST_NODE_TYPE)
        .set("--plugin-linea-liveness-enabled=", "true")
        .set(
            "--plugin-linea-liveness-max-block-age-milliseconds=",
            String.valueOf(TEST_MAX_BLOCK_AGE_MILLISECONDS))
        .set(
            "--plugin-linea-liveness-check-interval-milliseconds=",
            String.valueOf(TEST_CHECK_INTERVAL_MILLISECONDS))
        .set("--plugin-linea-liveness-contract-address=", TEST_CONTRACT_ADDRESS)
        .set("--plugin-linea-liveness-signer-url=", TEST_SIGNER_URL)
        .set("--plugin-linea-liveness-signer-key-id=", TEST_SIGNER_KEY_ID)
        .set("--plugin-linea-liveness-signer-address=", TEST_SIGNER_ADDRESS)
        .set("--plugin-linea-liveness-metrics-enabled=", String.valueOf(TEST_METRIC_ENABLED))
        .set("--plugin-linea-liveness-gas-limit=", String.valueOf(TEST_GAS_LIMIT))
        .set("--plugin-linea-liveness-max-retry-attempts=", String.valueOf(TEST_MAX_RETRY_ATTEMPTS))
        .set(
            "--plugin-linea-liveness-retry-delay-milliseconds=",
            String.valueOf(TEST_RETRY_DELAY_MILLISECONDS));
  }

  @Test
  public void pluginShouldMonitorBlockTimestampsAndReportMetrics() throws Exception {
    Web3j web3j = minerNode.nodeRequests().eth();

    // Get initial block
    EthBlock.Block initialBlock =
        web3j.ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false).send().getBlock();
    long initialTimestamp = initialBlock.getTimestamp().longValue();
    BigInteger initialBlockNumber = initialBlock.getNumber();
    BigInteger targetBlockNumber = initialBlockNumber.add(BigInteger.valueOf(3));

    // Wait for at least 3 new blocks to be produced using the framework's WaitUtils
    WaitUtils.waitFor(
        10,
        () -> {
          EthBlock.Block currentBlock =
              web3j.ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false).send().getBlock();
          assertThat(currentBlock.getNumber()).isGreaterThanOrEqualTo(targetBlockNumber);
        });

    // Get the latest block after waiting
    EthBlock.Block latestBlock =
        web3j.ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false).send().getBlock();
    long latestTimestamp = latestBlock.getTimestamp().longValue();

    // Verify that blocks are being produced and timestamps are increasing
    assertThat(latestBlock.getNumber()).isGreaterThan(initialBlock.getNumber());
    assertThat(latestTimestamp).isGreaterThan(initialTimestamp);

    // Verify that liveness metrics are available and initialized
    double uptimeTransactionsMetric =
        getMetricValue(LineaMetricCategory.SEQUENCER_LIVENESS, "uptime_transactions", List.of());
    assertThat(uptimeTransactionsMetric).isNotNaN().isGreaterThanOrEqualTo(0);

    // Verify that the plugin has started properly
    String log = getLog();
    assertThat(log).contains("LivenessPlugin started");
  }
}

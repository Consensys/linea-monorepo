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

import java.util.List;
import java.util.concurrent.TimeUnit;
import linea.plugin.acc.test.LineaPluginTestBase;
import linea.plugin.acc.test.TestCommandLineOptionsBuilder;
import net.consensys.linea.metrics.LineaMetricCategory;
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
  private static final long TEST_MAX_BLOCK_AGE_MILLISECONDS = 5;
  private static final long TEST_CHECK_INTERVAL_MILLISECONDS = 2;
  private static final boolean TEST_METRIC_ENABLED = true;
  private static final long TEST_GAS_LIMIT = 100000;

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
        .set("--plugin-linea-liveness-metrics-enabled=", String.valueOf(TEST_METRIC_ENABLED))
        .set("--plugin-linea-liveness-gas-limit=", String.valueOf(TEST_GAS_LIMIT));
  }

  @Test
  public void pluginShouldMonitorBlockTimestampsAndReportMetrics() throws Exception {
    Web3j web3j = minerNode.nodeRequests().eth();

    // Get initial block
    EthBlock.Block initialBlock =
        web3j.ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false).send().getBlock();
    long initialTimestamp = initialBlock.getTimestamp().longValue();

    // Wait for a few blocks to be produced
    Thread.sleep(TimeUnit.SECONDS.toMillis(TEST_CHECK_INTERVAL_MILLISECONDS * 3));

    // Get the latest block
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

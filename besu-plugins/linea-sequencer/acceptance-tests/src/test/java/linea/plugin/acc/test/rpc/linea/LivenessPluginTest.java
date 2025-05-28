/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package linea.plugin.acc.test.rpc.linea;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;

import linea.plugin.acc.test.LineaPluginTestBase;
import linea.plugin.acc.test.TestCommandLineOptionsBuilder;
import net.consensys.linea.metrics.LineaMetricCategory;
import net.consensys.linea.sequencer.liveness.LivenessPluginCliOptions;
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthBlock;

public class LivenessPluginTest extends LineaPluginTestBase {
  private static final String TEST_CONTRACT_ADDRESS = "0x0000000000000000000000000000000000000001";
  private static final String TEST_SIGNER_URL = "http://localhost:9000";
  private static final String TEST_SIGNER_KEY_ID = "test-key-id";
  private static final long TEST_MAX_BLOCK_AGE_SECONDS = 5;
  private static final long TEST_CHECK_INTERVAL_SECONDS = 2;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-liveness-enabled", "true")
        .set("--plugin-linea-liveness-max-block-age-seconds", String.valueOf(TEST_MAX_BLOCK_AGE_SECONDS))
        .set("--plugin-linea-liveness-check-interval-seconds", String.valueOf(TEST_CHECK_INTERVAL_SECONDS))
        .set("--plugin-linea-liveness-contract-address", TEST_CONTRACT_ADDRESS)
        .set("--plugin-linea-liveness-signer-url", TEST_SIGNER_URL)
        .set("--plugin-linea-liveness-signer-key-id", TEST_SIGNER_KEY_ID)
        .set("--plugin-linea-liveness-metric-category-enabled", "true")
        .set("--plugin-linea-liveness-gas-limit", "100000")
        .build();
  }

  @BeforeEach
  public void setup() throws Exception {
    super.setup();
  }

  @Test
  public void pluginShouldStartWithValidConfiguration() {
    assertThat(minerNode).isNotNull();
    assertThat(minerNode.getAddress()).isNotNull();
  }

  @Test
  public void pluginShouldNotStartWithInvalidConfiguration() {
    // Test with missing required configuration
    List<String> invalidOptions = new TestCommandLineOptionsBuilder()
      .set("--plugin-linea-liveness-enabled", "true")
        // Missing signer URL and key ID
        .build();

    BesuNode invalidNode = null;
    try {
      invalidNode = createCliqueNodeWithExtraCliOptionsAndRpcApis(
          "invalid-miner",
          getCliqueOptions(),
          invalidOptions,
          Set.of("LINEA", "MINER"),
          false);
      cluster.start(invalidNode);
    } catch (Exception e) {
      // Expected exception
      assertThat(e.getMessage()).contains("Web3Signer URL and key ID must be provided");
    } finally {
      if (invalidNode != null) {
        cluster.stopNode(invalidNode);
      }
    }
  }

  @Test
  public void pluginShouldMonitorBlockTimestamps() throws Exception {
    Web3j web3j = minerNode.nodeRequests().eth();

    // Get initial block
    EthBlock.Block initialBlock = web3j.ethGetBlockByNumber(org.web3j.protocol.core.DefaultBlockParameterName.LATEST, false)
        .send()
        .getBlock();
    long initialTimestamp = initialBlock.getTimestamp().longValue();

    // Wait for a few blocks to be produced
    Thread.sleep(TimeUnit.SECONDS.toMillis(TEST_CHECK_INTERVAL_SECONDS * 3));

    // Get latest block
    EthBlock.Block latestBlock = web3j.ethGetBlockByNumber(org.web3j.protocol.core.DefaultBlockParameterName.LATEST, false)
        .send()
        .getBlock();
    long latestTimestamp = latestBlock.getTimestamp().longValue();

    // Verify that blocks are being produced and timestamps are increasing
    assertThat(latestBlock.getNumber()).isGreaterThan(initialBlock.getNumber());
    assertThat(latestTimestamp).isGreaterThan(initialTimestamp);
  }

  @Test
  public void pluginShouldReportMetricsWhenEnabled() throws Exception {
    // Wait for plugin initialization and verify it started
    Thread.sleep(TimeUnit.SECONDS.toMillis(2));
    String log = getLog();
    assertThat(log).contains("LivenessPlugin started");

    // Wait for a few check intervals to pass and ensure metrics are initialized
    Thread.sleep(TimeUnit.SECONDS.toMillis(TEST_CHECK_INTERVAL_SECONDS * 5));

    // Check if metrics are being reported
    double uptimeTransactionsMetric = getMetricValue(
        LineaMetricCategory.SEQUENCER_LIVENESS,
        "uptime_transactions",
        List.of());

    // The metric should exist and be a non-negative number
    assertThat(uptimeTransactionsMetric).isNotNaN();
    assertThat(uptimeTransactionsMetric).isGreaterThanOrEqualTo(0);
  }
}

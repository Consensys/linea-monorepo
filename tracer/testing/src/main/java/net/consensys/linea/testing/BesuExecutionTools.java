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
package net.consensys.linea.testing;

import static org.assertj.core.api.Assertions.assertThat;
import static org.hyperledger.besu.tests.acceptance.dsl.WaitUtils.waitFor;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

import com.fasterxml.jackson.databind.JsonNode;
import io.netty.util.internal.ConcurrentSet;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.plugins.rpc.tracegeneration.TraceRequestParams;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.json.JsonConverter;
import okhttp3.Call;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions;
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode;
import org.hyperledger.besu.tests.acceptance.dsl.node.ThreadBesuNodeRunner;
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster;
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.ClusterConfigurationBuilder;
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeConfigurationBuilder;
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeFactory;
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.NodeConfigurationFactory;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.eth.EthTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions;

@Slf4j
public class BesuExecutionTools {

  private static final MediaType MEDIA_TYPE_JSON =
      MediaType.parse("application/json; charset=utf-8");
  private static final JsonConverter CONVERTER = JsonConverter.builder().build();

  private final ChainConfig chainConfig;
  private final OkHttpClient httpClient;
  private final BesuNode besuNode;
  private final Path tracesPath;
  private final List<Transaction> transactions;
  private final CorsetValidator corsetValidator;

  public BesuExecutionTools(
      ChainConfig chainConfig,
      Address coinbase,
      List<ToyAccount> accounts,
      List<Transaction> transactions) {
    this.httpClient = new OkHttpClient();
    this.chainConfig = chainConfig;
    GenesisConfigBuilder genesisConfigBuilder = new GenesisConfigBuilder();
    genesisConfigBuilder.setChainId(chainConfig.id);
    genesisConfigBuilder.setCoinbase(coinbase);
    genesisConfigBuilder.setGasLimit(chainConfig.gasLimitMaximum.longValue());
    accounts.forEach(genesisConfigBuilder::addAccount);
    try {
      this.tracesPath =
          Files.createTempDirectory(
              Path.of(System.getProperty("besu.traces.dir")), UUID.randomUUID().toString());
      this.besuNode = create(chainConfig, genesisConfigBuilder.buildAsString(), tracesPath);
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
    this.corsetValidator = new CorsetValidator(chainConfig);
    this.transactions = transactions;
  }

  private Call callRpcRequest(final String request) {
    return httpClient.newCall(
        new Request.Builder()
            .url(besuNode.jsonRpcBaseUrl().get())
            .post(RequestBody.create(request, MEDIA_TYPE_JSON))
            .build());
  }

  private static String createGeneratedConflatedFileV2Call(
      final long startBlockNumber, final long endBlockNumber) {
    TraceRequestParams traceRequestParams =
        new TraceRequestParams(startBlockNumber, endBlockNumber, "test");
    String params = CONVERTER.toJson(traceRequestParams);
    return "{\n"
        + "    \"jsonrpc\": \"2.0\",\n"
        + "    \"method\": \"linea_generateConflatedTracesToFileV2\",\n"
        + "    \"params\": ["
        + params
        + "],\n"
        + "    \"id\": 1\n"
        + "}";
  }

  private static BesuNode create(ChainConfig chainConfig, String genesisConfig, Path tracesPath)
      throws IOException {
    NodeConfigurationFactory node = new NodeConfigurationFactory();

    BesuNodeConfigurationBuilder besuNodeConfigurationBuilder =
        new BesuNodeConfigurationBuilder()
            .name("example-test-node")
            .genesisConfigProvider(nodes -> genesisConfig.describeConstable())
            .miningEnabled()
            .jsonRpcEnabled()
            .jsonRpcConfiguration(node.createJsonRpcWithRpcApiEnabledConfig("LINEA"))
            .requestedPlugins(
                List.of(
                    "TracerReadinessPlugin",
                    "TracesEndpointServicePlugin",
                    "LineCountsEndpointServicePlugin",
                    "CaptureEndpointServicePlugin"))
            .extraCLIOptions(
                List.of(
                    String.format(
                        "--plugin-linea-conflated-trace-generation-traces-output-path=%s",
                        tracesPath),
                    "--plugin-linea-rpc-concurrent-requests-limit=1",
                    String.format(
                        "--plugin-linea-l1l2-bridge-contract=%s",
                        chainConfig.bridgeConfiguration.contract().toHexString()),
                    String.format(
                        "--plugin-linea-l1l2-bridge-topic=%s",
                        chainConfig.bridgeConfiguration.topic().toHexString()),
                    "--plugin-linea-tracer-readiness-server-host=127.0.0.1",
                    "--plugin-linea-tracer-readiness-server-port=8548",
                    "--plugin-linea-tracer-readiness-max-blocks-behind=1"));
    return new BesuNodeFactory().create(besuNodeConfigurationBuilder.build());
  }

  public void executeTest() {
    try (Cluster cluster =
        new Cluster(
            new ClusterConfigurationBuilder().build(),
            new NetConditions(new NetTransactions()),
            new ThreadBesuNodeRunner())) {

      cluster.start(besuNode);

      EthTransactions ethTransactions = new EthTransactions();
      List<String> txHashes =
          transactions.stream()
              .map(
                  tx ->
                      besuNode.execute(
                          ethTransactions.sendRawTransaction(tx.encoded().toHexString())))
              .toList();
      ConcurrentSet<Long> blockNumbers = new ConcurrentSet<>();
      waitFor(
          10,
          () -> {
            txHashes.forEach(
                (txHash) -> {
                  var maybeTxReceipt =
                      besuNode.execute(ethTransactions.getTransactionReceipt(txHash));
                  assertThat(maybeTxReceipt).isPresent();
                  var txReceipt = maybeTxReceipt.get();
                  blockNumbers.add(txReceipt.getBlockNumber().longValue());
                  log.info(
                      "Example test txHash={}, blockNumber={}",
                      txReceipt.getTransactionHash(),
                      txReceipt.getBlockNumber());
                });
          });
      assertThat(blockNumbers).isNotEmpty();
      String request =
          createGeneratedConflatedFileV2Call(
              Collections.min(blockNumbers), Collections.max(blockNumbers));
      Response response = callRpcRequest(request).execute();
      String responseBody = response.body().string();
      assertThat(response.isSuccessful())
          .withFailMessage(
              String.format(
                  "Unexpected response code: %s, body: %s", response.code(), responseBody))
          .isTrue();
      JsonNode jsonRpcResponse = CONVERTER.fromJson(responseBody, JsonNode.class);
      Path traceFile =
          Path.of(jsonRpcResponse.get("result").get("conflatedTracesFileName").asText());
      waitFor(
          10,
          () -> {
            assertThat(traceFile.toFile().exists())
                .withFailMessage("Trace file %s does not exist", traceFile)
                .isTrue();
          });
      ExecutionEnvironment.checkTracer(traceFile, corsetValidator, false, Optional.of(log));
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }
}

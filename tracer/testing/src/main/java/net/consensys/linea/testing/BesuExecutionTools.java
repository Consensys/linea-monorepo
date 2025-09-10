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

import static net.consensys.linea.testing.ShomeiNode.MerkelProofResponse;
import static net.consensys.linea.zktracer.Fork.isPostParis;
import static org.assertj.core.api.Assertions.assertThat;
import static org.hyperledger.besu.tests.acceptance.dsl.WaitUtils.waitFor;

import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.net.ServerSocket;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.databind.node.ArrayNode;
import io.netty.util.internal.ConcurrentSet;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.plugins.rpc.tracegeneration.TraceFile;
import net.consensys.linea.plugins.rpc.tracegeneration.TraceRequestParams;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.json.JsonConverter;
import net.consensys.shomei.rpc.server.model.RollupGetZkEvmStateV0Parameter;
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
import org.hyperledger.besu.tests.acceptance.dsl.transaction.eth.EthTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions;
import org.junit.jupiter.api.TestInfo;
import org.web3j.protocol.core.DefaultBlockParameter;

@Slf4j
public class BesuExecutionTools {

  private static final MediaType MEDIA_TYPE_JSON =
      MediaType.parse("application/json; charset=utf-8");
  private static final JsonConverter CONVERTER = JsonConverter.builder().build();
  private static final ObjectMapper MAPPER =
      CONVERTER.getObjectMapper().enable(SerializationFeature.INDENT_OUTPUT);

  private final ChainConfig chainConfig;
  private final OkHttpClient httpClient;
  private final BesuNode besuNode;
  private final ShomeiNode shomeiNode;
  private final Path testDataDir;
  private final Path shomeiDataPath;
  private final List<Transaction> transactions;
  private final CorsetValidator corsetValidator;
  private final String testName;

  public BesuExecutionTools(
      String testName,
      ChainConfig chainConfig,
      Address coinbase,
      List<ToyAccount> accounts,
      List<Transaction> transactions) {
    String randomUUID = UUID.randomUUID().toString();
    String tmpTestName =
        Optional.ofNullable(testName)
            .map(name -> String.format("%s-%s", name, randomUUID))
            .orElse(randomUUID)
            .replaceAll("[,.:<>|*?\\r\\n\\[\\]() ]", "_");
    this.testName = tmpTestName.substring(0, Math.min(tmpTestName.length(), 200));
    int besuPort = findFreePort();
    int shomeiPort = findFreePort();
    this.httpClient = new OkHttpClient();
    this.chainConfig = chainConfig;
    // Generate file per fork in resources
    String genesisFileName = "BesuExecutionToolsGenesis_" + chainConfig.fork.name() + ".json";
    GenesisConfigBuilder genesisConfigBuilder = new GenesisConfigBuilder(genesisFileName);
    genesisConfigBuilder.setChainId(chainConfig.id);
    genesisConfigBuilder.setCoinbase(coinbase);
    genesisConfigBuilder.setGasLimit(chainConfig.gasLimitMaximum.longValue());
    accounts.forEach(genesisConfigBuilder::addAccount);
    try {
      this.testDataDir =
          Files.createDirectory(
              Path.of(System.getProperty("besu.traces.dir")).resolve(this.testName));
      this.shomeiDataPath = Files.createTempDirectory("shomei");
      this.besuNode =
          BesuNodeBuilder.create(
              testName,
              chainConfig.bridgeConfiguration,
              genesisConfigBuilder,
              besuPort,
              testDataDir,
              shomeiPort);

    } catch (IOException e) {
      throw new RuntimeException(e);
    }
    this.shomeiNode =
        new ShomeiNode.Builder()
            .setBesuRpcPort(besuPort)
            .setJsonRpcPort(shomeiPort)
            .setDataStoragePath(this.shomeiDataPath)
            .build();
    this.corsetValidator = new CorsetValidator(chainConfig);
    this.transactions = transactions;
  }

  private static String getTestName(Optional<TestInfo> testInfo) {
    return testInfo
        .map(
            info ->
                String.format(
                    "%s-%s-%s",
                    info.getTestClass().get().getSimpleName(),
                    info.getTestMethod().get().getName(),
                    info.getDisplayName()))
        .orElse(null);
  }

  public BesuExecutionTools(
      Optional<TestInfo> testInfo,
      ChainConfig chainConfig,
      Address coinbase,
      List<ToyAccount> accounts,
      List<Transaction> transactions) {
    this(getTestName(testInfo), chainConfig, coinbase, accounts, transactions);
  }

  public void executeTest() {
    Thread shomeiThread = new Thread(shomeiNode);
    Cluster besuCluster =
        new Cluster(
            new ClusterConfigurationBuilder().build(),
            new NetConditions(new NetTransactions()),
            new ThreadBesuNodeRunner());
    try {
      shomeiThread.start();
      besuCluster.start(besuNode);

      // Send transaction to the transaction pool with eth_sendRawTransaction
      EthTransactions ethTransactions = new EthTransactions();
      List<String> txHashes =
          transactions.stream()
              .map(
                  tx ->
                      besuNode.execute(
                          ethTransactions.sendRawTransaction(tx.encoded().toHexString())))
              .toList();
      Map<String, Boolean> txReceiptProcessed = new HashMap<>();
      ConcurrentSet<Long> blockNumbers = new ConcurrentSet<>();

      // If fork is Paris or after, Clique as a consensus layer defined in the genesis file
      // doesn't work anymore
      // We use EngineAPIService to mimick the consensus layer steps and build a new block
      if (isPostParis(chainConfig.fork)) {
        ObjectMapper mapper = new ObjectMapper();
        EngineAPIService engineApiService = new EngineAPIService(besuNode, ethTransactions, mapper);
        var latestTimestamp = this.besuNode.execute(ethTransactions.block()).getTimestamp();
        // TODO: could be done with genesis
        engineApiService.buildNewBlock(chainConfig.fork, latestTimestamp.longValue() + 1L, 1000);
      }

      // We check that the transactions are included in a block
      waitFor(
          100,
          () -> {
            txHashes.forEach(
                (txHash) -> {
                  if (txReceiptProcessed.containsKey(txHash)) {
                    return;
                  }
                  var maybeTxReceipt =
                      besuNode.execute(ethTransactions.getTransactionReceipt(txHash));
                  assertThat(maybeTxReceipt).isPresent();
                  var txReceipt = maybeTxReceipt.get();
                  blockNumbers.add(txReceipt.getBlockNumber().longValue());
                  txReceiptProcessed.put(txHash, true);
                  log.info(
                      "Executed transaction txHash={}, blockNumber={}",
                      txReceipt.getTransactionHash(),
                      txReceipt.getBlockNumber());
                });
          });

      assertThat(blockNumbers).isNotEmpty();
      long startBlockNumber = Collections.min(blockNumbers);
      long endBlockNumber = Collections.max(blockNumbers);
      String previousBlockStateRoot =
          besuNode
              .execute(
                  ethTransactions.block(
                      DefaultBlockParameter.valueOf(BigInteger.valueOf(startBlockNumber - 1))))
              .getStateRoot();
      TraceFile traceFile = lineaGenerateConflatedTracesToFileV2(startBlockNumber, endBlockNumber);
      Path traceFilePath = Path.of(traceFile.conflatedTracesFileName());
      waitFor(
          10,
          () -> {
            assertThat(traceFilePath.toFile().exists())
                .withFailMessage("Trace file %s does not exist", traceFilePath)
                .isTrue();
          });

      ExecutionEnvironment.checkTracer(traceFilePath, corsetValidator, false, Optional.of(log));
      MerkelProofResponse merkelProofResponse =
          rollupGetZkEVMStateMerkleProofV0(startBlockNumber, endBlockNumber);
      log.info("rollupGetZkEVMStateMerkleProofV0={}", merkelProofResponse);

      ExecutionProof.BatchExecutionProofRequestDto executionProofRequestDto =
          new ExecutionProof.BatchExecutionProofRequestDto(
              merkelProofResponse.zkParentStateRootHash(),
              previousBlockStateRoot,
              traceFilePath.getFileName().toString(),
              traceFile.tracesEngineVersion(),
              merkelProofResponse.zkStateManagerVersion(),
              merkelProofResponse.zkStateMerkleProof(),
              Collections.emptyList() /* blocksData */);

      String executionProofFileName =
          ExecutionProof.getExecutionProofRequestFilename(
              startBlockNumber,
              endBlockNumber,
              traceFile.tracesEngineVersion(),
              merkelProofResponse.zkStateManagerVersion());
      File executionProofRequestFile = testDataDir.resolve(executionProofFileName).toFile();
      assertThat(executionProofRequestFile.createNewFile()).isTrue();
      MAPPER.writeValue(executionProofRequestFile, executionProofRequestDto);
    } catch (IOException | InterruptedException e) {
      throw new RuntimeException(e);
    } finally {
      try {
        besuCluster.close();
      } catch (Exception e) {
        log.error("Error closing besu cluster: %s".formatted(e.getMessage()), e);
      }

      try {
        shomeiNode.close();
      } catch (Exception e) {
        log.error("Error closing shomei node: %s".formatted(e.getMessage()), e);
      }
      try {
        besuNode.close();
      } catch (Exception e) {
        log.error("Error closing besu node: %s".formatted(e.getMessage()), e);
      }
    }
  }

  private TraceFile lineaGenerateConflatedTracesToFileV2(
      final long startBlockNumber, final long endBlockNumber) throws IOException {
    return jsonRpcRequest(
        besuNode.jsonRpcBaseUrl().get(),
        "linea_generateConflatedTracesToFileV2",
        new TraceRequestParams(startBlockNumber, endBlockNumber, "test"),
        TraceFile.class);
  }

  private MerkelProofResponse rollupGetZkEVMStateMerkleProofV0(
      final long startBlockNumber, final long endBlockNumber) throws IOException {
    return jsonRpcRequest(
        shomeiNode.getJsonRpcUrl(),
        "rollup_getZkEVMStateMerkleProofV0",
        new RollupGetZkEvmStateV0Parameter(
            String.valueOf(startBlockNumber), String.valueOf(endBlockNumber), "test"),
        MerkelProofResponse.class);
  }

  private <R, P> R jsonRpcRequest(
      final String jsonRpcUrl, final String method, final P params, final Class<R> responseType)
      throws IOException {
    String request = createJsonRpcRequest(method, params, "1");
    Response response =
        httpClient
            .newCall(
                new Request.Builder()
                    .url(jsonRpcUrl)
                    .post(RequestBody.create(request, MEDIA_TYPE_JSON))
                    .build())
            .execute();
    String responseBody = response.body().string();
    assertThat(response.isSuccessful())
        .withFailMessage(
            String.format(
                "Unexpected response code: %s, body: %s, request: %s",
                response.code(), responseBody, request))
        .isTrue();
    JsonNode jsonRpcResponse = CONVERTER.fromJson(responseBody, JsonNode.class);
    JsonNode result = jsonRpcResponse.get("result");
    assertThat(result)
        .withFailMessage(
            String.format(
                "Request failed. response code: %s, body: %s, request: %s",
                response.code(), responseBody, request))
        .isNotNull();
    return parseJsonRpcResult(result, responseType);
  }

  private static <P> String createJsonRpcRequest(
      final String method, final P params, final String id) {
    return String.format(
        "{\n"
            + "    \"jsonrpc\": \"2.0\",\n"
            + "    \"method\": \"%s\",\n"
            + "    \"params\": [%s],\n"
            + "    \"id\": %s\n"
            + "}",
        method, CONVERTER.toJson(params), id);
  }

  @SuppressWarnings("unchecked")
  private static <R> R parseJsonRpcResult(final JsonNode result, final Class<R> responseType)
      throws JsonProcessingException {
    if (responseType.equals(MerkelProofResponse.class)) {
      return (R)
          new MerkelProofResponse(
              result.get("zkParentStateRootHash").asText(),
              result.get("zkEndStateRootHash").asText(),
              (ArrayNode) result.get("zkStateMerkleProof"),
              "test");
    } else {
      return MAPPER.treeToValue(result, responseType);
    }
  }

  private static int findFreePort() {
    int port = 0;
    try (ServerSocket socket = new ServerSocket(0)) {
      socket.setReuseAddress(true);
      port = socket.getLocalPort();
    } catch (IOException ignored) {
    }
    if (port > 0) {
      return port;
    }
    throw new RuntimeException("Could not find a free port");
  }
}

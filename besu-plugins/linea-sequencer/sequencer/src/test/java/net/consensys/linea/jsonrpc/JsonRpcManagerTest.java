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

package net.consensys.linea.jsonrpc;

import static com.github.tomakehurst.wiremock.client.WireMock.aResponse;
import static com.github.tomakehurst.wiremock.client.WireMock.equalToJson;
import static com.github.tomakehurst.wiremock.client.WireMock.exactly;
import static com.github.tomakehurst.wiremock.client.WireMock.post;
import static com.github.tomakehurst.wiremock.client.WireMock.postRequestedFor;
import static com.github.tomakehurst.wiremock.client.WireMock.stubFor;
import static com.github.tomakehurst.wiremock.client.WireMock.urlEqualTo;
import static com.github.tomakehurst.wiremock.client.WireMock.verify;
import static java.util.concurrent.TimeUnit.SECONDS;
import static org.assertj.core.api.Assertions.assertThat;
import static org.awaitility.Awaitility.await;
import static org.mockito.Mockito.when;

import java.io.IOException;
import java.net.URI;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;
import java.util.stream.Stream;

import com.github.tomakehurst.wiremock.http.Fault;
import com.github.tomakehurst.wiremock.junit5.WireMockRuntimeInfo;
import com.github.tomakehurst.wiremock.junit5.WireMockTest;
import com.github.tomakehurst.wiremock.stubbing.Scenario;
import net.consensys.linea.sequencer.txselection.selectors.TestTransactionEvaluationContext;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.PendingTransaction;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.data.TransactionSelectionResult;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@WireMockTest
@ExtendWith(MockitoExtension.class)
class JsonRpcManagerTest {
  @TempDir private Path tempDataDir;
  private JsonRpcManager jsonRpcManager;
  private final Bytes randomEncodedBytes = Bytes.random(32);
  @Mock private PendingTransaction pendingTransaction;
  @Mock private ProcessableBlockHeader pendingBlockHeader;
  @Mock private Transaction transaction;

  @BeforeEach
  void init(final WireMockRuntimeInfo wmInfo) {
    // mock stubbing
    when(pendingBlockHeader.getNumber()).thenReturn(1L);
    when(pendingTransaction.getTransaction()).thenReturn(transaction);
    when(transaction.encoded()).thenReturn(randomEncodedBytes);

    jsonRpcManager = new JsonRpcManager(tempDataDir, URI.create(wmInfo.getHttpBaseUrl()));
    jsonRpcManager.start();
  }

  @AfterEach
  void cleanup() {
    jsonRpcManager.shutdown();
  }

  @Test
  void rejectedTxIsReported() throws InterruptedException {
    // json-rpc stubbing
    stubFor(
        post(urlEqualTo("/"))
            .willReturn(
                aResponse()
                    .withStatus(200)
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        "{\"jsonrpc\":\"2.0\",\"result\":{ \"status\": \"SAVED\"},\"id\":1}")));

    final TestTransactionEvaluationContext context =
        new TestTransactionEvaluationContext(pendingBlockHeader, pendingTransaction);
    final TransactionSelectionResult result = TransactionSelectionResult.invalid("test");
    final Instant timestamp = Instant.now();

    // method under test
    final String jsonRpcCall =
        JsonRpcRequestBuilder.buildRejectedTxRequest(context, result, timestamp);
    jsonRpcManager.submitNewJsonRpcCall(jsonRpcCall);

    // Use Awaitility to wait for the condition to be met
    await()
        .atMost(2, SECONDS)
        .untilAsserted(
            () ->
                verify(
                    exactly(1),
                    postRequestedFor(urlEqualTo("/")).withRequestBody(equalToJson(jsonRpcCall))));
  }

  @Test
  void firstCallErrorSecondCallSuccessScenario() throws InterruptedException, IOException {
    stubFor(
        post(urlEqualTo("/"))
            .inScenario("RPC Calls")
            .whenScenarioStateIs(Scenario.STARTED)
            .willReturn(
                aResponse()
                    .withStatus(500)
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32000,\"message\":\"Internal error\"},\"id\":1}"))
            .willSetStateTo("Second Call"));

    stubFor(
        post(urlEqualTo("/"))
            .inScenario("RPC Calls")
            .whenScenarioStateIs("Second Call")
            .willReturn(
                aResponse()
                    .withStatus(200)
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        "{\"jsonrpc\":\"2.0\",\"result\":{ \"status\": \"SAVED\"},\"id\":1}")));

    // Prepare test data
    final TestTransactionEvaluationContext context =
        new TestTransactionEvaluationContext(pendingBlockHeader, pendingTransaction);
    final TransactionSelectionResult result = TransactionSelectionResult.invalid("test");
    final Instant timestamp = Instant.now();

    // Generate JSON-RPC call
    final String jsonRpcCall =
        JsonRpcRequestBuilder.buildRejectedTxRequest(context, result, timestamp);

    // Submit the call, the scheduler will retry the failed call
    jsonRpcManager.submitNewJsonRpcCall(jsonRpcCall);

    // Use Awaitility to wait for the condition to be met
    await()
        .atMost(2, SECONDS)
        .untilAsserted(
            () ->
                verify(
                    exactly(2),
                    postRequestedFor(urlEqualTo("/")).withRequestBody(equalToJson(jsonRpcCall))));

    // Verify that the JSON file no longer exists in the directory (as the second call was
    // successful)
    Path rejTxRpcDir = tempDataDir.resolve("rej_tx_rpc");
    try (Stream<Path> files = Files.list(rejTxRpcDir)) {
      long fileCount = files.filter(path -> path.toString().endsWith(".json")).count();
      assertThat(fileCount).isEqualTo(0);
    }
  }

  @Test
  void serverRespondingWithErrorScenario() throws InterruptedException, IOException {
    // Stub for error response
    stubFor(
        post(urlEqualTo("/"))
            .willReturn(
                aResponse()
                    .withStatus(500)
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32000,\"message\":\"Internal error\"},\"id\":1}")));

    // Prepare test data
    final TestTransactionEvaluationContext context =
        new TestTransactionEvaluationContext(pendingBlockHeader, pendingTransaction);
    final TransactionSelectionResult result = TransactionSelectionResult.invalid("test");
    final Instant timestamp = Instant.now();

    // Generate JSON-RPC call
    final String jsonRpcCall =
        JsonRpcRequestBuilder.buildRejectedTxRequest(context, result, timestamp);

    // Submit the call
    jsonRpcManager.submitNewJsonRpcCall(jsonRpcCall);

    // Use Awaitility to wait for the condition to be met
    await()
        .atMost(2, SECONDS)
        .untilAsserted(
            () ->
                verify(
                    exactly(1),
                    postRequestedFor(urlEqualTo("/")).withRequestBody(equalToJson(jsonRpcCall))));

    // Verify that the JSON file still exists in the directory (as the call was unsuccessful)
    final Path rejTxRpcDir = tempDataDir.resolve("rej_tx_rpc");
    try (Stream<Path> files = Files.list(rejTxRpcDir)) {
      long fileCount = files.filter(path -> path.toString().endsWith(".json")).count();
      assertThat(fileCount).as("JSON file should exist as server responded with error").isOne();
    }
  }

  @Test
  void firstTwoCallsErrorThenLastCallSuccessScenario() throws InterruptedException, IOException {
    stubFor(
        post(urlEqualTo("/"))
            .inScenario("RPC Calls")
            .whenScenarioStateIs(Scenario.STARTED)
            .willReturn(aResponse().withFault(Fault.MALFORMED_RESPONSE_CHUNK))
            .willSetStateTo("Second Call"));

    stubFor(
        post(urlEqualTo("/"))
            .inScenario("RPC Calls")
            .whenScenarioStateIs("Second Call")
            .willReturn(
                aResponse()
                    .withStatus(500)
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        "{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32000,\"message\":\"Internal error\"},\"id\":1}"))
            .willSetStateTo("Third Call"));

    stubFor(
        post(urlEqualTo("/"))
            .inScenario("RPC Calls")
            .whenScenarioStateIs("Third Call")
            .willReturn(
                aResponse()
                    .withStatus(200)
                    .withHeader("Content-Type", "application/json")
                    .withBody(
                        "{\"jsonrpc\":\"2.0\",\"result\":{ \"status\": \"SAVED\"},\"id\":1}")));

    // Prepare test data
    final TestTransactionEvaluationContext context =
        new TestTransactionEvaluationContext(pendingBlockHeader, pendingTransaction);
    final TransactionSelectionResult result = TransactionSelectionResult.invalid("test");
    final Instant timestamp = Instant.now();

    // Generate JSON-RPC call
    final String jsonRpcCall =
        JsonRpcRequestBuilder.buildRejectedTxRequest(context, result, timestamp);

    // Submit the call, the scheduler will retry the failed calls
    jsonRpcManager.submitNewJsonRpcCall(jsonRpcCall);

    // Use Awaitility to wait for the condition to be met
    await()
        .atMost(6, SECONDS)
        .untilAsserted(
            () ->
                verify(
                    exactly(3),
                    postRequestedFor(urlEqualTo("/")).withRequestBody(equalToJson(jsonRpcCall))));

    // Verify that the JSON file no longer exists in the directory (as the second call was
    // successful)
    Path rejTxRpcDir = tempDataDir.resolve("rej_tx_rpc");
    try (Stream<Path> files = Files.list(rejTxRpcDir)) {
      long fileCount = files.filter(path -> path.toString().endsWith(".json")).count();
      assertThat(fileCount).isEqualTo(0);
    }
  }
}

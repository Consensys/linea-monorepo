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

package net.consensys.linea.plugins.rpc.tracegeneration;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.Optional;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.plugins.rpc.RequestLimiter;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SignatureAlgorithm;
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.StateOverrideMap;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput;
import org.hyperledger.besu.plugin.data.BlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.BlockOverrides;
import org.hyperledger.besu.plugin.data.PluginBlockSimulationResult;
import org.hyperledger.besu.plugin.services.BlockSimulationService;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

/**
 * Integration tests for GenerateVirtualBlockConflatedTracesV1. These tests verify that the RPC
 * endpoint can properly handle requests and output trace files to the filesystem.
 */
class GenerateVirtualBlockConflatedTracesV1IntegrationTest {

  private static final long BLOCK_NUMBER = 100L;
  private static final long PARENT_BLOCK_NUMBER = 99L;
  private static final BigInteger CHAIN_ID = BigInteger.valueOf(59144L); // Linea mainnet

  @TempDir Path tempDir;

  private BlockSimulationService blockSimulationService;
  private BlockchainService blockchainService;
  private PluginRpcRequest rpcRequest;
  private BlockContext parentBlockContext;
  private BlockHeader parentBlockHeader;
  private PluginBlockSimulationResult simulationResult;
  private BlockHeader simulatedBlockHeader;

  private GenerateVirtualBlockConflatedTracesV1 rpcMethod;
  private TracesEndpointConfiguration endpointConfiguration;
  private LineaL1L2BridgeSharedConfiguration bridgeConfiguration;
  private RequestLimiter requestLimiter;

  @BeforeEach
  void setUp() {
    // Create mocks
    blockSimulationService = mock(BlockSimulationService.class);
    blockchainService = mock(BlockchainService.class);
    rpcRequest = mock(PluginRpcRequest.class);
    parentBlockContext = mock(BlockContext.class);
    parentBlockHeader = mock(BlockHeader.class);
    simulationResult = mock(PluginBlockSimulationResult.class);
    simulatedBlockHeader = mock(BlockHeader.class);

    endpointConfiguration =
        TracesEndpointConfiguration.builder()
            .tracesOutputPath(tempDir.toString())
            .caching(true)
            .traceCompression(false)
            .traceFileVersion(1)
            .build();

    bridgeConfiguration = LineaL1L2BridgeSharedConfiguration.builder().build();

    // Request limiter with 1 concurrent request allowed
    requestLimiter = RequestLimiter.builder().concurrentRequestsCount(1).build();

    rpcMethod =
        new GenerateVirtualBlockConflatedTracesV1(
            requestLimiter,
            endpointConfiguration,
            bridgeConfiguration,
            blockSimulationService,
            blockchainService);
  }

  @Test
  void namespaceAndMethodNameAreCorrect() {
    assertThat(rpcMethod.getNamespace()).isEqualTo("linea");
    assertThat(rpcMethod.getName()).isEqualTo("generateVirtualBlockConflatedTracesToFileV1");
  }

  @Test
  void executesSuccessfullyAndWritesTraceFile() throws IOException {
    // Prepare a signed transaction
    String signedTxRlp = createSignedTransactionRlp();

    // Setup request params
    VirtualBlockTraceRequestParams params =
        new VirtualBlockTraceRequestParams(BLOCK_NUMBER, new String[] {signedTxRlp});
    when(rpcRequest.getParams()).thenReturn(new Object[] {params});

    // Mock blockchain service
    when(blockchainService.getBlockByNumber(PARENT_BLOCK_NUMBER))
        .thenReturn(Optional.of(parentBlockContext));
    when(blockchainService.getBlockByNumber(anyLong())).thenReturn(Optional.of(parentBlockContext));
    when(blockchainService.getChainId()).thenReturn(Optional.of(CHAIN_ID));

    // Mock parent block
    when(parentBlockContext.getBlockHeader()).thenReturn(parentBlockHeader);
    when(parentBlockHeader.getTimestamp()).thenReturn(1000L);
    when(parentBlockHeader.getBlockHash()).thenReturn(Hash.ZERO);

    // Mock simulation result
    when(simulationResult.getBlockHeader()).thenReturn(simulatedBlockHeader);
    when(simulatedBlockHeader.getBlockHash()).thenReturn(Hash.ZERO);

    // Mock simulation - when simulate is called with tracer, return the result
    // Note: This uses the 5-parameter version from Besu PR #9708
    when(blockSimulationService.simulate(
            eq(PARENT_BLOCK_NUMBER),
            any(List.class),
            any(BlockOverrides.class),
            any(StateOverrideMap.class),
            any()))
        .thenReturn(simulationResult);

    // Execute the RPC method
    TraceFile result = rpcMethod.execute(rpcRequest);

    // Verify result
    assertThat(result).isNotNull();
    assertThat(result.tracesEngineVersion()).isNotNull();
    assertThat(result.conflatedTracesFileName()).isNotNull();

    // Verify file was created
    Path traceFilePath = Path.of(result.conflatedTracesFileName());
    assertThat(Files.exists(traceFilePath))
        .as("Trace file should exist at: %s", traceFilePath)
        .isTrue();

    // Verify file name follows the convention: blockNumber-.conflated.version.lt
    String fileName = traceFilePath.getFileName().toString();
    assertThat(fileName).startsWith(BLOCK_NUMBER + "-.conflated.");
    assertThat(fileName).endsWith(".lt");
  }

  @Test
  void throwsExceptionWhenParentBlockNotFound() {
    String signedTxRlp = createSignedTransactionRlp();

    VirtualBlockTraceRequestParams params =
        new VirtualBlockTraceRequestParams(BLOCK_NUMBER, new String[] {signedTxRlp});
    when(rpcRequest.getParams()).thenReturn(new Object[] {params});

    // Parent block not found
    when(blockchainService.getBlockByNumber(PARENT_BLOCK_NUMBER)).thenReturn(Optional.empty());

    assertThatThrownBy(() -> rpcMethod.execute(rpcRequest))
        .isInstanceOf(PluginRpcEndpointException.class)
        .hasMessageContaining("BLOCK_MISSING_IN_CHAIN");
  }

  @Test
  void returnsCachedFileWhenAvailable() throws IOException {
    // First, create a cached file
    String tracesEngineVersion = TraceRequestParams.getTracerRuntime();
    if (tracesEngineVersion == null) {
      tracesEngineVersion = "test-version";
    }
    String cachedFileName =
        String.format("%d-.conflated.%s.lt", BLOCK_NUMBER, tracesEngineVersion);
    Path cachedFilePath = tempDir.resolve(cachedFileName);
    Files.writeString(cachedFilePath, "cached trace content");

    // Setup request
    String signedTxRlp = createSignedTransactionRlp();
    VirtualBlockTraceRequestParams params =
        new VirtualBlockTraceRequestParams(BLOCK_NUMBER, new String[] {signedTxRlp});
    when(rpcRequest.getParams()).thenReturn(new Object[] {params});

    // Note: blockchainService and blockSimulationService should NOT be called
    // when cache hit occurs

    // Execute
    TraceFile result = rpcMethod.execute(rpcRequest);

    // Verify it returned the cached file
    assertThat(result).isNotNull();
    assertThat(result.conflatedTracesFileName()).isEqualTo(cachedFilePath.toString());
  }

  @Test
  void validatesBlockNumberMustBeAtLeastOne() {
    VirtualBlockTraceRequestParams params =
        new VirtualBlockTraceRequestParams(0L, new String[] {"0x1234"});
    when(rpcRequest.getParams()).thenReturn(new Object[] {params});

    assertThatThrownBy(() -> rpcMethod.execute(rpcRequest))
        .hasMessageContaining("INVALID_BLOCK_NUMBER");
  }

  @Test
  void validatesTransactionsMustNotBeEmpty() {
    VirtualBlockTraceRequestParams params =
        new VirtualBlockTraceRequestParams(BLOCK_NUMBER, new String[] {});
    when(rpcRequest.getParams()).thenReturn(new Object[] {params});

    assertThatThrownBy(() -> rpcMethod.execute(rpcRequest))
        .hasMessageContaining("INVALID_TRANSACTIONS");
  }

  /**
   * Creates a valid signed transaction RLP for testing. This creates a minimal signed transaction
   * that can be decoded by the RPC endpoint.
   */
  private String createSignedTransactionRlp() {
    SignatureAlgorithm signatureAlgorithm = SignatureAlgorithmFactory.getInstance();
    // Use a well-known test private key
    BigInteger privateKeyValue =
        new BigInteger("8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63", 16);
    KeyPair keyPair =
        signatureAlgorithm.createKeyPair(signatureAlgorithm.createPrivateKey(privateKeyValue));

    Transaction tx =
        Transaction.builder()
            .type(org.hyperledger.besu.datatypes.TransactionType.FRONTIER)
            .nonce(0)
            .gasPrice(Wei.of(1000000000L))
            .gasLimit(21000)
            .to(Address.fromHexString("0x0000000000000000000000000000000000000001"))
            .value(Wei.of(1))
            .payload(Bytes.EMPTY)
            .chainId(CHAIN_ID)
            .signAndBuild(keyPair);

    BytesValueRLPOutput rlpOutput = new BytesValueRLPOutput();
    tx.writeTo(rlpOutput);
    return rlpOutput.encoded().toHexString();
  }
}

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

import static net.consensys.linea.zktracer.Fork.getForkFromBesuBlockchainService;
import static net.consensys.linea.zktracer.types.PublicInputs.defaultEmptyHistoricalBlockhashes;

import com.google.common.base.Stopwatch;
import java.math.BigInteger;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.plugins.BesuServiceProvider;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.plugins.rpc.RequestLimiter;
import net.consensys.linea.plugins.rpc.Validator;
import net.consensys.linea.tracewriter.TraceWriter;
import net.consensys.linea.zktracer.Fork;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.json.JsonConverter;
import net.consensys.linea.zktracer.types.PublicInputs;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.StateOverrideMap;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.ethereum.api.util.DomainObjectDecodeUtils;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.data.BlockContext;
import org.hyperledger.besu.plugin.data.BlockOverrides;
import org.hyperledger.besu.plugin.data.PluginBlockSimulationResult;
import org.hyperledger.besu.plugin.services.BlockSimulationService;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.rpc.RpcMethodError;

/**
 * RPC endpoint for generating conflated traces for virtual blocks. This is used for invalidity
 * proof generation for BadPrecompile and TooManyLogs forced transaction rejection scenarios.
 *
 * <p>The endpoint simulates execution of transactions on top of a past chain state (blockNumber -
 * 1) and generates execution traces that can be used for ZK proof generation.
 *
 * <p><b>IMPORTANT:</b> This feature requires Besu with PR #9708 merged (adds tracer support to
 * BlockSimulationService). See: https://github.com/hyperledger/besu/commit/2cfe3320fa5ef00d1d4acc49e9be0909be10393f
 */
@Slf4j
public class GenerateVirtualBlockConflatedTracesV1 {
  private static final JsonConverter CONVERTER = JsonConverter.builder().build();

  private final int traceFileVersion;
  private final RequestLimiter requestLimiter;
  private final TraceWriter traceWriter;
  private final ServiceManager besuContext;
  private final LineaL1L2BridgeSharedConfiguration l1L2BridgeSharedConfiguration;
  private final BlockSimulationService blockSimulationService;

  public GenerateVirtualBlockConflatedTracesV1(
      final ServiceManager besuContext,
      final RequestLimiter requestLimiter,
      final TracesEndpointConfiguration endpointConfiguration,
      final LineaL1L2BridgeSharedConfiguration lineaL1L2BridgeSharedConfiguration,
      final BlockSimulationService blockSimulationService) {
    this.besuContext = besuContext;
    this.requestLimiter = requestLimiter;
    this.traceWriter =
        new TraceWriter(
            Paths.get(endpointConfiguration.tracesOutputPath()),
            endpointConfiguration.traceCompression());
    this.l1L2BridgeSharedConfiguration = lineaL1L2BridgeSharedConfiguration;
    this.traceFileVersion = endpointConfiguration.traceFileVersion();
    this.blockSimulationService = blockSimulationService;
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "generateVirtualBlockConflatedTracesToFileV1";
  }

  /**
   * Handles virtual block execution traces generation.
   *
   * @param request holds parameters of the RPC request.
   * @return an execution file trace.
   */
  public TraceFile execute(final PluginRpcRequest request) {
    return requestLimiter.execute(request, this::generateVirtualBlockTraceFile);
  }

  private TraceFile generateVirtualBlockTraceFile(PluginRpcRequest request) {
    final Stopwatch sw = Stopwatch.createStarted();

    final Object[] rawParams = request.getParams();
    Validator.validatePluginRpcRequestParams(rawParams);

    final VirtualBlockTraceRequestParams params =
        CONVERTER.fromJson(CONVERTER.toJson(rawParams[0]), VirtualBlockTraceRequestParams.class);

    params.validate();

    final long blockNumber = params.blockNumber();
    final long parentBlockNumber = blockNumber - 1;

    // Get blockchain service and validate parent block exists
    final BlockchainService blockchainService =
        BesuServiceProvider.getBesuService(besuContext, BlockchainService.class);

    final BlockContext parentBlock =
        blockchainService
            .getBlockByNumber(parentBlockNumber)
            .orElseThrow(
                () ->
                    new PluginRpcEndpointException(
                        new BlockMissingError(parentBlockNumber, blockNumber)));

    log.info(
        "[VIRTUAL_BLOCK_TRACING] Generating traces for virtual block {} on top of parent block {}",
        blockNumber,
        parentBlockNumber);

    // Decode RLP transactions
    final List<Transaction> transactions = decodeTransactions(params.txsRlpEncoded());
    log.debug(
        "[VIRTUAL_BLOCK_TRACING] Decoded {} transactions for virtual block {}",
        transactions.size(),
        blockNumber);

    // Get fork for the virtual block
    final Fork fork = getForkFromBesuBlockchainService(blockchainService, blockNumber, blockNumber);
    final BigInteger chainId =
        blockchainService
            .getChainId()
            .orElseThrow(() -> new IllegalStateException("ChainId must be provided"));

    // Create public inputs for the virtual block (single block conflation)
    final PublicInputs publicInputs = defaultEmptyHistoricalBlockhashes(blockNumber, blockNumber);

    // Create ZkTracer
    final ZkTracer tracer =
        new ZkTracer(fork, l1L2BridgeSharedConfiguration, chainId, publicInputs);
    tracer.setLtFileMajorVersion(traceFileVersion);

    // Build block overrides for the virtual block
    final BlockOverrides blockOverrides =
        BlockOverrides.builder()
            .blockNumber(blockNumber)
            .timestamp(parentBlock.getBlockHeader().getTimestamp() + 1)
            .blockHashLookup(this::getBlockHashByNumber)
            .build();

    // Start conflation for single block
    tracer.traceStartConflation(1);

    PluginBlockSimulationResult simulationResult;
    try {
      // Simulate the virtual block with our tracer
      // Note: This requires Besu PR #9708 which adds tracer support to BlockSimulationService
      // See: https://github.com/hyperledger/besu/commit/2cfe3320fa5ef00d1d4acc49e9be0909be10393f
      simulationResult =
          blockSimulationService.simulate(
              parentBlockNumber, transactions, blockOverrides, new StateOverrideMap(), tracer);
    } finally {
      // End conflation - use empty WorldView as we don't have direct access to post-simulation state
      // The tracer captures state during execution via the OperationTracer callbacks
      tracer.traceEndConflation(EmptyWorldView.INSTANCE);
    }

    log.info(
        "[VIRTUAL_BLOCK_TRACING] Virtual block {} simulation completed in {} (simulated block hash: {})",
        blockNumber,
        sw,
        simulationResult.getBlockHeader().getBlockHash());
    sw.reset().start();

    // Get tracer runtime version
    final String tracesEngineVersion = VirtualBlockTraceRequestParams.getTracerRuntime();

    // Write trace file with virtual block naming convention
    final Path path =
        traceWriter.writeVirtualBlockTraceToFile(tracer, blockNumber, tracesEngineVersion);

    log.info(
        "[VIRTUAL_BLOCK_TRACING] Trace for virtual block {} serialized to {} in {}",
        blockNumber,
        path,
        sw);

    return new TraceFile(tracesEngineVersion, path.toString());
  }

  private List<Transaction> decodeTransactions(String[] txsRlpEncoded) {
    final List<Transaction> transactions = new ArrayList<>(txsRlpEncoded.length);
    for (int i = 0; i < txsRlpEncoded.length; i++) {
      try {
        final Transaction tx = DomainObjectDecodeUtils.decodeRawTransaction(txsRlpEncoded[i]);
        transactions.add(tx);
      } catch (Exception e) {
        throw new IllegalArgumentException(
            "Failed to decode transaction at index " + i + ": " + e.getMessage(), e);
      }
    }
    return transactions;
  }

  private Hash getBlockHashByNumber(long blockNumber) {
    final BlockchainService blockchainService =
        BesuServiceProvider.getBesuService(besuContext, BlockchainService.class);
    return blockchainService
        .getBlockByNumber(blockNumber)
        .map(block -> block.getBlockHeader().getBlockHash())
        .orElse(Hash.ZERO);
  }

  /** Error returned when the parent block (blockNumber - 1) is not found in the chain. */
  static class BlockMissingError implements RpcMethodError {
    private static final int BLOCK_MISSING_ERROR_CODE = -32001;
    private final long parentBlockNumber;
    private final long requestedBlockNumber;

    BlockMissingError(long parentBlockNumber, long requestedBlockNumber) {
      this.parentBlockNumber = parentBlockNumber;
      this.requestedBlockNumber = requestedBlockNumber;
    }

    @Override
    public int getCode() {
      return BLOCK_MISSING_ERROR_CODE;
    }

    @Override
    public String getMessage() {
      return "BLOCK_MISSING_IN_CHAIN: Parent block %d not found (required for virtual block %d)"
          .formatted(parentBlockNumber, requestedBlockNumber);
    }
  }

  /**
   * Empty WorldView implementation for traceEndConflation. The actual state is captured during
   * execution via OperationTracer callbacks from BlockSimulationService.
   */
  private enum EmptyWorldView implements WorldView {
    INSTANCE;

    @Override
    public org.hyperledger.besu.evm.account.Account get(
        final org.hyperledger.besu.datatypes.Address address) {
      return null;
    }
  }
}

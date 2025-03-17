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

import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Optional;

import com.google.common.base.Stopwatch;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.plugins.BesuServiceProvider;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.plugins.rpc.RequestLimiter;
import net.consensys.linea.plugins.rpc.Validator;
import net.consensys.linea.tracewriter.TraceWriter;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.json.JsonConverter;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/**
 * Sets up an RPC endpoint for generating conflated file trace. This class provides an RPC endpoint
 * named 'generateConflatedTracesToFileV0' under the 'rollup' namespace. When this endpoint is
 * called, it triggers the execution of the 'execute' method, which generates conflated file traces
 * based on the provided request parameters and writes them to a file.
 */
@Slf4j
public class GenerateConflatedTracesV2 {
  private static final JsonConverter CONVERTER = JsonConverter.builder().build();

  private final RequestLimiter requestLimiter;

  private final Path tracesOutputPath;
  private final ServiceManager besuContext;
  private TraceService traceService;
  private final LineaL1L2BridgeSharedConfiguration l1L2BridgeSharedConfiguration;

  public GenerateConflatedTracesV2(
      final ServiceManager besuContext,
      final RequestLimiter requestLimiter,
      final TracesEndpointConfiguration endpointConfiguration,
      LineaL1L2BridgeSharedConfiguration lineaL1L2BridgeSharedConfiguration) {
    this.besuContext = besuContext;
    this.requestLimiter = requestLimiter;
    this.tracesOutputPath = Paths.get(endpointConfiguration.tracesOutputPath());
    this.l1L2BridgeSharedConfiguration = lineaL1L2BridgeSharedConfiguration;
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "generateConflatedTracesToFileV2";
  }

  /**
   * Handles execution traces generation logic.
   *
   * @param request holds parameters of the RPC request.
   * @return an execution file trace.
   */
  public TraceFile execute(final PluginRpcRequest request) {
    return requestLimiter.execute(request, this::generateTraceFile);
  }

  private TraceFile generateTraceFile(PluginRpcRequest request) {
    Stopwatch sw = Stopwatch.createStarted();

    this.traceService =
        Optional.ofNullable(traceService).orElse(BesuServiceProvider.getTraceService(besuContext));

    final Object[] rawParams = request.getParams();

    Validator.validatePluginRpcRequestParams(rawParams);

    TraceRequestParams params =
        CONVERTER.fromJson(CONVERTER.toJson(rawParams[0]), TraceRequestParams.class);

    params.validate();

    final long fromBlock = params.startBlockNumber();
    final long toBlock = params.endBlockNumber();
    final ZkTracer tracer =
        new ZkTracer(
            l1L2BridgeSharedConfiguration,
            BesuServiceProvider.getBesuService(besuContext, BlockchainService.class)
                .getChainId()
                .orElseThrow());
    final TraceWriter traceWriter = new TraceWriter(tracer);

    traceService.trace(
        fromBlock,
        toBlock,
        worldStateBeforeTracing -> tracer.traceStartConflation(toBlock - fromBlock + 1),
        tracer::traceEndConflation,
        tracer);

    log.info("[TRACING] trace for {}-{} computed in {}", fromBlock, toBlock, sw);
    sw.reset().start();

    final Path path =
        traceWriter.writeTraceToFile(
            tracesOutputPath,
            params.startBlockNumber(),
            params.endBlockNumber(),
            params.expectedTracesEngineVersion());
    log.info("[TRACING] trace for {}-{} serialized to {} in {}", path, toBlock, fromBlock, sw);

    return new TraceFile(params.expectedTracesEngineVersion(), path.toString());
  }
}

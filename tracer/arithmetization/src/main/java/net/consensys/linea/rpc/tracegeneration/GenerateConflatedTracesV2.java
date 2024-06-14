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

package net.consensys.linea.rpc.tracegeneration;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.InvalidParameterException;

import com.google.common.base.Stopwatch;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.json.JsonConverter;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.response.RpcErrorType;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/**
 * Sets up an RPC endpoint for generating conflated file trace. This class provides an RPC endpoint
 * named 'generateConflatedTracesToFileV0' under the 'rollup' namespace. When this endpoint is
 * called, it triggers the execution of the 'execute' method, which generates conflated file traces
 * based on the provided request parameters and writes them to a file.
 */
@Slf4j
@RequiredArgsConstructor
public class GenerateConflatedTracesV2 {
  private static final JsonConverter CONVERTER = JsonConverter.builder().build();

  private final BesuContext besuContext;
  private Path tracesPath;
  private TraceService traceService;

  public String getNamespace() {
    return "rollup";
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
    Stopwatch sw = Stopwatch.createStarted();
    if (this.traceService == null) {
      this.traceService = getTraceService();
    }
    if (this.tracesPath == null) {
      this.tracesPath = getTracesPath();
    }

    final Object[] rawParams = request.getParams();

    // validate params size
    if (rawParams.length != 1) {
      throw new InvalidParameterException(
          "Expected a single params object in the params array but got %d"
              .formatted(rawParams.length));
    }

    try {
      TraceRequestParams params =
          CONVERTER.fromJson(CONVERTER.toJson(rawParams[0]), TraceRequestParams.class);

      params.validateTracerVersion();

      final long fromBlock = params.startBlockNumber();
      final long toBlock = params.endBlockNumber();
      final ZkTracer tracer = new ZkTracer();
      traceService.trace(
          fromBlock,
          toBlock,
          worldStateBeforeTracing -> tracer.traceStartConflation(toBlock - fromBlock + 1),
          tracer::traceEndConflation,
          tracer);
      log.info("[TRACING] trace for {}-{} computed in {}", fromBlock, toBlock, sw);
      sw.reset().start();
      final String path = writeTraceToFile(tracer, params.expectedTracesEngineVersion());
      log.info("[TRACING] trace for {}-{} serialized to {} in {}", path, toBlock, fromBlock, sw);
      return new TraceFile(params.expectedTracesEngineVersion(), path);
    } catch (Exception ex) {
      throw new PluginRpcEndpointException(RpcErrorType.PLUGIN_INTERNAL_ERROR, ex.getMessage());
    }
  }

  private Path getTracesPath() {
    final String envVar = System.getenv("TRACES_DIR");
    if (envVar == null) {
      return this.besuContext
          .getService(BesuConfiguration.class)
          .map(BesuConfiguration::getDataPath)
          .map(x -> x.resolve("traces"))
          .orElseThrow(
              () ->
                  new RuntimeException(
                      "Unable to find data path. Please ensure BesuConfiguration is registered."));
    } else {
      return Paths.get(envVar);
    }
  }

  private TraceService getTraceService() {
    return this.besuContext
        .getService(TraceService.class)
        .orElseThrow(
            () ->
                new RuntimeException(
                    "Unable to find trace service. Please ensure TraceService is registered."));
  }

  private String writeTraceToFile(final ZkTracer tracer, final String traceRuntimeVersion) {
    final Path fileName = generateOutputFileName(traceRuntimeVersion);
    tracer.writeToFile(fileName);
    return fileName.toAbsolutePath().toString();
  }

  private Path generateOutputFileName(final String tracesEngineVersion) {
    if (!Files.isDirectory(tracesPath) && !tracesPath.toFile().mkdirs()) {
      throw new RuntimeException(
          String.format(
              "Trace directory '%s' does not exist and could not be made.",
              tracesPath.toAbsolutePath()));
    }

    return tracesPath.resolve(
        Paths.get(
            String.format(
                "%.10s-%s.traces.%s",
                System.currentTimeMillis(), tracesEngineVersion, getFileFormat())));
  }

  private String getFileFormat() {
    return "lt";
  }
}

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

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.Optional;

import com.google.common.base.Stopwatch;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.plugins.BesuServiceProvider;
import net.consensys.linea.plugins.rpc.RequestLimiter;
import net.consensys.linea.plugins.rpc.Validator;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.json.JsonConverter;
import org.hyperledger.besu.plugin.BesuContext;
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
  private static final String TRACE_FILE_EXTENSION = ".lt";
  private static final String TRACE_TEMP_FILE_EXTENSION = ".lt.tmp";

  private final RequestLimiter requestLimiter;

  private final Path tracesOutputPath;
  private final BesuContext besuContext;
  private TraceService traceService;

  public GenerateConflatedTracesV2(
      final BesuContext besuContext,
      final RequestLimiter requestLimiter,
      final TracesEndpointConfiguration endpointConfiguration) {
    this.besuContext = besuContext;
    this.requestLimiter = requestLimiter;

    this.tracesOutputPath = Paths.get(endpointConfiguration.tracesOutputPath());
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
    final ZkTracer tracer = new ZkTracer();

    traceService.trace(
        fromBlock,
        toBlock,
        worldStateBeforeTracing -> tracer.traceStartConflation(toBlock - fromBlock + 1),
        tracer::traceEndConflation,
        tracer);

    log.info("[TRACING] trace for {}-{} computed in {}", fromBlock, toBlock, sw);
    sw.reset().start();

    final String path = writeTraceToFile(tracer, params);
    log.info("[TRACING] trace for {}-{} serialized to {} in {}", path, toBlock, fromBlock, sw);

    return new TraceFile(params.expectedTracesEngineVersion(), path);
  }

  @SneakyThrows(IOException.class)
  private String writeTraceToFile(
      final ZkTracer tracer, final TraceRequestParams traceRequestParams) {
    // Generate the original and final trace file name.
    final String origTraceFileName = generateOutputFileName(traceRequestParams);
    // Generate and resolve the original and final trace file path.
    final Path origTraceFilePath = generateOutputFilePath(origTraceFileName + TRACE_FILE_EXTENSION);

    // Write the trace at the original and final trace file path, but with the suffix .tmp at the
    // end of the file.
    final Path tmpTraceFilePath =
        tracer.writeToTmpFile(tracesOutputPath, origTraceFileName + ".", TRACE_TEMP_FILE_EXTENSION);
    // After trace writing is complete, rename the file by removing the .tmp prefix, indicating
    // the file is complete and should not be corrupted due to trace writing issues.
    final Path finalizedTraceFilePath =
        Files.move(tmpTraceFilePath, origTraceFilePath, StandardCopyOption.ATOMIC_MOVE);

    return finalizedTraceFilePath.toAbsolutePath().toString();
  }

  private Path generateOutputFilePath(final String traceFileName) {
    if (!Files.isDirectory(tracesOutputPath) && !tracesOutputPath.toFile().mkdirs()) {
      throw new RuntimeException(
          String.format(
              "Trace directory '%s' does not exist and could not be made.",
              tracesOutputPath.toAbsolutePath()));
    }

    return tracesOutputPath.resolve(Paths.get(traceFileName));
  }

  private String generateOutputFileName(final TraceRequestParams traceRequestParams) {
    return String.format(
        "%s-%s.conflated.%s",
        traceRequestParams.startBlockNumber(),
        traceRequestParams.endBlockNumber(),
        traceRequestParams.expectedTracesEngineVersion());
  }
}

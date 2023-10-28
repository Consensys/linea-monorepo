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

package net.consensys.linea.tracegeneration.rpc;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.zip.GZIPOutputStream;

import com.fasterxml.jackson.core.JsonEncoding;
import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonGenerator;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/** Responsible for conflated file traces generation. */
public class RollupGenerateConflatedTracesToFileV0 {

  private final BesuContext besuContext;
  private final JsonFactory jsonFactory = new JsonFactory();
  private final boolean isGzipEnabled = true;

  private Path tracesPath;
  private TraceService traceService;

  public RollupGenerateConflatedTracesToFileV0(final BesuContext besuContext) {
    this.besuContext = besuContext;
  }

  public String getNamespace() {
    return "rollup";
  }

  public String getName() {
    return "generateConflatedTracesToFileV0";
  }

  /**
   * Handles execution traces generation logic.
   *
   * @param request holds parameters of the RPC request.
   * @return an execution file trace.
   */
  public FileTrace execute(final PluginRpcRequest request) {
    if (traceService == null) {
      traceService = initTraceService();
    }

    if (tracesPath == null) {
      tracesPath = initTracesPath();
    }

    try {
      TraceRequestParams params = TraceRequestParams.createTraceParams(request.getParams());

      final long fromBlock = params.fromBlock();
      final long toBlock = params.toBlock();
      final ZkTracer tracer = new ZkTracer();

      traceService.trace(
          fromBlock,
          toBlock,
          worldStateBeforeTracing -> {
            // before tracing
            tracer.traceStartConflation(toBlock - fromBlock + 1);
          },
          worldStateAfterTracing -> {
            // after tracing
            tracer.traceEndConflation();
          },
          tracer);

      final String path = writeTraceToFile(tracer, params.runtimeVersion());

      return new FileTrace(params.runtimeVersion(), path);
    } catch (Exception ex) {
      throw new PluginRpcEndpointException(ex.getMessage());
    }
  }

  private Path initTracesPath() {
    final Path dataPath =
        besuContext
            .getService(BesuConfiguration.class)
            .map(BesuConfiguration::getDataPath)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Unable to find data path. Please ensure BesuConfiguration is registered."));

    return dataPath.resolve("traces");
  }

  private TraceService initTraceService() {
    return besuContext
        .getService(TraceService.class)
        .orElseThrow(
            () ->
                new RuntimeException(
                    "Unable to find trace service. Please ensure TraceService is registered."));
  }

  private String writeTraceToFile(final ZkTracer tracer, final String traceRuntimeVersion) {
    final File file = generateOutputFile(traceRuntimeVersion);
    final OutputStream outputStream = createOutputStream(file);

    try (JsonGenerator jsonGenerator =
        jsonFactory.createGenerator(outputStream, JsonEncoding.UTF8)) {
      jsonGenerator.useDefaultPrettyPrinter();
      jsonGenerator.writeObject(tracer.getJsonTrace());

    } catch (IOException e) {
      throw new RuntimeException(e);
    }

    return file.getAbsolutePath();
  }

  private OutputStream createOutputStream(final File file) {
    try {
      FileOutputStream fileOutputStream = new FileOutputStream(file);
      if (isGzipEnabled) {
        return new GZIPOutputStream(fileOutputStream);
      }

      return fileOutputStream;
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  private File generateOutputFile(final String tracesEngineVersion) {

    if (!Files.isDirectory(tracesPath) && !tracesPath.toFile().mkdirs()) {
      throw new RuntimeException(
          String.format(
              "Trace directory '%s' does not exist and could not be made.",
              tracesPath.toAbsolutePath()));
    }

    return tracesPath
        .resolve(
            String.format(
                "%.10s-%s.traces.%s",
                System.currentTimeMillis(), tracesEngineVersion, getFileFormat()))
        .toFile();
  }

  private String getFileFormat() {
    return isGzipEnabled ? "json.gz" : "json";
  }
}

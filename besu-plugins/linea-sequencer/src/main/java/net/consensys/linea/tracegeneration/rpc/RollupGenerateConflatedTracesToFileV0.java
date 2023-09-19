/*
 * Copyright ConsenSys AG.
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
import java.nio.file.Paths;
import java.util.zip.GZIPOutputStream;

import com.fasterxml.jackson.core.JsonEncoding;
import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonGenerator;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/** Responsible for conflated file traces generation. */
public class RollupGenerateConflatedTracesToFileV0 {

  private final BesuContext context;
  private final JsonFactory jsonFactory = new JsonFactory();
  private final boolean isGzipEnabled = true;

  public RollupGenerateConflatedTracesToFileV0(final BesuContext context) {
    this.context = context;
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
    try {
      final TraceService traceService = context.getService(TraceService.class).orElseThrow();
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

  private String writeTraceToFile(final ZkTracer tracer, final String traceRuntimeVersion) {
    final String dataDir = "traces";
    final File file = generateOutputFile(dataDir, traceRuntimeVersion);
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

  private File generateOutputFile(final String traceDir, final String tracesEngineVersion) {

    Path path = Paths.get(traceDir);
    if (!Files.isDirectory(path) && !path.toFile().mkdirs()) {
      throw new RuntimeException(
          String.format(
              "Trace directory '%s' does not exist and could not be made.", path.toAbsolutePath()));
    }

    return path.resolve(
            String.format(
                "%.10s-%s.traces.%s",
                System.currentTimeMillis(), tracesEngineVersion, getFileFormat()))
        .toFile();
  }

  private String getFileFormat() {
    return isGzipEnabled ? "json.gz" : "json";
  }
}

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
package net.consensys.besu.tracing.rpc;

import static tracing.TraceRequestParams.createTraceParams;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.zip.GZIPOutputStream;

import com.fasterxml.jackson.core.JsonEncoding;
import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonGenerator;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.data.BlockContext;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import tracers.TracerFactory;
import tracers.ZkTracerFactory;
import tracing.FileTrace;
import tracing.TraceRequestParams;

public class RollupGenerateConflatedTracesToFileV0 {

  private final BesuContext context;
  private final JsonFactory jsonFactory = new JsonFactory();
  private final boolean isGZIPEnabled = true;
  private final TracerFactory tracerFactory = new ZkTracerFactory();

  public RollupGenerateConflatedTracesToFileV0(final BesuContext context) {
    this.context = context;
  }

  public String getNamespace() {
    return "rollup";
  }

  public String getName() {
    return "generateConflatedTracesToFileV0";
  }

  public FileTrace execute(final PluginRpcRequest request) {
    try {
      TraceRequestParams params = createTraceParams(request.getParams());
      ArrayList<String> paths = new ArrayList<>();
      getBlocks(params.getFromBlock(), params.getToBlock())
          .forEach(
              blockContext ->
                  paths.add(traceBlockAndReturnPath(blockContext, params.getRuntimeVersion())));
      return new FileTrace(params.getRuntimeVersion(), paths);
    } catch (Exception ex) {
      throw new PluginRpcEndpointException(ex.getMessage());
    }
  }

  private String traceBlockAndReturnPath(BlockContext block, String traceRuntimeVersion) {
    final TraceService traceService = context.getService(TraceService.class).orElseThrow();
    final String dataDir = "traces";
    final File file = generateOutputFile(dataDir, block, traceRuntimeVersion);
    final OutputStream outputStream = createOutputStream(file);
    try (JsonGenerator jsonGenerator =
        jsonFactory.createGenerator(outputStream, JsonEncoding.UTF8)) {
      jsonGenerator.useDefaultPrettyPrinter();
      jsonGenerator.writeStartArray();
      traceService.traceBlock(
          block.getBlockHeader().getNumber(), tracerFactory.create(jsonGenerator));
      jsonGenerator.writeEndArray();
      return file.getAbsolutePath();
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  private OutputStream createOutputStream(final File file) {
    try {
      FileOutputStream fileOutputStream = new FileOutputStream(file);
      if (isGZIPEnabled) {
        return new GZIPOutputStream(fileOutputStream);
      }
      return fileOutputStream;
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  private List<BlockContext> getBlocks(final long fromBlock, final long toBlock) {
    BlockchainService blockchainService = context.getService(BlockchainService.class).orElseThrow();
    ArrayList<BlockContext> blockContexts = new ArrayList<>();
    for (long blockNumber = fromBlock; blockNumber <= toBlock; blockNumber++) {
      Optional<BlockContext> block = blockchainService.getBlockByNumber(blockNumber);
      blockContexts.add(block.orElseThrow(() -> new RuntimeException("Block not found")));
    }
    return blockContexts;
  }

  private File generateOutputFile(
      final String traceDir, final BlockContext block, final String tracesEngineVersion) {

    Path path = Paths.get(traceDir);
    if (!Files.isDirectory(path) && !path.toFile().mkdirs()) {
      throw new RuntimeException(
          String.format(
              "Trace directory '%s' does not exist and could not be made.", path.toAbsolutePath()));
    }

    return path.resolve(
            String.format(
                "%d-%s-%.10s-%s.traces.%s",
                block.getBlockHeader().getNumber(),
                block.getBlockHeader().getBlockHash(),
                System.currentTimeMillis(),
                tracesEngineVersion,
                getFileFormat()))
        .toFile();
  }

  private String getFileFormat() {
    return isGZIPEnabled ? "json.gz" : "json";
  }
}

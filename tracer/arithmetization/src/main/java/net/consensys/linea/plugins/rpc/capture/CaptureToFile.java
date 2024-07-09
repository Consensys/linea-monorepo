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

package net.consensys.linea.plugins.rpc.capture;

import com.google.common.base.Stopwatch;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.blockcapture.BlockCapturer;
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
public class CaptureToFile {
  private final BesuContext besuContext;
  private TraceService traceService;

  public CaptureToFile(final BesuContext besuContext) {
    this.besuContext = besuContext;
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "captureConflation";
  }

  /**
   * Handles execution traces generation logic.
   *
   * @param request holds parameters of the RPC request.
   * @return an execution file trace.
   */
  public Capture execute(final PluginRpcRequest request) {
    if (this.traceService == null) {
      this.traceService = getTraceService();
    }

    CaptureParams params = CaptureParams.createTraceParams(request.getParams());
    final long fromBlock = params.fromBlock();
    final long toBlock = params.toBlock();
    final BlockCapturer tracer = new BlockCapturer();

    Stopwatch sw = Stopwatch.createStarted();
    traceService.trace(
        fromBlock,
        toBlock,
        worldStateBeforeTracing -> {
          tracer.setWorld(worldStateBeforeTracing);
          tracer.traceStartConflation(toBlock - fromBlock + 1);
        },
        tracer::traceEndConflation,
        tracer);
    log.info("[CAPTURE] capture for {}-{} computed in {}", fromBlock, toBlock, sw);
    return new Capture(tracer.toJson());
  }

  private TraceService getTraceService() {
    return this.besuContext
        .getService(TraceService.class)
        .orElseThrow(
            () ->
                new RuntimeException(
                    "Unable to find trace service. Please ensure TraceService is registered."));
  }
}

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

package net.consensys.linea.rpc.counters;

import java.util.Map;

import com.google.common.base.Stopwatch;
import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/** Responsible for trace counters generation. */
@Slf4j
public class RollupGenerateCountersV0 {
  private static final int CACHE_SIZE = 10_000;
  static final Cache<Long, Map<String, Integer>> cache =
      CacheBuilder.newBuilder().maximumSize(CACHE_SIZE).build();

  private final BesuContext besuContext;
  private TraceService traceService;

  public RollupGenerateCountersV0(final BesuContext besuContext) {
    this.besuContext = besuContext;
  }

  public String getNamespace() {
    return "rollup";
  }

  public String getName() {
    return "getTracesCountersByBlockNumberV0";
  }

  /**
   * Handles execution traces generation logic.
   *
   * @param request holds parameters of the RPC request.
   * @return an execution file trace.
   */
  public Counters execute(final PluginRpcRequest request) {
    if (traceService == null) {
      traceService = initTraceService();
    }

    try {
      final Stopwatch sw = Stopwatch.createStarted();
      final CountersRequestParams params =
          CountersRequestParams.createTraceParams(request.getParams());
      final long requestedBlockNumber = params.blockNumber();

      final Counters r =
          new Counters(
              params.runtimeVersion(),
              requestedBlockNumber,
              cache
                  .asMap()
                  .computeIfAbsent(
                      requestedBlockNumber,
                      blockNumber -> {
                        final ZkTracer tracer = new ZkTracer();
                        traceService.trace(
                            blockNumber,
                            blockNumber,
                            worldStateBeforeTracing -> tracer.traceStartConflation(1),
                            worldStateAfterTracing -> tracer.traceEndConflation(),
                            tracer);

                        return tracer.getModulesLineCount();
                      }));
      log.info("counters for {} returned in {}", requestedBlockNumber, sw);
      return r;
    } catch (Exception ex) {
      throw new PluginRpcEndpointException(ex.getMessage());
    }
  }

  private TraceService initTraceService() {
    return besuContext
        .getService(TraceService.class)
        .orElseThrow(
            () ->
                new RuntimeException(
                    "Unable to find trace service. Please ensure TraceService is registered."));
  }
}

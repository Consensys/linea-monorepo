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
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.response.RpcErrorType;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.services.TraceService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

/** This class is used to generate trace counters. */
@Slf4j
public class GenerateCountersV0 {
  private static final int CACHE_SIZE = 10_000;
  static final Cache<Long, Map<String, Integer>> cache =
      CacheBuilder.newBuilder().maximumSize(CACHE_SIZE).build();

  private final BesuContext besuContext;
  private TraceService traceService;

  /**
   * Constructor for RollupGenerateCountersV0.
   *
   * @param besuContext the BesuContext to be used.
   */
  public GenerateCountersV0(final BesuContext besuContext) {
    this.besuContext = besuContext;
  }

  public String getNamespace() {
    return "rollup";
  }

  public String getName() {
    return "getTracesCountersByBlockNumberV0";
  }

  /**
   * Executes an RPC request to generate trace counters.
   *
   * @param request The PluginRpcRequest object encapsulating the parameters of the RPC request.
   * @return A Counters object encapsulating the results of the counters generation (Modules Line
   *     Count). The method uses a caching mechanism to store and retrieve previously computed trace
   *     counters for specific block numbers
   *     <p>If an exception occurs during the execution of the request, it is caught and wrapped in
   *     a PluginRpcEndpointException and rethrown.
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
                            tracer::traceEndConflation,
                            tracer);

                        return tracer.getModulesLineCount();
                      }));
      log.info("counters for {} returned in {}", requestedBlockNumber, sw);
      return r;
    } catch (Exception ex) {
      throw new PluginRpcEndpointException(RpcErrorType.PLUGIN_INTERNAL_ERROR, ex.getMessage());
    }
  }

  /**
   * Initialize the TraceService.
   *
   * @return the initialized TraceService.
   */
  private TraceService initTraceService() {
    return besuContext
        .getService(TraceService.class)
        .orElseThrow(
            () ->
                new RuntimeException(
                    "Unable to find trace service. Please ensure TraceService is registered."));
  }
}

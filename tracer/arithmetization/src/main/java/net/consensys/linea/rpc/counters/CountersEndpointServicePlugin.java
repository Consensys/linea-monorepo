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

import java.util.Optional;

import com.google.auto.service.AutoService;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.RpcEndpointService;

/**
 * Sets up an RPC endpoint for generating trace counters.
 *
 * <p>The CountersEndpointServicePlugin registers an RPC endpoint named
 * 'getTracesCountersByBlockNumberV0' under the 'rollup' namespace. When this endpoint is called,
 * returns trace counters based on the provided request parameters. See {@link
 * RollupGenerateCountersV0}
 */
@AutoService(BesuPlugin.class)
public class CountersEndpointServicePlugin extends AbstractLineaRequiredPlugin {

  /**
   * Register the RPC service.
   *
   * @param context the BesuContext to be used.
   */
  @Override
  public void doRegister(final BesuContext context) {
    RollupGenerateCountersV0 method = new RollupGenerateCountersV0(context);

    Optional<RpcEndpointService> service = context.getService(RpcEndpointService.class);
    createAndRegister(
        method,
        service.orElseThrow(
            () ->
                new RuntimeException("Failed to obtain RpcEndpointService from the BesuContext.")));
  }

  /**
   * Create and register the RPC service.
   *
   * @param method the RollupGenerateCountersV0 method to be used.
   * @param rpcEndpointService the RpcEndpointService to be registered.
   */
  private void createAndRegister(
      final RollupGenerateCountersV0 method, final RpcEndpointService rpcEndpointService) {
    rpcEndpointService.registerRPCEndpoint(
        method.getNamespace(), method.getName(), method::execute);
  }

  /** Start the RPC service. This method loads the OpCodes. */
  @Override
  public void start() {
    OpCodes.load();
  }
}

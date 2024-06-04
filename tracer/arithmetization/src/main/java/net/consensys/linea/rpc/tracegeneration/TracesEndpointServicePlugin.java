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

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaSharedOptionsPlugin;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.RpcEndpointService;

/**
 * Registers RPC endpoints .This class provides an RPC endpoint named
 * 'generateConflatedTracesToFileV0' under the 'rollup' namespace. It uses {@link
 * GenerateConflatedTracesV0} to generate conflated file traces. This class provides an RPC endpoint
 * named 'generateConflatedTracesToFileV0' under the 'rollup' namespace.
 */
@AutoService(BesuPlugin.class)
@Slf4j
public class TracesEndpointServicePlugin extends AbstractLineaSharedOptionsPlugin {
  private BesuContext besuContext;
  private RpcEndpointService rpcEndpointService;
  /**
   * Register the RPC service.
   *
   * @param context the BesuContext to be used.
   */
  @Override
  public void register(final BesuContext context) {
    super.register(context);
    besuContext = context;
    rpcEndpointService =
        context
            .getService(RpcEndpointService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain RpcEndpointService from the BesuContext."));
  }

  @Override
  public void beforeExternalServices() {
    super.beforeExternalServices();
    GenerateConflatedTracesV0 method = new GenerateConflatedTracesV0(besuContext);

    createAndRegister(method, rpcEndpointService);
  }

  /**
   * Create and register the RPC service.
   *
   * @param method the RollupGenerateConflatedTracesToFileV0 method to be used.
   * @param rpcEndpointService the RpcEndpointService to be registered.
   */
  private void createAndRegister(
      final GenerateConflatedTracesV0 method, final RpcEndpointService rpcEndpointService) {
    rpcEndpointService.registerRPCEndpoint(
        method.getNamespace(), method.getName(), method::execute);
  }

  /** Start the RPC service. This method loads the OpCodes. */
  @Override
  public void start() {}
}

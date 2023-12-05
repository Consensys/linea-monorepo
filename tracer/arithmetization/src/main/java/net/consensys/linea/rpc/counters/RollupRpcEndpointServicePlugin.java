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
import net.consensys.linea.LineaRequiredPlugin;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.RpcEndpointService;

/** Plugin with RPC endpoints. */
@AutoService(BesuPlugin.class)
public class RollupRpcEndpointServicePlugin extends LineaRequiredPlugin {
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

  private void createAndRegister(
      final RollupGenerateCountersV0 method, final RpcEndpointService rpcEndpointService) {
    rpcEndpointService.registerRPCEndpoint(
        method.getNamespace(), method.getName(), method::execute);
  }

  @Override
  public void start() {
    OpCodes.load();
  }
}

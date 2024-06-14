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

package net.consensys.linea.rpc;

import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.createLimitModules;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.RpcEndpointService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;

/** Registers RPC endpoints. This class provides RPC endpoints under the 'linea' namespace. */
@AutoService(BesuPlugin.class)
@Slf4j
public class LineaEndpointServicePlugin extends AbstractLineaRequiredPlugin {
  private BesuConfiguration besuConfiguration;
  private RpcEndpointService rpcEndpointService;
  private TransactionSimulationService transactionSimulationService;
  private BlockchainService blockchainService;
  private LineaEstimateGas lineaEstimateGasMethod;

  /**
   * Register the RPC service.
   *
   * @param context the BesuContext to be used.
   */
  @Override
  public void doRegister(final BesuContext context) {
    besuConfiguration =
        context
            .getService(BesuConfiguration.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain BesuConfiguration from the BesuContext."));

    rpcEndpointService =
        context
            .getService(RpcEndpointService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain RpcEndpointService from the BesuContext."));

    transactionSimulationService =
        context
            .getService(TransactionSimulationService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain TransactionSimulatorService from the BesuContext."));

    blockchainService =
        context
            .getService(BlockchainService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain BlockchainService from the BesuContext."));

    lineaEstimateGasMethod =
        new LineaEstimateGas(besuConfiguration, transactionSimulationService, blockchainService);

    rpcEndpointService.registerRPCEndpoint(
        lineaEstimateGasMethod.getNamespace(),
        lineaEstimateGasMethod.getName(),
        lineaEstimateGasMethod::execute);
  }

  @Override
  public void beforeExternalServices() {
    super.beforeExternalServices();
    lineaEstimateGasMethod.init(
        rpcConfiguration,
        transactionPoolValidatorConfiguration,
        profitabilityConfiguration,
        createLimitModules(tracerConfiguration),
        l1L2BridgeConfiguration);
  }
}

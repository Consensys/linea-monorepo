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
package net.consensys.linea.sequencer.txpoolvalidation.validators;

import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.ModuleLineCountResult.MODULE_NOT_DEFINED;
import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.ModuleLineCountResult.TX_MODULE_LINE_COUNT_OVERFLOW;

import java.util.Map;
import java.util.Optional;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.config.LineaL1L2BridgeConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.sequencer.modulelimit.ModuleLimitsValidationResult;
import net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidator;

@Slf4j
public class SimulationValidator implements PluginTransactionPoolValidator {
  private final BlockchainService blockchainService;
  private final TransactionSimulationService transactionSimulationService;
  private final LineaTransactionPoolValidatorConfiguration txPoolValidatorConf;
  private final Map<String, Integer> moduleLineLimitsMap;
  private final LineaL1L2BridgeConfiguration l1L2BridgeConfiguration;

  public SimulationValidator(
      final BlockchainService blockchainService,
      final TransactionSimulationService transactionSimulationService,
      final LineaTransactionPoolValidatorConfiguration txPoolValidatorConf,
      final Map<String, Integer> moduleLineLimitsMap,
      final LineaL1L2BridgeConfiguration l1L2BridgeConfiguration) {
    this.blockchainService = blockchainService;
    this.transactionSimulationService = transactionSimulationService;
    this.txPoolValidatorConf = txPoolValidatorConf;
    this.moduleLineLimitsMap = moduleLineLimitsMap;
    this.l1L2BridgeConfiguration = l1L2BridgeConfiguration;
  }

  @Override
  public Optional<String> validateTransaction(
      final Transaction transaction, final boolean isLocal, final boolean hasPriority) {

    final boolean isLocalAndApiEnabled =
        isLocal && txPoolValidatorConf.txPoolSimulationCheckApiEnabled();
    final boolean isRemoteAndP2pEnabled =
        !isLocal && txPoolValidatorConf.txPoolSimulationCheckP2pEnabled();
    if (isRemoteAndP2pEnabled || isLocalAndApiEnabled) {

      final ModuleLineCountValidator moduleLineCountValidator =
          new ModuleLineCountValidator(moduleLineLimitsMap);
      final var chainHeadHeader = blockchainService.getChainHeadHeader();

      final var zkTracer = createZkTracer(chainHeadHeader);
      final var maybeSimulationResults =
          transactionSimulationService.simulate(
              transaction, chainHeadHeader.getBlockHash(), zkTracer, true);

      ModuleLimitsValidationResult moduleLimit =
          moduleLineCountValidator.validate(zkTracer.getModulesLineCount());

      if (moduleLimit.getResult() != ModuleLineCountValidator.ModuleLineCountResult.VALID) {
        return Optional.of(handleModuleOverLimit(moduleLimit));
      }

      if (maybeSimulationResults.isPresent()) {
        final var simulationResult = maybeSimulationResults.get();
        if (simulationResult.isInvalid()) {
          return Optional.of(
              "Invalid transaction"
                  + simulationResult.getInvalidReason().map(ir -> ": " + ir).orElse(""));
        }
        if (!simulationResult.isSuccessful()) {
          return Optional.of(
              "Reverted transaction"
                  + simulationResult
                      .getRevertReason()
                      .map(rr -> ": " + rr.toHexString())
                      .orElse(""));
        }
      }
    }

    return Optional.empty();
  }

  private ZkTracer createZkTracer(final BlockHeader chainHeadHeader) {
    var zkTracer = new ZkTracer(l1L2BridgeConfiguration);
    zkTracer.traceStartConflation(1L);
    zkTracer.traceStartBlock(chainHeadHeader);
    return zkTracer;
  }

  private String handleModuleOverLimit(ModuleLimitsValidationResult moduleLimitResult) {
    if (moduleLimitResult.getResult() == MODULE_NOT_DEFINED) {
      String moduleNotDefinedMsg =
          String.format(
              "Module %s does not exist in the limits file.", moduleLimitResult.getModuleName());
      log.error(moduleNotDefinedMsg);
      return moduleNotDefinedMsg;
    }
    if (moduleLimitResult.getResult() == TX_MODULE_LINE_COUNT_OVERFLOW) {
      String txOverflowMsg =
          String.format(
              "Transaction line count for module %s=%s is above the limit %s",
              moduleLimitResult.getModuleName(),
              moduleLimitResult.getModuleLineCount(),
              moduleLimitResult.getModuleLineLimit());
      log.warn(txOverflowMsg);
      return txOverflowMsg;
    }
    return "Internal Error: do not know what to do with result: " + moduleLimitResult.getResult();
  }
}

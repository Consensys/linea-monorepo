/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.sequencer.txpoolvalidation;

import java.util.Arrays;
import java.util.Optional;
import java.util.Set;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.jsonrpc.JsonRpcManager;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.sequencer.txpoolvalidation.validators.AllowedAddressValidator;
import net.consensys.linea.sequencer.txpoolvalidation.validators.CalldataValidator;
import net.consensys.linea.sequencer.txpoolvalidation.validators.GasLimitValidator;
import net.consensys.linea.sequencer.txpoolvalidation.validators.ProfitabilityValidator;
import net.consensys.linea.sequencer.txpoolvalidation.validators.SimulationValidator;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidator;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionPoolValidatorFactory;

/** Represents a factory for creating transaction pool validators. */
public class LineaTransactionPoolValidatorFactory implements PluginTransactionPoolValidatorFactory {

  private final BesuConfiguration besuConfiguration;
  private final BlockchainService blockchainService;
  private final TransactionSimulationService transactionSimulationService;
  private final LineaTransactionPoolValidatorConfiguration txPoolValidatorConf;
  private final LineaProfitabilityConfiguration profitabilityConf;
  private final Set<Address> denied;
  private final LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration;
  private final LineaTracerConfiguration tracerConfiguration;
  private final Optional<JsonRpcManager> rejectedTxJsonRpcManager;

  public LineaTransactionPoolValidatorFactory(
      final BesuConfiguration besuConfiguration,
      final BlockchainService blockchainService,
      final TransactionSimulationService transactionSimulationService,
      final LineaTransactionPoolValidatorConfiguration txPoolValidatorConf,
      final LineaProfitabilityConfiguration profitabilityConf,
      final Set<Address> deniedAddresses,
      final LineaTracerConfiguration tracerConfiguration,
      final LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration,
      final Optional<JsonRpcManager> rejectedTxJsonRpcManager) {
    this.besuConfiguration = besuConfiguration;
    this.blockchainService = blockchainService;
    this.transactionSimulationService = transactionSimulationService;
    this.txPoolValidatorConf = txPoolValidatorConf;
    this.profitabilityConf = profitabilityConf;
    this.denied = deniedAddresses;
    this.tracerConfiguration = tracerConfiguration;
    this.l1L2BridgeConfiguration = l1L2BridgeConfiguration;
    this.rejectedTxJsonRpcManager = rejectedTxJsonRpcManager;
  }

  /**
   * Creates a new transaction pool validator, that simply calls in sequence all the actual
   * validators, in a fail-fast mode.
   *
   * @return the new transaction pool validator
   */
  @Override
  public PluginTransactionPoolValidator createTransactionValidator() {
    final var validators =
        new PluginTransactionPoolValidator[] {
          new AllowedAddressValidator(denied),
          new GasLimitValidator(txPoolValidatorConf),
          new CalldataValidator(txPoolValidatorConf),
          new ProfitabilityValidator(besuConfiguration, blockchainService, profitabilityConf),
          new SimulationValidator(
              blockchainService,
              transactionSimulationService,
              txPoolValidatorConf,
              tracerConfiguration,
              l1L2BridgeConfiguration,
              rejectedTxJsonRpcManager)
        };

    return (transaction, isLocal, hasPriority) ->
        Arrays.stream(validators)
            .map(v -> v.validateTransaction(transaction, isLocal, hasPriority))
            .filter(Optional::isPresent)
            .findFirst()
            .map(Optional::get);
  }
}

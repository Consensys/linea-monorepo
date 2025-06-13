/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.sequencer.txvalidation;

import static net.consensys.linea.metrics.LineaMetricCategory.TX_POOL_PROFITABILITY;
import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.createLimitModules;

import com.google.auto.service.AutoService;
import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.config.LineaRejectedTxReportingConfiguration;
import net.consensys.linea.jsonrpc.JsonRpcManager;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.sequencer.txpoolvalidation.metrics.TransactionPoolProfitabilityMetrics;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.hyperledger.besu.plugin.services.TransactionPoolValidatorService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.transactionpool.TransactionPoolService;
import org.hyperledger.besu.plugin.services.TransactionValidatorService;

/**
 * This class extends the default transaction validation rules for adding transactions to the
 * transaction pool. It leverages the PluginTransactionValidatorService to manage and customize the
 * process of transaction validation. This includes, for example, setting a deny list of addresses
 * that are not allowed to add transactions to the pool.
 */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionValidatorPlugin extends AbstractLineaRequiredPlugin {
  private ServiceManager serviceManager;
  private TransactionValidatorService transactionValidatorService;

  @Override
  public void doRegister(final ServiceManager serviceManager) {
    this.serviceManager = serviceManager;

    transactionValidatorService =
        serviceManager
            .getService(TransactionValidatorService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain TransactionValidatorService from the ServiceManager."));
  }

  @Override
  public void doStart() {
  }

  @Override
  public void stop() {
    super.stop();
  }
}


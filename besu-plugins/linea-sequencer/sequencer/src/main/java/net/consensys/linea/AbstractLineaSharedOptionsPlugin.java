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

package net.consensys.linea;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.compress.LibCompress;
import net.consensys.linea.config.LineaL1L2BridgeCliOptions;
import net.consensys.linea.config.LineaL1L2BridgeConfiguration;
import net.consensys.linea.config.LineaProfitabilityCliOptions;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaRpcCliOptions;
import net.consensys.linea.config.LineaRpcConfiguration;
import net.consensys.linea.config.LineaTracerCliOptions;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorCliOptions;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorCliOptions;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;

/**
 * This abstract class is used as superclass for all the plugins that share one or more
 * configuration options.
 *
 * <p>Configuration options that are exclusive of a single plugin, are not required to be added
 * here, but they could stay in the class that implement a plugin, but in case that configuration
 * becomes to be used by multiple plugins, then to avoid code duplications and possible different
 * management of the options, it is better to move the configuration here so all plugins will
 * automatically see it.
 */
@Slf4j
public abstract class AbstractLineaSharedOptionsPlugin implements BesuPlugin {
  private static final String CLI_OPTIONS_PREFIX = "linea";
  private static boolean cliOptionsRegistered = false;
  private static boolean configured = false;
  private static LineaTransactionSelectorCliOptions transactionSelectorCliOptions;
  private static LineaTransactionPoolValidatorCliOptions transactionPoolValidatorCliOptions;
  private static LineaL1L2BridgeCliOptions l1L2BridgeCliOptions;
  private static LineaRpcCliOptions rpcCliOptions;
  private static LineaTracerCliOptions tracerCliOptions;
  private static LineaProfitabilityCliOptions profitabilityCliOptions;
  protected static LineaTransactionSelectorConfiguration transactionSelectorConfiguration;
  protected static LineaTransactionPoolValidatorConfiguration transactionPoolValidatorConfiguration;
  protected static LineaL1L2BridgeConfiguration l1L2BridgeConfiguration;
  protected static LineaRpcConfiguration rpcConfiguration;
  protected static LineaProfitabilityConfiguration profitabilityConfiguration;
  protected static LineaTracerConfiguration tracerConfiguration;

  static {
    // force the initialization of the gnark compress native library to fail fast in case of issues
    LibCompress.CompressedSize(new byte[0], 0);
  }

  @Override
  public synchronized void register(final BesuContext context) {
    if (!cliOptionsRegistered) {
      final PicoCLIOptions cmdlineOptions =
          context
              .getService(PicoCLIOptions.class)
              .orElseThrow(
                  () ->
                      new IllegalStateException(
                          "Failed to obtain PicoCLI options from the BesuContext"));
      transactionSelectorCliOptions = LineaTransactionSelectorCliOptions.create();
      transactionPoolValidatorCliOptions = LineaTransactionPoolValidatorCliOptions.create();
      l1L2BridgeCliOptions = LineaL1L2BridgeCliOptions.create();
      rpcCliOptions = LineaRpcCliOptions.create();
      profitabilityCliOptions = LineaProfitabilityCliOptions.create();
      tracerCliOptions = LineaTracerCliOptions.create();

      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, transactionSelectorCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, transactionPoolValidatorCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, l1L2BridgeCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, rpcCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, profitabilityCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, tracerCliOptions);
      cliOptionsRegistered = true;
    }
  }

  @Override
  public void beforeExternalServices() {
    if (!configured) {
      transactionSelectorConfiguration = transactionSelectorCliOptions.toDomainObject();
      transactionPoolValidatorConfiguration = transactionPoolValidatorCliOptions.toDomainObject();
      l1L2BridgeConfiguration = l1L2BridgeCliOptions.toDomainObject();
      rpcConfiguration = rpcCliOptions.toDomainObject();
      profitabilityConfiguration = profitabilityCliOptions.toDomainObject();
      tracerConfiguration = tracerCliOptions.toDomainObject();
      configured = true;
    }

    log.debug(
        "Configured plugin {} with transaction selector configuration: {}",
        getName(),
        transactionSelectorCliOptions);

    log.debug(
        "Configured plugin {} with transaction pool validator configuration: {}",
        getName(),
        transactionPoolValidatorCliOptions);

    log.debug(
        "Configured plugin {} with L1 L2 bridge configuration: {}",
        getName(),
        l1L2BridgeCliOptions);

    log.debug("Configured plugin {} with RPC configuration: {}", getName(), rpcConfiguration);

    log.debug(
        "Configured plugin {} with profitability calculator configuration: {}",
        getName(),
        profitabilityConfiguration);

    log.debug("Configured plugin {} with tracer configuration: {}", getName(), tracerConfiguration);
  }

  @Override
  public void start() {}

  @Override
  public void stop() {
    cliOptionsRegistered = false;
    configured = false;
  }
}

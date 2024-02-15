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
import net.consensys.linea.config.LineaL1L2BridgeCliOptions;
import net.consensys.linea.config.LineaL1L2BridgeConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorCliOptions;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.config.LineaTransactionValidatorCliOptions;
import net.consensys.linea.config.LineaTransactionValidatorConfiguration;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;

@Slf4j
public abstract class AbstractLineaSharedOptionsPlugin implements BesuPlugin {
  private static String CLI_OPTIONS_PREFIX = "linea";
  private static boolean cliOptionsRegistered = false;
  private static boolean configured = false;
  private static LineaTransactionSelectorCliOptions transactionSelectorCliOptions;
  private static LineaTransactionValidatorCliOptions transactionValidatorCliOptions;
  private static LineaL1L2BridgeCliOptions l1L2BridgeCliOptions;
  protected static LineaTransactionSelectorConfiguration transactionSelectorConfiguration;
  protected static LineaTransactionValidatorConfiguration transactionValidatorConfiguration;
  protected static LineaL1L2BridgeConfiguration l1L2BridgeConfiguration;

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
      transactionValidatorCliOptions = LineaTransactionValidatorCliOptions.create();
      l1L2BridgeCliOptions = LineaL1L2BridgeCliOptions.create();

      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, transactionSelectorCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, transactionValidatorCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, l1L2BridgeCliOptions);
      cliOptionsRegistered = true;
    }
  }

  @Override
  public void beforeExternalServices() {
    if (!configured) {
      transactionSelectorConfiguration = transactionSelectorCliOptions.toDomainObject();
      transactionValidatorConfiguration = transactionValidatorCliOptions.toDomainObject();
      l1L2BridgeConfiguration = l1L2BridgeCliOptions.toDomainObject();
      configured = true;
    }

    log.debug(
        "Configured plugin {} with transaction selector configuration: {}",
        getName(),
        transactionSelectorCliOptions);

    log.debug(
        "Configured plugin {} with transaction validator configuration: {}",
        getName(),
        transactionValidatorCliOptions);

    log.debug(
        "Configured plugin {} with L1 L2 bridge configuration: {}",
        getName(),
        l1L2BridgeCliOptions);
  }

  @Override
  public void start() {}

  @Override
  public void stop() {
    cliOptionsRegistered = false;
    configured = false;
  }
}

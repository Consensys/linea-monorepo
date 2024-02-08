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
  private static LineaTransactionSelectorCliOptions transactionSelectorCliOptions;
  private static LineaTransactionValidatorCliOptions transactionValidatorCliOptions;
  protected static LineaTransactionSelectorConfiguration transactionSelectorConfiguration;
  protected static LineaTransactionValidatorConfiguration transactionValidatorConfiguration;

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

      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, transactionSelectorCliOptions);
      cmdlineOptions.addPicoCLIOptions(CLI_OPTIONS_PREFIX, transactionValidatorCliOptions);
      cliOptionsRegistered = true;
    }
  }

  @Override
  public void beforeExternalServices() {
    log.debug(
        "Configuring plugin {} with transaction selector configuration: {}",
        getName(),
        transactionSelectorCliOptions);
    transactionSelectorConfiguration = transactionSelectorCliOptions.toDomainObject();

    log.debug(
        "Configuring plugin {} with transaction validator configuration: {}",
        getName(),
        transactionValidatorCliOptions);
    transactionValidatorConfiguration = transactionValidatorCliOptions.toDomainObject();
  }

  /**
   * Linea plugins need to implement this method. Called by {@link BesuPlugin} register method
   *
   * @param context
   */
  public abstract void doRegister(final BesuContext context);

  @Override
  public void start() {}

  @Override
  public void stop() {
    cliOptionsRegistered = false;
  }
}

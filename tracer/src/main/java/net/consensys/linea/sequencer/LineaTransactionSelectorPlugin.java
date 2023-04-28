/*
 * Copyright ConsenSys AG.
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
package net.consensys.linea.sequencer;

import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.TransactionSelectionService;

import java.util.Optional;

import com.google.auto.service.AutoService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@AutoService(BesuPlugin.class)
public class LineaTransactionSelectorPlugin implements BesuPlugin {

  private static final Logger LOG = LoggerFactory.getLogger(TransactionSelectionService.class);
  private static final String NAME = "linea";
  private final LineaCLIOptions options;
  private Optional<TransactionSelectionService> service;

  public LineaTransactionSelectorPlugin() {
    options = LineaCLIOptions.create();
  }

  @Override
  public Optional<String> getName() {
    return BesuPlugin.super.getName();
  }

  @Override
  public void register(final BesuContext context) {
    final Optional<PicoCLIOptions> cmdlineOptions = context.getService(PicoCLIOptions.class);

    if (cmdlineOptions.isEmpty()) {
      throw new IllegalStateException(
          "Expecting a PicoCLI options to register CLI options with, but none found.");
    }

    cmdlineOptions.get().addPicoCLIOptions(NAME, options);

    service = context.getService(TransactionSelectionService.class);
    if (service.isEmpty()) {
      LOG.error(
          "Failed to register TransactionSelectionService due to a missing TransactionSelectionService.");
    }
    createAndRegister(service.orElseThrow());
  }

  @Override
  public void start() {
    LOG.debug("Starting Linea plugin with configuration: {}", options.toString());
  }

  @Override
  public void stop() {}

  private void createAndRegister(final TransactionSelectionService transactionSelectionService) {
    transactionSelectionService.registerTransactionSeclectorFactory(
        new LineaTransactionSelectorFactory(options.toDomainObject()));
  }
}

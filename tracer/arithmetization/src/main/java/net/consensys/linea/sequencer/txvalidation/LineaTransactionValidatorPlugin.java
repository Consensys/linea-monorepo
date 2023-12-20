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

package net.consensys.linea.sequencer.txvalidation;

import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.HashSet;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Stream;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.PluginTransactionValidatorService;

/**
 * This class extends the default transaction validation rules for adding transactions to the
 * transaction pool. It leverages the PluginTransactionValidatorService to manage and customize the
 * process of transaction validation. This includes, for example, setting a deny list of addresses
 * that are not allowed to add transactions to the pool.
 */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionValidatorPlugin extends AbstractLineaRequiredPlugin {
  public static final String NAME = "linea";
  private final LineaTransactionValidatorCliOptions options;
  private final Set<Address> denied = new HashSet<>();

  public LineaTransactionValidatorPlugin() {
    options = LineaTransactionValidatorCliOptions.create();
  }

  @Override
  public Optional<String> getName() {
    return Optional.of(NAME);
  }

  @Override
  public void doRegister(final BesuContext context) {
    final Optional<PicoCLIOptions> cmdlineOptions = context.getService(PicoCLIOptions.class);

    if (cmdlineOptions.isEmpty()) {
      throw new IllegalStateException("Failed to obtain PicoCLI options from the BesuContext");
    }

    cmdlineOptions.get().addPicoCLIOptions(NAME, options);

    Optional<PluginTransactionValidatorService> service =
        context.getService(PluginTransactionValidatorService.class);
    createAndRegister(
        service.orElseThrow(
            () ->
                new RuntimeException(
                    "Failed to obtain TransactionValidationService from the BesuContext.")));
  }

  @Override
  public void start() {
    final LineaTransactionValidatorConfiguration config = options.toDomainObject();

    try (Stream<String> lines = Files.lines(Paths.get(config.denyListPath()))) {
      lines.forEach(
          l -> {
            final Address address = Address.fromHexString(l.trim());
            denied.add(address);
          });
    } catch (Exception e) {
      throw new RuntimeException(e);
    }

    log.debug("Starting {} with configuration: {}", NAME, options);
  }

  private void createAndRegister(
      final PluginTransactionValidatorService transactionValidationService) {
    transactionValidationService.registerTransactionValidatorFactory(
        new LineaTransactionValidatorFactory(options, denied));
  }
}

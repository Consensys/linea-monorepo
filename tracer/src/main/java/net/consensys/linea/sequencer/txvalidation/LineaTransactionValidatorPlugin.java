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

package net.consensys.linea.sequencer.txvalidation;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Optional;
import java.util.stream.Stream;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.PluginTransactionValidatorService;

/** Implementation of the base {@link BesuPlugin} interfaces for Linea. */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionValidatorPlugin implements BesuPlugin {
  public static final String NAME = "linea";
  private final LineaTransactionValidatorCliOptions options;
  private ArrayList<Address> denied = new ArrayList<>();
  ;

  public LineaTransactionValidatorPlugin() {
    options = LineaTransactionValidatorCliOptions.create();
  }

  @Override
  public Optional<String> getName() {
    return Optional.of(NAME);
  }

  @Override
  public void register(final BesuContext context) {
    final Optional<PicoCLIOptions> cmdlineOptions = context.getService(PicoCLIOptions.class);

    if (cmdlineOptions.isEmpty()) {
      throw new IllegalStateException(
          "Expecting a PicoCLI options to be available, but none found.");
    }

    cmdlineOptions.get().addPicoCLIOptions(NAME, options);

    Optional<PluginTransactionValidatorService> service =
        context.getService(PluginTransactionValidatorService.class);
    if (service.isEmpty()) {
      log.error(
          "Failed to register TransactionValidatorService because it is not available from the BesuContext.");
    }
    createAndRegister(service.orElseThrow());
  }

  @Override
  public void start() {
    final LineaTransactionValidatorConfiguration config = options.toDomainObject();
    Path filePath = Paths.get(config.denyListPath());

    try (Stream<String> lines = Files.lines(filePath)) {
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

  @Override
  public void stop() {}

  private void createAndRegister(
      final PluginTransactionValidatorService transactionValidationService) {
    transactionValidationService.registerTransactionValidatorFactory(
        new LineaTransactionValidatorFactory(denied));
  }
}

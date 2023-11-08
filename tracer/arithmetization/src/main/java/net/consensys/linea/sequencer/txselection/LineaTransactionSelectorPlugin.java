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

package net.consensys.linea.sequencer.txselection;

import java.io.File;
import java.util.Map;
import java.util.Optional;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.LineaRequiredPlugin;
import net.consensys.linea.sequencer.LineaCliOptions;
import net.consensys.linea.sequencer.LineaConfiguration;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.TransactionSelectionService;

/** Implementation of the base {@link BesuPlugin} interface for Linea Transaction Selection. */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionSelectorPlugin extends LineaRequiredPlugin {
  public static final String NAME = "linea";
  private final LineaCliOptions options;
  private Optional<TransactionSelectionService> service;
  private Map<String, Integer> limitsMap;

  public LineaTransactionSelectorPlugin() {
    options = LineaCliOptions.create();
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

    cmdlineOptions.get().addPicoCLIOptions(getName().get(), options);

    service = context.getService(TransactionSelectionService.class);
    createAndRegister(
        service.orElseThrow(
            () ->
                new RuntimeException(
                    "Failed to obtain TransactionSelectionService from the BesuContext.")));
  }

  @Override
  public void start() {
    log.debug("Starting {} with configuration: {}", NAME, options);
    final LineaConfiguration lineaConfiguration = options.toDomainObject();
    ObjectMapper objectMapper = new ObjectMapper();

    try {
      limitsMap =
          objectMapper.readValue(
              new File(lineaConfiguration.moduleLimitsFilePath()),
              new TypeReference<Map<String, Integer>>() {});
    } catch (final Exception e) {
      final String errorMsg =
          "Problem reading the json file containing the limits for the modules: "
              + lineaConfiguration.moduleLimitsFilePath();
      log.error(errorMsg);
      throw new RuntimeException(errorMsg, e);
    }
  }

  @Override
  public void stop() {}

  private void createAndRegister(final TransactionSelectionService transactionSelectionService) {
    transactionSelectionService.registerTransactionSelectorFactory(
        new LineaTransactionSelectorFactory(options, () -> this.limitsMap));
  }
}

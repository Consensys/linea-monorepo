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
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.auto.service.AutoService;
import com.google.common.io.Resources;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import org.apache.tuweni.toml.Toml;
import org.apache.tuweni.toml.TomlParseResult;
import org.apache.tuweni.toml.TomlTable;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.TransactionSelectionService;

/**
 * This class extends the default transaction selection rules used by Besu. It leverages the
 * TransactionSelectionService to manage and customize the process of transaction selection. This
 * includes setting limits such as 'TraceLineLimit', 'maxBlockGas', and 'maxCallData'.
 */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionSelectorPlugin extends AbstractLineaRequiredPlugin {
  public static final String NAME = "linea";
  private final LineaTransactionSelectorCliOptions options;
  private Optional<TransactionSelectionService> service;
  private final Map<String, Integer> limitsMap = new ConcurrentHashMap<>();

  public LineaTransactionSelectorPlugin() {
    options = LineaTransactionSelectorCliOptions.create();
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
    final LineaTransactionSelectorConfiguration lineaConfiguration = options.toDomainObject();
    ObjectMapper objectMapper = new ObjectMapper();

    try {
      URL url = new File(lineaConfiguration.getModuleLimitsFilePath()).toURI().toURL();
      final String tomlString = Resources.toString(url, StandardCharsets.UTF_8);
      TomlParseResult result = Toml.parse(tomlString);
      final TomlTable table = result.getTable("traces-limits");
      table
          .toMap()
          .keySet()
          .forEach(key -> limitsMap.put(key, Math.toIntExact(table.getLong(key))));
    } catch (final Exception e) {
      final String errorMsg =
          "Problem reading the toml file containing the limits for the modules: "
              + lineaConfiguration.getModuleLimitsFilePath();
      log.error(errorMsg);
      throw new RuntimeException(errorMsg, e);
    }
  }

  private void createAndRegister(final TransactionSelectionService transactionSelectionService) {
    transactionSelectionService.registerTransactionSelectorFactory(
        new LineaTransactionSelectorFactory(options, () -> this.limitsMap));
  }
}

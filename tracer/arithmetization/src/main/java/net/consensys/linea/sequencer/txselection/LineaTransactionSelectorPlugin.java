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
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.auto.service.AutoService;
import com.google.common.io.Resources;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.LineaRequiredPlugin;
import org.apache.tuweni.toml.Toml;
import org.apache.tuweni.toml.TomlParseResult;
import org.apache.tuweni.toml.TomlTable;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.PicoCLIOptions;
import org.hyperledger.besu.plugin.services.TransactionSelectionService;

/** Implementation of the base {@link BesuPlugin} interface for Linea Transaction Selection. */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionSelectorPlugin extends LineaRequiredPlugin {
  public static final String NAME = "linea";
  private final LineaTransactionSelectorCliOptions options;
  private Optional<TransactionSelectionService> service;
  private Map<String, Integer> limitsMap = new HashMap<>();

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
      URL url = new File(lineaConfiguration.moduleLimitsFilePath()).toURI().toURL();
      final String tomlString = Resources.toString(url, StandardCharsets.UTF_8);
      TomlParseResult result = Toml.parse(tomlString);
      final TomlTable table = result.getTable("traces-limits");
      table
          .toMap()
          .keySet()
          .forEach(key -> limitsMap.put(toCamelCase(key), Math.toIntExact(table.getLong(key))));
    } catch (final Exception e) {
      final String errorMsg =
          "Problem reading the toml file containing the limits for the modules: "
              + lineaConfiguration.moduleLimitsFilePath();
      log.error(errorMsg);
      throw new RuntimeException(errorMsg, e);
    }
  }

  private String toCamelCase(final String in) {
    final String[] parts = in.toLowerCase().split("_");
    final StringBuilder sb = new StringBuilder();
    Arrays.stream(parts)
        .forEach(p -> sb.append(p.substring(0, 1).toUpperCase()).append(p.substring(1)));
    return sb.toString();
  }

  @Override
  public void stop() {}

  private void createAndRegister(final TransactionSelectionService transactionSelectionService) {
    transactionSelectionService.registerTransactionSelectorFactory(
        new LineaTransactionSelectorFactory(options, () -> this.limitsMap));
  }
}

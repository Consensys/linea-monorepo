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

package net.consensys.linea.sequencer.txpoolvalidation;

import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.createLimitModules;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TransactionPoolValidatorService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;

/**
 * This class extends the default transaction validation rules for adding transactions to the
 * transaction pool. It leverages the PluginTransactionValidatorService to manage and customize the
 * process of transaction validation. This includes, for example, setting a deny list of addresses
 * that are not allowed to add transactions to the pool.
 */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaTransactionPoolValidatorPlugin extends AbstractLineaRequiredPlugin {
  public static final String NAME = "linea";
  private BesuConfiguration besuConfiguration;
  private BlockchainService blockchainService;
  private TransactionPoolValidatorService transactionPoolValidatorService;
  private TransactionSimulationService transactionSimulationService;

  @Override
  public Optional<String> getName() {
    return Optional.of(NAME);
  }

  @Override
  public void doRegister(final BesuContext context) {
    besuConfiguration =
        context
            .getService(BesuConfiguration.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain BesuConfiguration from the BesuContext."));

    blockchainService =
        context
            .getService(BlockchainService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain BlockchainService from the BesuContext."));

    transactionPoolValidatorService =
        context
            .getService(TransactionPoolValidatorService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain TransactionPoolValidationService from the BesuContext."));

    transactionSimulationService =
        context
            .getService(TransactionSimulationService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain TransactionSimulatorService from the BesuContext."));
  }

  @Override
  public void beforeExternalServices() {
    super.beforeExternalServices();
    try (Stream<String> lines =
        Files.lines(
            Path.of(new File(transactionPoolValidatorConfiguration().denyListPath()).toURI()))) {
      final Set<Address> deniedAddresses =
          lines.map(l -> Address.fromHexString(l.trim())).collect(Collectors.toUnmodifiableSet());

      transactionPoolValidatorService.registerPluginTransactionValidatorFactory(
          new LineaTransactionPoolValidatorFactory(
              besuConfiguration,
              blockchainService,
              transactionSimulationService,
              transactionPoolValidatorConfiguration(),
              profitabilityConfiguration(),
              deniedAddresses,
              createLimitModules(tracerConfiguration()),
              l1L2BridgeSharedConfiguration()));

    } catch (Exception e) {
      throw new RuntimeException(e);
    }
  }
}

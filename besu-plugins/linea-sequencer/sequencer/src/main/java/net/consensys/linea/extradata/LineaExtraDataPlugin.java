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

package net.consensys.linea.extradata;

import java.util.Optional;

import com.google.auto.service.AutoService;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.AbstractLineaRequiredPlugin;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.metrics.LineaMetricCategory;
import org.hyperledger.besu.plugin.BesuContext;
import org.hyperledger.besu.plugin.BesuPlugin;
import org.hyperledger.besu.plugin.services.BesuEvents;
import org.hyperledger.besu.plugin.services.RpcEndpointService;

/** This plugin registers handlers that are activated when new blocks are imported */
@Slf4j
@AutoService(BesuPlugin.class)
public class LineaExtraDataPlugin extends AbstractLineaRequiredPlugin {
  public static final String NAME = "linea";
  private BesuContext besuContext;
  private RpcEndpointService rpcEndpointService;

  @Override
  public Optional<String> getName() {
    return Optional.of(NAME);
  }

  @Override
  public void doRegister(final BesuContext context) {
    besuContext = context;
    rpcEndpointService =
        context
            .getService(RpcEndpointService.class)
            .orElseThrow(
                () ->
                    new RuntimeException(
                        "Failed to obtain RpcEndpointService from the BesuContext."));
  }

  /**
   * Starts this plugin and in case the extra data pricing is enabled, as first thing it tries to
   * extract extra data pricing configuration from the chain head, then it starts listening for new
   * imported block, in order to update the extra data pricing on every incoming block.
   */
  @Override
  public void start() {
    super.start();
    if (profitabilityConfiguration().extraDataPricingEnabled()) {
      final var extraDataHandler =
          new LineaExtraDataHandler(rpcEndpointService, profitabilityConfiguration());
      final var chainHeadHeader = blockchainService.getChainHeadHeader();
      final var initialExtraData = chainHeadHeader.getExtraData();
      try {
        extraDataHandler.handle(initialExtraData);
      } catch (final Exception e) {
        log.warn(
            "Failed setting initial pricing conf from extra data field ({}) of the chain head block {}({})",
            initialExtraData,
            chainHeadHeader.getNumber(),
            chainHeadHeader.getBlockHash(),
            e);
      }

      final var besuEventsService =
          besuContext
              .getService(BesuEvents.class)
              .orElseThrow(
                  () -> new RuntimeException("Failed to obtain BesuEvents from the BesuContext."));

      besuEventsService.addBlockAddedListener(
          addedBlockContext -> {
            final var importedBlockHeader = addedBlockContext.getBlockHeader();
            final var latestExtraData = importedBlockHeader.getExtraData();

            try {
              extraDataHandler.handle(latestExtraData);
            } catch (final Exception e) {
              log.warn(
                  "Failed setting pricing conf from extra data field ({}) of latest imported block {}({})",
                  latestExtraData,
                  importedBlockHeader.getNumber(),
                  importedBlockHeader.getBlockHash(),
                  e);
            }
          });
    }

    initMetrics(profitabilityConfiguration());
  }

  private void initMetrics(final LineaProfitabilityConfiguration lineaProfitabilityConfiguration) {
    final var confLabelledGauge =
        metricsSystem.createLabelledGauge(
            LineaMetricCategory.PROFITABILITY,
            "conf",
            "Profitability configuration values at runtime",
            "field");
    confLabelledGauge.labels(lineaProfitabilityConfiguration::fixedCostWei, "fixed_cost_wei");
    confLabelledGauge.labels(lineaProfitabilityConfiguration::variableCostWei, "variable_cost_wei");
    confLabelledGauge.labels(lineaProfitabilityConfiguration::ethGasPriceWei, "eth_gas_price_wei");
  }
}

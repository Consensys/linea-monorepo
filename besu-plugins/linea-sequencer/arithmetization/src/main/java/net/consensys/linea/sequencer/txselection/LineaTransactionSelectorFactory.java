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

import java.util.Map;

import net.consensys.linea.config.LineaL1L2BridgeConfiguration;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.sequencer.txselection.selectors.LineaTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelector;
import org.hyperledger.besu.plugin.services.txselection.PluginTransactionSelectorFactory;

/** Represents a factory for creating transaction selectors. */
public class LineaTransactionSelectorFactory implements PluginTransactionSelectorFactory {
  private final LineaTransactionSelectorConfiguration txSelectorConfiguration;
  private final LineaL1L2BridgeConfiguration l1L2BridgeConfiguration;
  private final LineaProfitabilityConfiguration profitabilityConfiguration;
  private final Map<String, Integer> limitsMap;

  public LineaTransactionSelectorFactory(
      final LineaTransactionSelectorConfiguration txSelectorConfiguration,
      final LineaL1L2BridgeConfiguration l1L2BridgeConfiguration,
      final LineaProfitabilityConfiguration profitabilityConfiguration,
      final Map<String, Integer> limitsMap) {
    this.txSelectorConfiguration = txSelectorConfiguration;
    this.l1L2BridgeConfiguration = l1L2BridgeConfiguration;
    this.profitabilityConfiguration = profitabilityConfiguration;
    this.limitsMap = limitsMap;
  }

  @Override
  public PluginTransactionSelector create() {
    return new LineaTransactionSelector(
        txSelectorConfiguration, l1L2BridgeConfiguration, profitabilityConfiguration, limitsMap);
  }
}

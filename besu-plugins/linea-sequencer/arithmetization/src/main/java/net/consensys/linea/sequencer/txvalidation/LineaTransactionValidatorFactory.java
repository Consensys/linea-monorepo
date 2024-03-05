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

import java.util.Set;

import net.consensys.linea.config.LineaTransactionValidatorConfiguration;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionValidator;
import org.hyperledger.besu.plugin.services.txvalidator.PluginTransactionValidatorFactory;

/** Represents a factory for creating transaction validators. */
public class LineaTransactionValidatorFactory implements PluginTransactionValidatorFactory {

  private final LineaTransactionValidatorConfiguration txValidatorConf;
  private final Set<Address> denied;

  public LineaTransactionValidatorFactory(
      final LineaTransactionValidatorConfiguration txValidatorConf, final Set<Address> denied) {
    this.txValidatorConf = txValidatorConf;
    this.denied = denied;
  }

  @Override
  public PluginTransactionValidator create() {
    return new LineaTransactionValidator(txValidatorConf, denied);
  }
}

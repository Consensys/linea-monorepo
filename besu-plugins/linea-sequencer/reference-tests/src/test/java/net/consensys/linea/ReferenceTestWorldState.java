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

package net.consensys.linea;

import java.util.Map;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import net.consensys.linea.services.kvstore.BonsaiReferenceTestWorldState;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.account.MutableAccount;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;

/** Represent a worldState for testing. */
@JsonIgnoreProperties(ignoreUnknown = true)
public interface ReferenceTestWorldState
    extends org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState {

  static void insertAccount(
      final WorldUpdater updater,
      final Address address,
      final ReferenceTestWorldState.AccountMock toCopy) {
    final MutableAccount account = updater.getOrCreate(address);
    account.setNonce(toCopy.getNonce());
    account.setBalance(toCopy.getBalance());
    account.setCode(toCopy.getCode());
    for (final Map.Entry<UInt256, UInt256> entry : toCopy.getStorage().entrySet()) {
      account.setStorageValue(entry.getKey(), entry.getValue());
    }
  }

  ReferenceTestWorldState copy();

  @JsonCreator
  static ReferenceTestWorldState create(
      final Map<String, ReferenceTestWorldState.AccountMock> accounts) {
    // delegate to a Bonsai reference test world state:
    return BonsaiReferenceTestWorldState.create(accounts);
  }
}

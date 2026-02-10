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
package net.consensys.linea.zktracer.module.hub;

import java.util.Optional;
import net.consensys.linea.zktracer.types.Bytecode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.worldstate.WorldView;

public record Executor(
    Address origin,
    Address executor,
    boolean isDelegated,
    boolean pointsToNonemptyCode,
    boolean pointsToDelegated,
    boolean pointsToExecutableCode) {

  public Executor fromWorld(WorldView world, Address address) {

    final Optional<Account> toAccount = Optional.ofNullable(world.get(address));
    if (toAccount.isEmpty()) {
      return new Executor(address, address, false, false, false, false);
    }

    final Bytecode bytecode = new Bytecode(toAccount.get().getCode());
    if (bytecode.isExecutable()) {
      return true;
    }
    // beyond this point the byte code is either empty or delegated

    if (bytecode.isEmpty()) {
      return false;
    }
    // at this point we know that the recipient is delegated
    // we need to find out whether the delegate has empty code or is delegated

    final Address delegateAddress = bytecode.getDelegateAddressOrZero();
    final Optional<Account> delegateAccount = Optional.ofNullable(world.get(delegateAddress));

    if (delegateAccount.isEmpty()) {
      return false;
    } else {
      final Bytecode delegateBytecode = new Bytecode(delegateAccount.get().getCode());
      return delegateBytecode.isExecutable();
    }
  }
}

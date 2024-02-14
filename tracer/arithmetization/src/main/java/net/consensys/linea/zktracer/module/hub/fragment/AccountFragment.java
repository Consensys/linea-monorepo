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

package net.consensys.linea.zktracer.module.hub.fragment;

import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Accessors(fluent = true)
public final class AccountFragment implements TraceFragment {
  @Getter private final Address who;
  private final AccountSnapshot oldState;
  private final AccountSnapshot newState;
  private final boolean debit;
  private final long cost;
  private final boolean createAddress;
  @Setter private int deploymentNumberInfnty = 0; // retconned on conflation end
  @Setter private boolean existsInfinity = false; // retconned on conflation end

  public AccountFragment(AccountSnapshot oldState, AccountSnapshot newState) {
    this(oldState, newState, false, 0, false);
  }

  public AccountFragment(
      AccountSnapshot oldState,
      AccountSnapshot newState,
      boolean debit,
      long cost,
      boolean createAddress) {
    Preconditions.checkArgument(oldState.address().equals(newState.address()));

    this.who = oldState.address();
    this.oldState = oldState;
    this.newState = newState;
    this.debit = debit;
    this.cost = cost;
    this.createAddress = createAddress;
  }

  @Override
  public Trace trace(Trace trace) {
    final EWord eWho = EWord.of(who);
    final EWord eCodeHash = EWord.of(oldState.code().getCodeHash());
    final EWord eCodeHashNew = EWord.of(newState.code().getCodeHash());

    return trace
        .peekAtAccount(true)
        .pAccountAddrHi(eWho.hi())
        .pAccountAddrLo(eWho.lo())
        .pAccountIsPrecompile(isPrecompile(who))
        .pAccountNonce(Bytes.ofUnsignedLong(oldState.nonce()))
        .pAccountNonceNew(Bytes.ofUnsignedLong(newState.nonce()))
        .pAccountBalance(oldState.balance())
        .pAccountBalanceNew(newState.balance())
        .pAccountCodeSize(Bytes.ofUnsignedInt(oldState.code().getSize()))
        .pAccountCodeSizeNew(Bytes.ofUnsignedInt(newState.code().getSize()))
        .pAccountCodeHashHi(eCodeHash.hi())
        .pAccountCodeHashLo(eCodeHash.lo())
        .pAccountCodeHashHiNew(eCodeHashNew.hi())
        .pAccountCodeHashLoNew(eCodeHashNew.lo())
        .pAccountHasCode(oldState.code().getCodeHash() != Hash.EMPTY)
        .pAccountHasCodeNew(newState.code().getCodeHash() != Hash.EMPTY)
        .pAccountExists(
            oldState.nonce() > 0
                || oldState.code().getCodeHash() != Hash.EMPTY
                || !oldState.balance().isZero())
        .pAccountExistsNew(
            newState.nonce() > 0
                || newState.code().getCodeHash() != Hash.EMPTY
                || !newState.balance().isZero())
        .pAccountWarm(oldState.warm())
        .pAccountWarmNew(newState.warm())
        .pAccountDepNum(Bytes.ofUnsignedInt(oldState.deploymentNumber()))
        .pAccountDepNumNew(Bytes.ofUnsignedInt(newState.deploymentNumber()))
        .pAccountDepStatus(oldState.deploymentStatus())
        .pAccountDepStatusNew(newState.deploymentStatus())
        //      .pAccountDebit(debit)
        //      .pAccountCost(cost)
        //      .pAccountCreateAddress(createAddress)
        .pAccountDeploymentNumberInfty(Bytes.ofUnsignedInt(deploymentNumberInfnty))
    //        .pAccountExistsInfty(existsInfinity)
    ;
  }

  @Override
  public void postConflationRetcon(final Hub hub, final WorldView world) {
    this.deploymentNumberInfnty = hub.transients().conflation().deploymentInfo().number(this.who);
    this.existsInfinity = world.get(this.who) != null;
  }
}

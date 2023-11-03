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

import static net.consensys.linea.zktracer.types.Address.isPrecompile;

import java.math.BigInteger;

import lombok.Getter;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;

@Accessors(fluent = true)
public final class AccountFragment implements TraceFragment {
  @Getter private final Address who;
  private final AccountSnapshot oldState;
  private final AccountSnapshot newState;
  private final boolean debit;
  private final long cost;
  private final boolean createAddress;
  @Setter private int deploymentNumberInfnty;
  @Setter private boolean existsInfinity;

  public AccountFragment(AccountSnapshot oldState, AccountSnapshot newState) {
    this(oldState, newState, false, 0, false);
  }

  public AccountFragment(
      AccountSnapshot oldState,
      AccountSnapshot newState,
      boolean debit,
      long cost,
      boolean createAddress) {
    assert oldState.address() == newState.address();
    this.who = oldState.address();
    this.oldState = oldState;
    this.newState = newState;
    this.debit = debit;
    this.cost = cost;
    this.createAddress = createAddress;
    this.deploymentNumberInfnty = 0; // will be retconned on conflation end
    this.existsInfinity = false; // will be retconned on conflation end
  }

  @Override
  public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
    final EWord eWho = EWord.of(who);
    final EWord eCodeHash = EWord.of(oldState.code().getCodeHash());
    final EWord eCodeHashNew = EWord.of(newState.code().getCodeHash());

    return trace
        .peekAtAccount(true)
        .pAccountAddrHi(eWho.hiBigInt())
        .pAccountAddrLo(eWho.loBigInt())
        .pAccountIsPrecompile(isPrecompile(who))
        .pAccountNonce(BigInteger.valueOf(oldState.nonce()))
        .pAccountNonceNew(BigInteger.valueOf(newState.nonce()))
        .pAccountBalance(oldState.balance().getAsBigInteger())
        .pAccountBalanceNew(newState.balance().getAsBigInteger())
        .pAccountCodeSize(BigInteger.valueOf(oldState.code().getSize()))
        .pAccountCodeSizeNew(BigInteger.valueOf(newState.code().getSize()))
        .pAccountCodeHashHi(eCodeHash.hiBigInt())
        .pAccountCodeHashLo(eCodeHash.loBigInt())
        .pAccountCodeHashHiNew(eCodeHashNew.hiBigInt())
        .pAccountCodeHashLoNew(eCodeHashNew.loBigInt())
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
        .pAccountDepNum(BigInteger.valueOf(oldState.deploymentNumber()))
        .pAccountDepNumNew(BigInteger.valueOf(newState.deploymentNumber()))
        .pAccountDepStatus(oldState.deploymentStatus())
        .pAccountDepStatusNew(newState.deploymentStatus())
        //      .pAccountDebit(debit)
        //      .pAccountCost(cost)
        //      .pAccountCreateAddress(createAddress)
        .pAccountDeploymentNumberInfty(BigInteger.valueOf(deploymentNumberInfnty))
    //    .pAccountExistsInfty(existsInfinity)
    ;
  }

  @Override
  public void postConflationRetcon(Hub hub /* TODO WorldState state */) {
    this.deploymentNumberInfnty = hub.conflation().deploymentInfo().number(this.who);
    this.existsInfinity =
        false; // TODO should be account != null; see with Besu team if we can get a view on
    // the state in traceEndConflation
  }
}

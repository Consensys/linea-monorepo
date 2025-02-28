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

package net.consensys.linea.zktracer.module.hub.fragment.account;

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.Trace.Hub.MULTIPLIER___DOM_SUB_STAMPS;
import static net.consensys.linea.zktracer.types.AddressUtils.highPart;
import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;
import static net.consensys.linea.zktracer.types.AddressUtils.lowPart;

import java.util.Map;
import java.util.Optional;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.DeferRegistry;
import net.consensys.linea.zktracer.module.hub.defer.EndTransactionDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostConflationDefer;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.section.halt.EphemeralAccount;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.worldstate.WorldView;

@Accessors(fluent = true)
public final class AccountFragment
    implements TraceFragment, EndTransactionDefer, PostConflationDefer {

  @Getter private final AccountSnapshot oldState;
  @Getter private final AccountSnapshot newState;
  @Setter private boolean requiresRomlex;
  private int codeFragmentIndex;
  private final Optional<Bytes> addressToTrim;
  private final DomSubStampsSubFragment domSubStampsSubFragment;
  @Setter private RlpAddrSubFragment rlpAddrSubFragment;
  private boolean markedForSelfDestruct;
  private boolean markedForSelfDestructNew;
  final int hubStamp;
  final TransactionProcessingMetadata transactionProcessingMetadata;

  /**
   * {@link AccountFragment} creation requires access to a {@link DeferRegistry} for post-conflation
   * data gathering, which is provided by this factory.
   */
  @RequiredArgsConstructor
  public static class AccountFragmentFactory {
    private final Hub hub;

    public AccountFragment make(
        AccountSnapshot oldState,
        AccountSnapshot newState,
        DomSubStampsSubFragment domSubStampsSubFragment) {
      return new AccountFragment(
          hub, oldState, newState, Optional.empty(), domSubStampsSubFragment);
    }

    public AccountFragment makeWithTrm(
        AccountSnapshot oldState,
        AccountSnapshot newState,
        Bytes toTrim,
        DomSubStampsSubFragment domSubStampsSubFragment) {
      hub.trm().callTrimming(toTrim);
      return new AccountFragment(
          hub, oldState, newState, Optional.of(toTrim), domSubStampsSubFragment);
    }
  }

  public AccountFragment(
      Hub hub,
      AccountSnapshot oldState,
      AccountSnapshot newState,
      Optional<Bytes> addressToTrim,
      DomSubStampsSubFragment domSubStampsSubFragment) {
    checkArgument(oldState.address().equals(newState.address()));

    transactionProcessingMetadata = hub.txStack().current();
    hubStamp = hub.stamp();

    this.oldState = oldState;
    this.newState = newState;
    this.addressToTrim = addressToTrim;
    this.domSubStampsSubFragment = domSubStampsSubFragment;

    // This allows us to properly fill EXISTS_INFTY, DEPLOYMENT_NUMBER_INFTY and CODE_FRAGMENT_INDEX
    hub.defers().scheduleForPostConflation(this);

    // This allows us to properly fill MARKED_FOR_SELFDESTRUCT and MARKED_FOR_SELFDESTRUCT_NEW,
    // among other things
    hub.defers().scheduleForEndTransaction(this);
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    final EWord eCodeHash =
        EWord.of(oldState.deploymentStatus() ? Hash.EMPTY : oldState.code().getCodeHash());
    final EWord eCodeHashNew =
        EWord.of(newState.deploymentStatus() ? Hash.EMPTY : newState.code().getCodeHash());

    // tracing
    domSubStampsSubFragment.trace(trace);
    if (rlpAddrSubFragment != null) {
      rlpAddrSubFragment.trace(trace);
    }

    final boolean hasCode = !eCodeHash.equals(EWord.of(Hash.EMPTY));
    final boolean hasCodeNew = !eCodeHashNew.equals(EWord.of(Hash.EMPTY));

    return trace
        .peekAtAccount(true)
        .pAccountAddressHi(highPart(oldState.address()))
        .pAccountAddressLo(lowPart(oldState.address()))
        .pAccountNonce(Bytes.ofUnsignedLong(oldState.nonce()))
        .pAccountNonceNew(Bytes.ofUnsignedLong(newState.nonce()))
        .pAccountBalance(oldState.balance())
        .pAccountBalanceNew(newState.balance())
        .pAccountCodeSize(oldState.code().getSize())
        .pAccountCodeSizeNew(newState.code().getSize())
        .pAccountCodeHashHi(eCodeHash.hi())
        .pAccountCodeHashHiNew(eCodeHashNew.hi())
        .pAccountCodeHashLo(eCodeHash.lo())
        .pAccountCodeHashLoNew(eCodeHashNew.lo())
        .pAccountHasCode(hasCode)
        .pAccountHasCodeNew(hasCodeNew)
        .pAccountCodeFragmentIndex(codeFragmentIndex)
        .pAccountRomlexFlag(requiresRomlex)
        .pAccountExists(oldState.nonce() > 0 || hasCode || !oldState.balance().isZero())
        .pAccountExistsNew(newState.nonce() > 0 || hasCodeNew || !newState.balance().isZero())
        .pAccountWarmth(oldState.isWarm())
        .pAccountWarmthNew(newState.isWarm())
        .pAccountMarkedForSelfdestruct(markedForSelfDestruct)
        .pAccountMarkedForSelfdestructNew(markedForSelfDestructNew)
        .pAccountDeploymentNumber(oldState.deploymentNumber())
        .pAccountDeploymentStatus(oldState.deploymentStatus())
        .pAccountDeploymentNumberNew(newState.deploymentNumber())
        .pAccountDeploymentStatusNew(newState.deploymentStatus())
        .pAccountTrmFlag(addressToTrim.isPresent())
        .pAccountTrmRawAddressHi(addressToTrim.map(a -> EWord.of(a).hi()).orElse(Bytes.EMPTY))
        .pAccountIsPrecompile(isPrecompile(oldState.address()));
  }

  @Override
  public void resolveAtEndTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    final Map<EphemeralAccount, Integer> effectiveSelfDestructMap =
        transactionProcessingMetadata.getEffectiveSelfDestructMap();
    final EphemeralAccount ephemeralAccount =
        new EphemeralAccount(oldState.address(), oldState.deploymentNumber());
    if (effectiveSelfDestructMap.containsKey(ephemeralAccount)) {
      final int selfDestructTime = effectiveSelfDestructMap.get(ephemeralAccount);
      markedForSelfDestruct =
          domSubStampsSubFragment.domStamp() > MULTIPLIER___DOM_SUB_STAMPS * selfDestructTime;
      markedForSelfDestructNew = hubStamp >= selfDestructTime;
    } else {
      markedForSelfDestruct = false;
      markedForSelfDestructNew = false;
    }
  }

  @Override
  public void resolvePostConflation(Hub hub, WorldView world) {
    // Note: we DO need the CFI, even if we don't have the romlex flag on, for CFI consistency
    // argument.
    try {
      codeFragmentIndex =
          hub.getCodeFragmentIndexByMetaData(
              newState.address(), newState.deploymentNumber(), newState.deploymentStatus());
    } catch (RuntimeException e) {
      // getCfi should NEVER throw en exception when requiresRomLex ≡ true
      checkState(!requiresRomlex);
      codeFragmentIndex = 0;
    }
  }
}

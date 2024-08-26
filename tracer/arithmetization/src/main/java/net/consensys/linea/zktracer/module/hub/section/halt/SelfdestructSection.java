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
package net.consensys.linea.zktracer.module.hub.section.halt;

import java.util.List;
import java.util.Map;

import com.google.common.base.Preconditions;
import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostRollbackDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.SelfdestructScenarioFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata.AttemptedSelfDestruct;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata.EphemeralAccount;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class SelfdestructSection extends TraceSection
    implements PostRollbackDefer, PostTransactionDefer {

  final int id;
  final int hubStamp;
  final TransactionProcessingMetadata transactionProcessingMetadata;
  final short exceptions;

  SelfdestructScenarioFragment selfdestructScenarioFragment;

  final Address address;
  AccountSnapshot accountBefore;
  AccountSnapshot accountAfter;

  final Bytes recipientRawAddress;
  final Address recipientAddress;
  AccountFragment selfDestroyerFirstAccountFragment;
  AccountFragment recipientFirstAccountFragment;
  AccountSnapshot recipientAccountBefore;
  AccountSnapshot recipientAccountAfter;

  final boolean selfDestructTargetsItself;
  @Getter boolean selfDestructWasReverted = false;

  public SelfdestructSection(Hub hub) {
    // up to 8 = 1 + 7 rows
    super(hub, (short) 8);

    // Init
    this.id = hub.currentFrame().id();
    this.transactionProcessingMetadata = hub.txStack().current();
    this.hubStamp = hub.stamp();
    this.exceptions = hub.pch().exceptions();

    final MessageFrame frame = hub.messageFrame();

    // Account
    this.address = frame.getSenderAddress();
    this.accountBefore = AccountSnapshot.canonical(hub, this.address);

    // Recipient
    this.recipientRawAddress = frame.getStackItem(0);
    this.recipientAddress = Address.extract((Bytes32) this.recipientRawAddress);

    this.selfDestructTargetsItself = this.address.equals(this.recipientAddress);

    this.recipientAccountBefore =
        selfDestructTargetsItself
            ? this.accountAfter.deepCopy()
            : AccountSnapshot.canonical(hub, this.recipientAddress);

    selfdestructScenarioFragment = new SelfdestructScenarioFragment();
    // SCN fragment
    this.addFragment(selfdestructScenarioFragment);
    if (Exceptions.any(exceptions)) {
      selfdestructScenarioFragment.setScenario(
          SelfdestructScenarioFragment.SelfdestructScenario.SELFDESTRUCT_EXCEPTION);
    }

    // CON fragment (1)
    final ContextFragment contextFragment = ContextFragment.readCurrentContextData(hub);
    this.addFragment(contextFragment);

    // STATICX case
    if (Exceptions.staticFault(exceptions)) {
      return;
    }

    // OOGX case
    if (Exceptions.any(exceptions)) {
      Preconditions.checkArgument(exceptions == Exceptions.OUT_OF_GAS_EXCEPTION);

      selfDestroyerFirstAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  this.accountBefore,
                  this.accountBefore,
                  DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

      this.addFragment(selfDestroyerFirstAccountFragment);

      recipientFirstAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  this.recipientAccountBefore,
                  this.recipientAccountBefore,
                  this.recipientRawAddress,
                  DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

      this.addFragment(recipientFirstAccountFragment);

      return;
    }

    // Unexceptional case
    Map<EphemeralAccount, List<AttemptedSelfDestruct>> unexceptionalSelfDestructMap =
        hub.txStack().current().getUnexceptionalSelfDestructMap();

    EphemeralAccount ephemeralAccount =
        new EphemeralAccount(this.address, this.accountBefore.deploymentNumber());

    if (unexceptionalSelfDestructMap.containsKey(ephemeralAccount)) {
      List<AttemptedSelfDestruct> attemptedSelfDestructs =
          unexceptionalSelfDestructMap.get(ephemeralAccount);
      attemptedSelfDestructs.add(new AttemptedSelfDestruct(hubStamp, hub.currentFrame()));
      // TODO: double check that re-adding the list to the map is not necessary, as the list is a
      //  reference
      //  unexceptionalSelfDestructMap.put(addressDeploymentNumberKey, hubStampCallFrameValues);
    } else {
      unexceptionalSelfDestructMap.put(
          new EphemeralAccount(this.address, this.accountBefore.deploymentNumber()),
          List.of(new AttemptedSelfDestruct(hubStamp, hub.currentFrame())));
    }

    hub.defers().scheduleForPostRollback(this, hub.currentFrame());
    hub.defers().scheduleForPostTransaction(this);

    // Modify the current account and the recipient account
    // - The current account has its balance reduced to 0 (i+2)
    //   * selfDestroyerFirstAccountFragment
    // - The recipient account, if it is not the current account, receive that balance (+= balance),
    // otherwise remains 0 (i+3)
    //   * recipientFirstAccountFragment
    // - The recipient address will become warm (i+3)
    //   * recipientFirstAccountFragment

    this.accountAfter = this.accountBefore.debit(this.accountBefore.balance());
    selfDestroyerFirstAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                this.accountBefore,
                this.accountAfter,
                DomSubStampsSubFragment.selfdestructDomSubStamps(hub));
    this.addFragment(selfDestroyerFirstAccountFragment);

    this.recipientAccountAfter =
        this.selfDestructTargetsItself
            ? this.accountAfter.deepCopy()
            : this.recipientAccountBefore.credit(
                this.accountBefore
                    .balance()); // NOTE: in the second case account is equal to recipientAccount
    this.recipientAccountAfter = this.recipientAccountAfter.turnOnWarmth();
    recipientFirstAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                this.recipientAccountBefore,
                this.recipientAccountAfter,
                DomSubStampsSubFragment.selfdestructDomSubStamps(hub));
    this.addFragment(recipientFirstAccountFragment);
  }

  @Override
  public void resolvePostRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {
    // Undo the modifications we applied to selfDestroyerFirstAccountFragment and
    // recipientFirstAccountFragment
    final AccountFragment selfDestroyerUndoingAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                this.accountAfter,
                this.accountBefore,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    hubStamp, callFrame.revertStamp(), 2));
    this.addFragment(selfDestroyerUndoingAccountFragment);

    final AccountFragment recipientUndoingAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                this.recipientAccountAfter,
                this.recipientAccountBefore,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    hubStamp, callFrame.revertStamp(), 3));
    this.addFragment(recipientUndoingAccountFragment);

    ContextFragment squashParentContextReturnData =
        ContextFragment.executionProvidesEmptyReturnData(
            hub, hub.callStack().getParentContextNumberById(this.id));
    this.addFragment(squashParentContextReturnData);

    selfDestructWasReverted = true;

    selfdestructScenarioFragment.setScenario(
        SelfdestructScenarioFragment.SelfdestructScenario.SELFDESTRUCT_WILL_REVERT);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {
    if (selfDestructWasReverted) {
      return;
    }

    Map<EphemeralAccount, Integer> effectiveSelfDestructMap =
        this.transactionProcessingMetadata.getEffectiveSelfDestructMap();
    EphemeralAccount ephemeralAccount =
        new EphemeralAccount(this.address, this.accountAfter.deploymentNumber());

    Preconditions.checkArgument(effectiveSelfDestructMap.containsKey(ephemeralAccount));

    // We modify the account fragment to reflect the self-destruct time

    final int selfDestructTime = effectiveSelfDestructMap.get(ephemeralAccount);

    Preconditions.checkArgument(this.hubStamp >= selfDestructTime);

    AccountSnapshot accountBeforeSelfDestruct =
        this.transactionProcessingMetadata.getDestructedAccountsSnapshot().stream()
            .filter(accountSnapshot -> accountSnapshot.address().equals(this.address))
            .findFirst()
            .orElseThrow(() -> new IllegalStateException("Account not found"));

    if (this.hubStamp == selfDestructTime) {
      selfdestructScenarioFragment.setScenario(
          SelfdestructScenarioFragment.SelfdestructScenario
              .SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED);

      AccountFragment accountWipingFragment =
          hub.factories()
              .accountFragment()
              .make(
                  accountBeforeSelfDestruct,
                  this.accountAfter.wipe(),
                  DomSubStampsSubFragment.selfdestructDomSubStamps(hub));
      this.addFragment(accountWipingFragment);
    } else {
      selfdestructScenarioFragment.setScenario(
          SelfdestructScenarioFragment.SelfdestructScenario
              .SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED);
    }
  }
}

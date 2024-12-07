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

import static com.google.common.base.Preconditions.*;
import static net.consensys.linea.zktracer.module.hub.signals.Exceptions.OUT_OF_GAS_EXCEPTION;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

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

  final Address addressWhichMaySelfDestruct;
  AccountSnapshot selfdestructorAccountBefore;
  AccountSnapshot selfdestructorAccountAfter;

  final Bytes recipientAddressUntrimmed;
  final Address recipientAddress;
  AccountFragment selfdestructorFirstAccountFragment;
  AccountFragment recipientFirstAccountFragment;
  AccountSnapshot recipientAccountBefore;
  AccountSnapshot recipientAccountAfter;

  @Getter boolean selfDestructWasReverted = false;

  ContextFragment finalUnexceptionalContextFragment;

  public SelfdestructSection(Hub hub, MessageFrame frame) {
    // up to 8 = 1 + 7 rows
    super(hub, (short) 8);

    // Init
    id = hub.currentFrame().id();
    transactionProcessingMetadata = hub.txStack().current();
    hubStamp = hub.stamp();
    exceptions = hub.pch().exceptions();

    // Account
    addressWhichMaySelfDestruct = frame.getRecipientAddress();
    selfdestructorAccountBefore =
        AccountSnapshot.canonical(hub, frame.getWorldUpdater(), addressWhichMaySelfDestruct);

    // Recipient
    recipientAddressUntrimmed = frame.getStackItem(0);
    recipientAddress = Address.extract(Bytes32.leftPad(recipientAddressUntrimmed));

    // SCN fragment
    selfdestructScenarioFragment = new SelfdestructScenarioFragment();
    if (Exceptions.any(exceptions)) {
      selfdestructScenarioFragment.setScenario(
          SelfdestructScenarioFragment.SelfdestructScenario.SELFDESTRUCT_EXCEPTION);
    }

    // CON fragment (1)
    final ContextFragment readCurrentContext = ContextFragment.readCurrentContextData(hub);

    this.addStack(hub); // stack fragments
    this.addFragment(selfdestructScenarioFragment); // scenario fragment
    this.addFragment(readCurrentContext);

    // STATICX case
    if (Exceptions.staticFault(exceptions)) {
      return;
    }

    // OOGX case
    if (Exceptions.any(exceptions)) {
      checkArgument(exceptions == OUT_OF_GAS_EXCEPTION);

      recipientAccountBefore =
          selfdestructTargetsItself()
              ? selfdestructorAccountBefore
              : AccountSnapshot.canonical(hub, frame.getWorldUpdater(), recipientAddress);

      selfdestructorFirstAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  selfdestructorAccountBefore,
                  selfdestructorAccountBefore,
                  DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 0));

      recipientFirstAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  recipientAccountBefore,
                  recipientAccountBefore,
                  recipientAddressUntrimmed,
                  DomSubStampsSubFragment.standardDomSubStamps(this.hubStamp(), 1));

      this.addFragment(selfdestructorFirstAccountFragment);
      this.addFragment(recipientFirstAccountFragment);
      return;
    }

    // Unexceptional case
    finalUnexceptionalContextFragment = ContextFragment.executionProvidesEmptyReturnData(hub);

    final Map<EphemeralAccount, List<AttemptedSelfDestruct>> unexceptionalSelfDestructMap =
        hub.txStack().current().getUnexceptionalSelfDestructMap();

    final EphemeralAccount ephemeralAccount =
        new EphemeralAccount(
            addressWhichMaySelfDestruct, selfdestructorAccountBefore.deploymentNumber());

    if (unexceptionalSelfDestructMap.containsKey(ephemeralAccount)) {
      unexceptionalSelfDestructMap
          .get(ephemeralAccount)
          .add(new AttemptedSelfDestruct(hubStamp, hub.currentFrame()));
    } else {
      unexceptionalSelfDestructMap.put(
          new EphemeralAccount(
              addressWhichMaySelfDestruct, selfdestructorAccountBefore.deploymentNumber()),
          new ArrayList<>(List.of(new AttemptedSelfDestruct(hubStamp, hub.currentFrame()))));
    }

    hub.defers().scheduleForPostRollback(this, hub.currentFrame());
    hub.defers().scheduleForPostTransaction(this);

    // Modify the current account and the recipient account
    // - The current account has its balance reduced to 0 (i+2)
    //   * selfdestructorFirstAccountFragment
    // - The recipient account, if it is not the current account, receive that balance (+= balance),
    // otherwise remains 0 (i+3)
    //   * recipientFirstAccountFragment
    // - The recipient address will become warm (i+3)
    //   * recipientFirstAccountFragment

    selfdestructorAccountAfter = selfdestructorAccountBefore.deepCopy().setBalanceToZero();

    if (selfdestructTargetsItself()) {
      recipientAccountBefore = selfdestructorAccountAfter.deepCopy();
      recipientAccountAfter = recipientAccountBefore.deepCopy();
    } else {
      recipientAccountBefore =
          AccountSnapshot.canonical(hub, frame.getWorldUpdater(), recipientAddress);
      recipientAccountAfter =
          recipientAccountBefore
              .deepCopy()
              .incrementBalanceBy(selfdestructorAccountBefore.balance())
              .turnOnWarmth();
    }
    checkArgument(recipientAccountAfter.isWarm());

    selfdestructorFirstAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                selfdestructorAccountBefore,
                selfdestructorAccountAfter,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0));
    recipientFirstAccountFragment =
        hub.factories()
            .accountFragment()
            .makeWithTrm(
                recipientAccountBefore,
                recipientAccountAfter,
                recipientAddressUntrimmed,
                DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1));

    this.addFragment(selfdestructorFirstAccountFragment);
    this.addFragment(recipientFirstAccountFragment);
  }

  @Override
  public void resolveUponRollback(Hub hub, MessageFrame messageFrame, CallFrame callFrame) {
    // Undo the modifications we applied to selfdestructorFirstAccountFragment and
    // recipientFirstAccountFragment
    final AccountFragment selfDestroyerUndoingAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                selfdestructorAccountAfter,
                selfdestructorAccountBefore,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    hubStamp, callFrame.revertStamp(), 2));

    final AccountFragment recipientUndoingAccountFragment =
        hub.factories()
            .accountFragment()
            .make(
                recipientAccountAfter,
                recipientAccountBefore,
                DomSubStampsSubFragment.revertWithCurrentDomSubStamps(
                    hubStamp, callFrame.revertStamp(), 3));
    this.addFragment(selfDestroyerUndoingAccountFragment);
    this.addFragment(recipientUndoingAccountFragment);

    selfDestructWasReverted = true;

    selfdestructScenarioFragment.setScenario(
        SelfdestructScenarioFragment.SelfdestructScenario.SELFDESTRUCT_WILL_REVERT);
  }

  @Override
  public void resolvePostTransaction(
      Hub hub, WorldView state, Transaction tx, boolean isSuccessful) {

    if (selfDestructWasReverted) {
      this.addFragment(finalUnexceptionalContextFragment);
      return;
    }

    // beyond this point the self destruct was not reverted
    final Map<EphemeralAccount, Integer> effectiveSelfDestructMap =
        transactionProcessingMetadata.getEffectiveSelfDestructMap();
    final EphemeralAccount ephemeralAccount =
        new EphemeralAccount(
            addressWhichMaySelfDestruct, selfdestructorAccountAfter.deploymentNumber());

    checkArgument(effectiveSelfDestructMap.containsKey(ephemeralAccount));

    // We modify the account fragment to reflect the self-destruct time
    final int hubStampOfTheSelfDestructThatSealedTheDeal =
        effectiveSelfDestructMap.get(ephemeralAccount);

    checkArgument(hubStamp >= hubStampOfTheSelfDestructThatSealedTheDeal);

    final AccountSnapshot accountBeforeSelfDestruct =
        transactionProcessingMetadata.getDestructedAccountsSnapshot().stream()
            .filter(
                accountSnapshot -> accountSnapshot.address().equals(addressWhichMaySelfDestruct))
            .findFirst()
            .orElseThrow(() -> new IllegalStateException("Account not found"));

    if (hubStamp == hubStampOfTheSelfDestructThatSealedTheDeal) {
      selfdestructScenarioFragment.setScenario(
          SelfdestructScenarioFragment.SelfdestructScenario
              .SELFDESTRUCT_WONT_REVERT_NOT_YET_MARKED);

      // the hub's defers.resolvePostTransaction() gets called after the
      // hub's completeLineaTransaction which in turn calls
      // freshDeploymentNumberFinishingSelfdestruct()
      // which raises the deployment number and sets the deployment status to false
      final AccountFragment accountWipingFragment =
          hub.factories()
              .accountFragment()
              .make(
                  accountBeforeSelfDestruct,
                  selfdestructorAccountAfter.wipe(hub.transients().conflation().deploymentInfo()),
                  DomSubStampsSubFragment.selfdestructDomSubStamps(hub));

      this.addFragment(accountWipingFragment);
    } else {
      selfdestructScenarioFragment.setScenario(
          SelfdestructScenarioFragment.SelfdestructScenario
              .SELFDESTRUCT_WONT_REVERT_ALREADY_MARKED);
    }

    this.addFragment(finalUnexceptionalContextFragment);
  }

  private boolean selfdestructTargetsItself() {
    return addressWhichMaySelfDestruct.equals(recipientAddress);
  }
}

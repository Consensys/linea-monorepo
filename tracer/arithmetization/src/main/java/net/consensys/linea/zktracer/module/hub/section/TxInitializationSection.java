/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use hub file except in compliance with
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

package net.consensys.linea.zktracer.module.hub.section;

import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_EXEC;

import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TransactionFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.transients.DeploymentInfo;
import net.consensys.linea.zktracer.types.Bytecode;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class TxInitializationSection extends TraceSection {
  public TxInitializationSection(Hub hub, WorldView world) {
    super(hub, (short) 5);

    final TransactionProcessingMetadata tx = hub.txStack().current();
    final boolean isDeployment = tx.isDeployment();
    final Address toAddress = tx.getEffectiveTo();
    final DeploymentInfo deploymentInfo = hub.transients().conflation().deploymentInfo();

    final Address senderAddress = tx.getSender();
    final Account senderAccount = world.get(senderAddress);
    final AccountSnapshot senderBeforePayingForTransaction =
        AccountSnapshot.fromAccount(
            senderAccount,
            tx.isSenderPreWarmed(),
            deploymentInfo.number(senderAddress),
            deploymentInfo.isDeploying(senderAddress));
    final DomSubStampsSubFragment senderDomSubStamps =
        DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0);

    final Wei transactionGasPrice = Wei.of(tx.getEffectiveGasPrice());
    final Wei value = (Wei) tx.getBesuTransaction().getValue();
    final Wei valueAndGasCost =
        transactionGasPrice.multiply(tx.getBesuTransaction().getGasLimit()).add(value);
    final AccountSnapshot senderAfterPayingForTransaction =
        senderBeforePayingForTransaction.debit(valueAndGasCost).turnOnWarmth().raiseNonce();

    final boolean isSelfCredit = toAddress.equals(senderAddress);

    final Account recipientAccount = world.get(toAddress);

    AccountSnapshot recipientBeforeValueTransfer;

    if (recipientAccount != null) {
      recipientBeforeValueTransfer =
          isSelfCredit
              ? senderAfterPayingForTransaction
              : AccountSnapshot.fromAccount(
                  recipientAccount,
                  tx.isReceiverPreWarmed(),
                  deploymentInfo.number(toAddress),
                  deploymentInfo.isDeploying(toAddress));
    } else {
      recipientBeforeValueTransfer =
          AccountSnapshot.fromAddress(
              toAddress,
              tx.isReceiverPreWarmed(),
              deploymentInfo.number(toAddress),
              deploymentInfo.isDeploying(toAddress));
    }

    if (isDeployment) {
      deploymentInfo.deploy(toAddress);
    }

    final Bytecode initBytecode =
        new Bytecode(tx.getBesuTransaction().getInit().orElse(Bytes.EMPTY));
    final AccountSnapshot recipientAfterValueTransfer =
        isDeployment
            ? recipientBeforeValueTransfer.initiateDeployment(
                value, initBytecode, deploymentInfo.getDeploymentNumber(toAddress))
            : recipientBeforeValueTransfer.credit(value).turnOnWarmth();
    final DomSubStampsSubFragment recipientDomSubStamps =
        DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 1);

    final TransactionFragment txFragment = TransactionFragment.prepare(tx);

    final AccountFragment.AccountFragmentFactory accountFragmentFactory =
        hub.factories().accountFragment();

    this.addFragment(
        accountFragmentFactory.make(
            senderBeforePayingForTransaction, senderAfterPayingForTransaction, senderDomSubStamps));
    this.addFragment(
        accountFragmentFactory
            .makeWithTrm(
                recipientBeforeValueTransfer,
                recipientAfterValueTransfer,
                toAddress,
                recipientDomSubStamps)
            .requiresRomlex(true));
    this.addFragments(
        ImcFragment.forTxInit(hub), ContextFragment.initializeNewExecutionContext(hub), txFragment);

    hub.state.setProcessingPhase(TX_EXEC);
  }
}

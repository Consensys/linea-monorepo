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
package net.consensys.linea.zktracer.module.hub.section;

import static graphql.com.google.common.base.Preconditions.checkArgument;

import java.util.*;
import java.util.stream.Collectors;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.TransactionProcessingType;
import net.consensys.linea.zktracer.module.hub.fragment.AuthorizationFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.types.Bytecode;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class TxAuthorizationMacroPhase {

  public TxAuthorizationMacroPhase(
      WorldView world, Hub hub, TransactionProcessingMetadata txMetadata) {

    checkArgument(
        txMetadata.requiresAuthorizationPhase(), "Transaction does not require TX_AUTH phase");
    checkArgument(
        txMetadata.getBesuTransaction().codeDelegationListSize() > 0,
        "Transaction has empty delegation list");

    final Address senderAddress = txMetadata.getBesuTransaction().getSender();
    int tupleIndex = 0;
    int senderIsAuthorityAcc = 0;

    // Note: precompiles can't sign delegation tuples
    Set<Address> warmAddresses =
        (txMetadata.requiresEvmExecution()
                && txMetadata.getBesuTransaction().getAccessList().isPresent())
            ? hub.txStack().current().getBesuTransaction().getAccessList().get().stream()
                .map(AccessListEntry::address)
                .collect(Collectors.toSet())
            : new HashSet<>();

    /**
     * contains the latest "updated" account snapshots; we need to track:
     *
     * <ul>
     *   <li>nonces
     *   <li>warmths
     *   <li>latest delegation addresses
     * </ul>
     */
    Map<Address, AccountSnapshot> accountSnapshots = new HashMap<>();

    List<CodeDelegation> delegations =
        txMetadata.getBesuTransaction().getCodeDelegationList().get();

    /**
     * For each delegation tuple insert an {@link AuthorizationFragment}. If the tuple's signature
     * manages to recover an address, insert an {@link
     * net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment}
     */
    for (CodeDelegation delegation : delegations) {

      tupleIndex++;

      final AuthorizationFragment authorizationFragment =
          new AuthorizationFragment(
              delegation,
              tupleIndex,
              senderIsAuthorityAcc,
              delegation.authorizer().isPresent()
                  && delegation.authorizer().get().equals(senderAddress),
              false, // TODO
              0);
      new TxAuthorizationSection( hub, authorizationFragment);

      // no address could be recovered
      if (delegation.authorizer().isEmpty()) {
        continue;
      }

      final Address authorityAddress = delegation.authorizer().get();
      AccountSnapshot oldAuthoritySnapshot;

      if (!accountSnapshots.containsKey(authorityAddress)) {
        final boolean isWarm = warmAddresses.contains(authorityAddress);
        final int deploymentNumber =
            hub.transients().conflation().deploymentInfo().deploymentNumber(authorityAddress);
        final int delegationNumber = hub.delegationNumberOf(authorityAddress);

        oldAuthoritySnapshot =
            world.get(authorityAddress) == null
                ? AccountSnapshot.fromAddress(
                    authorityAddress, isWarm, deploymentNumber, false, delegationNumber)
                : AccountSnapshot.fromAccount(
                    world.get(authorityAddress), isWarm, deploymentNumber, false, delegationNumber);
      } else {
        oldAuthoritySnapshot = accountSnapshots.get(authorityAddress);
      }

      // get the correct nonce
      authorizationFragment.authorityNonce(oldAuthoritySnapshot.nonce());

      AccountSnapshot newAuthoritySnapshot = oldAuthoritySnapshot.deepCopy();

      // for invalid tuples
      if (tupleIsInvalid(oldAuthoritySnapshot, delegation)) {
        new TxAuthorizationSection( hub, new AccountFragment(
          hub,
          oldAuthoritySnapshot,
          oldAuthoritySnapshot,
          Optional.empty(),
          DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0),
          TransactionProcessingType.USER
        ));
        continue;
      }

      // for valid tuples
      Bytecode newCode = authorizationFragment.getBytecode();
      newAuthoritySnapshot
          .turnOnWarmth()
          .incrementNonceByOne()
          .incrementDelegationNumberByOne()
          .code(newCode)
          .conditionallyCheckForDelegation(!newCode.isEmpty());

      if (senderIsAuthorityTuple(delegation, senderAddress)) {
        senderIsAuthorityAcc++;
      }

      AccountFragment authorityAccountFragment =
          hub.factories()
              .accountFragment()
              .make(
                  oldAuthoritySnapshot,
                  newAuthoritySnapshot,
                  DomSubStampsSubFragment.standardDomSubStamps(hub.stamp(), 0),
                  TransactionProcessingType.USER);

      new TxAuthorizationSection( hub, authorityAccountFragment);

      // updates
      hub.transients().conflation().updateDelegationNumber(authorityAddress);
      accountSnapshots.put(authorityAddress, newAuthoritySnapshot);
      warmAddresses.add(authorityAddress);
    }
  }

  boolean senderIsAuthorityTuple(CodeDelegation delegation, Address senderAddress) {
    return delegation.authorizer().isPresent()
        && delegation.authorizer().get().equals(senderAddress);
  }

  boolean tupleIsInvalid(AccountSnapshot incomingAccountSnapshot, CodeDelegation delegation) {
    return delegation.authorizer().isPresent();
  }

  public static class TxAuthorizationSection extends TraceSection {
    public TxAuthorizationSection(Hub hub, TraceFragment fragment) {
      super(hub, (short) 1);
      this.addFragment(fragment);
    }
  }
}

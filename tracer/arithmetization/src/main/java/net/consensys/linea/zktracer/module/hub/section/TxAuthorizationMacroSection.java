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
import static graphql.com.google.common.base.Preconditions.checkState;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_PER_AUTH_BASE_COST;
import static net.consensys.linea.zktracer.Trace.GAS_CONST_PER_EMPTY_ACCOUNT;
import static org.hyperledger.besu.evm.account.Account.MAX_NONCE;

import java.math.BigInteger;
import java.util.*;
import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.TransactionProcessingType;
import net.consensys.linea.zktracer.module.hub.fragment.AuthorizationFragment;
import net.consensys.linea.zktracer.module.hub.fragment.DomSubStampsSubFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.transaction.UserTransactionFragment;
import net.consensys.linea.zktracer.types.Bytecode;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class TxAuthorizationMacroSection {

  /**
   * <b>latestAccountSnapshots</b> tracks the latest account snapshots of accounts touched during
   * the TX_WARM phase (provided the transaction requires EVM execution) and the TX_AUTH phase.
   *
   * <ul>
   *   <li>(initially) prewarmed addresses if the transaction requires EVM execution
   *   <li>(over time) the successful delegation authorities
   * </ul>
   *
   * <p>Since we don't perform Ethereum state / accrued state updates ourselves, we need to track:
   *
   * <ul>
   *   <li>nonces
   *   <li>warmths
   *   <li>latest delegation addresses
   * </ul>
   *
   * <p>After the present phase these data may get used in the TX_INIT / TX_SKIP phases.
   */
  @Getter public final Map<Address, AccountSnapshot> latestAccountSnapshots;

  public TxAuthorizationMacroSection(
      Hub hub,
      WorldView world,
      TransactionProcessingMetadata txMetadata,
      Map<Address, AccountSnapshot> initialAccountSnapshots) {

    checkArgument(
        txMetadata.requiresAuthorizationPhase(), "Transaction does not require TX_AUTH phase");
    checkArgument(
        txMetadata.getBesuTransaction().codeDelegationListSize() > 0,
        "Transaction has empty delegation list");

    this.latestAccountSnapshots = initialAccountSnapshots;

    final Address senderAddress = txMetadata.getBesuTransaction().getSender();
    int tupleIndex = 0;
    int successfulDelegationsAcc = 0;
    int successfulSenderIsAuthorityDelegationsAcc = 0;

    /**
     * For each delegation tuple insert an {@link AuthorizationFragment}. If the tuple's signature
     * manages to recover an address, insert an {@link
     * net.consensys.linea.zktracer.module.hub.fragment.account.AccountFragment}
     */
    for (CodeDelegation delegation :
        txMetadata.getBesuTransaction().getCodeDelegationList().get()) {

      tupleIndex++;

      final AuthorizationFragment authorizationFragment =
          new AuthorizationFragment(
              delegation,
              tupleIndex,
              successfulSenderIsAuthorityDelegationsAcc,
              delegation.authorizer().isPresent()
                  && delegation.authorizer().get().equals(senderAddress),
              false, // authorityHasEmptyCodeOrIsDelegated: updated later (if necessary)
              0, // nonce: updated later (if necessary)
              hub.stamp() + 1, // the hub stamp gets updated when we create the TraceSection
              txMetadata,
              hub.blockdata().getChain().id);

      // no address could be recovered
      if (delegation.authorizer().isEmpty()) {
        new TxAuthorizationSection(hub, false, authorizationFragment);
        continue;
      }

      final Address authorityAddress = delegation.authorizer().get();
      AccountSnapshot currAuthoritySnapshot;

      if (latestAccountSnapshots.containsKey(authorityAddress)) {
        currAuthoritySnapshot = latestAccountSnapshots.get(authorityAddress);
      } else {
        final boolean isWarm = false;
        final int deploymentNumber =
            hub.transients().conflation().deploymentInfo().deploymentNumber(authorityAddress);
        final boolean deploymentStatus =
            hub.transients().conflation().deploymentInfo().getDeploymentStatus(authorityAddress);
        final int delegationNumber = hub.delegationNumberOf(authorityAddress);

        checkState(
            !deploymentStatus, "Addresses in the TX_AUTH phase cannot be undergoing deployment");

        currAuthoritySnapshot =
            world.get(authorityAddress) == null
                ? AccountSnapshot.fromAddress(
                    authorityAddress, isWarm, deploymentNumber, deploymentStatus, delegationNumber)
                : AccountSnapshot.fromAccount(
                    world.get(authorityAddress),
                    isWarm,
                    deploymentNumber,
                    deploymentStatus,
                    delegationNumber);
      }

      // check for delegation if account has code;
      currAuthoritySnapshot.checkForDelegationIfAccountHasCode(hub);

      // update the authorization fragment
      authorizationFragment
          .authorityNonce(currAuthoritySnapshot.nonce())
          .authorityHasEmptyCodeOrIsDelegated(
              currAuthoritySnapshot.accountHasEmptyCodeOrIsDelegated());

      AccountSnapshot nextAuthoritySnapshot = currAuthoritySnapshot.deepCopy();
      final int hubStampPlusOne = hub.stamp() + 1;

      // for invalid tuples
      if (!tupleIsValid(
          delegation, currAuthoritySnapshot, senderAddress, hub.blockdata().getChain().id)) {
        new TxAuthorizationSection(
            hub,
            false,
            authorizationFragment,
            new AccountFragment(
                hub,
                currAuthoritySnapshot,
                currAuthoritySnapshot,
                Optional.of(currAuthoritySnapshot.address()),
                DomSubStampsSubFragment.standardDomSubStamps(hubStampPlusOne, 0),
                TransactionProcessingType.USER));

        // We use ``hub.stamp() + 1'' since the hub stamp only gets updated in the TraceSection
        // constructor
        continue;
      }

      // beyond this point the tuple is valid
      if (currAuthoritySnapshot.exists()) {
        successfulDelegationsAcc++;
      }
      Bytecode newCode = authorizationFragment.getBytecode();
      nextAuthoritySnapshot
          .turnOnWarmth()
          .incrementNonceByOne()
          .incrementDelegationNumberByOne()
          .code(newCode)
          .checkForDelegationIfAccountHasCode(hub);

      if (senderIsAuthorityTuple(delegation, senderAddress)) {
        successfulSenderIsAuthorityDelegationsAcc++;
      }

      AccountFragment authorityAccountFragment =
          hub.factories()
              .accountFragment()
              .makeWithTrm(
                  currAuthoritySnapshot,
                  nextAuthoritySnapshot,
                  currAuthoritySnapshot.address(),
                  DomSubStampsSubFragment.standardDomSubStamps(hubStampPlusOne, 0),
                  TransactionProcessingType.USER);

      new TxAuthorizationSection(
          hub, currAuthoritySnapshot.exists(), authorizationFragment, authorityAccountFragment);

      // updates
      hub.transients().conflation().updateDelegationNumber(authorityAddress);
      latestAccountSnapshots.put(authorityAddress, nextAuthoritySnapshot);
    }

    txMetadata.setNumberOfSuccessfulDelegations(successfulDelegationsAcc);
    txMetadata.setNumberOfSuccessfulSenderDelegations(successfulSenderIsAuthorityDelegationsAcc);

    // we finish by including a PEEK_AT_TRANSACTION row
    // this is expected to raise the hub.stamp() so we make it its own TraceSection.
    new TxAuthorizationSection(hub);
  }

  boolean senderIsAuthorityTuple(CodeDelegation delegation, Address senderAddress) {
    return delegation.authorizer().isPresent()
        && delegation.authorizer().get().equals(senderAddress);
  }

  /**
   * Logic shamelessly stolen from <a
   * href=https://github.com/hyperledger/besu/blob/bba22edc005cabab975efe39d98977b666f2bc83/ethereum/core/src/main/java/org/hyperledger/besu/ethereum/mainnet/CodeDelegationProcessor.java#L86">CodeDelegationProcessor.java</a>
   *
   * <p>Documentation taken from <a href="https://eips.ethereum.org/EIPS/eip-7702">the EIP</a>.
   */
  boolean tupleIsValid(
      CodeDelegation delegation,
      AccountSnapshot latestAccountSnapshot,
      Address senderAddress,
      BigInteger networkChainId) {

    /**
     * NOTE: this seems to be the correct definition of <b>halfCurveOrder</b>, compare with
     * https://github.com/ethereum/execution-specs/blob/a32148175b3ea1db5a34caba939627af5be60c9a/tests/prague/eip7702_set_code_tx/test_set_code_txs.py#L2485
     */
    BigInteger halfCurveOrder =
        new BigInteger("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
            .shiftRight(1);

    // we duplicate the logic of CodeDelegationProcessor.java

    // 1. Verify the chain ID is 0 or the ID of the current chain.
    final boolean delegationTupleChainIdIsZeroOrMatchesNetworkChainId =
        (delegation.chainId().equals(BigInteger.ZERO)
            || delegation.chainId().equals(networkChainId));
    if (!delegationTupleChainIdIsZeroOrMatchesNetworkChainId) {
      return false;
    }

    // 2. Verify the nonce is less than 2**64 - 1.
    if (delegation.nonce() == MAX_NONCE) {
      return false;
    }

    // 3.i: noop
    // 3.ii Verify s is less than or equal to secp256k1n/2
    if (delegation.signature().getS().compareTo(halfCurveOrder) > 0) {
      return false;
    }

    // 3. Let authority = ecrecover(msg, y_parity, r, s)
    final Optional<Address> authority = delegation.authorizer();
    if (authority.isEmpty()) {
      return false;
    }

    // sanity check
    checkState(
        authority.get().equals(latestAccountSnapshot.address()),
        "Account snapshot / delegation authority mismatch:"
            + "snapshot address:  %s,"
            + "authority address: %s",
        latestAccountSnapshot.address(),
        authority.get());

    // 4: noop
    // 5. Verify the code of authority is empty or already delegated
    if (!latestAccountSnapshot.accountHasEmptyCodeOrIsDelegated()) {
      return false;
    }

    // 6. Verify the nonce of authority is equal to nonce
    if (delegation.nonce()
        != latestAccountSnapshot.nonce()
            + (senderIsAuthorityTuple(delegation, senderAddress) ? 1 : 0)) {
      return false;
    }

    // 7: noop
    // 8: noop
    // 9: noop

    return true;
  }

  public static class TxAuthorizationSection extends TraceSection {
    public TxAuthorizationSection(
        Hub hub, boolean authorityTupleIsValidAndAuthorityExists, TraceFragment... fragment) {
      super(hub, (short) 2);

      // valid authority tuples whose underlying authority account exists accrue refunds
      if (authorityTupleIsValidAndAuthorityExists) {
        commonValues.refundDelta(GAS_CONST_PER_EMPTY_ACCOUNT - GAS_CONST_PER_AUTH_BASE_COST);
      }
      this.addFragments(fragment);
    }

    /**
     * Adds the final TX_AUTH-phase fragment (which is a single PEEK_AT_TRANSACTION row)
     *
     * @param hub
     */
    public TxAuthorizationSection(Hub hub) {
      super(hub, (short) 1);
      UserTransactionFragment currentTransaction =
          hub.txStack().current().userTransactionFragment();
      this.addFragments(currentTransaction);
    }
  }
}

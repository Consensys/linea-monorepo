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

package net.consensys.linea.zktracer.delegation;

import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.DEFAULT_COINBASE_ADDRESS;
import static net.consensys.linea.zktracer.Trace.LINEA_CHAIN_ID;
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.Optional;
import java.util.stream.Stream;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedList;
import net.consensys.linea.zktracer.module.hub.section.TupleAnalysis;
import net.consensys.linea.zktracer.module.rlpAuth.RlpAuthOperation;
import net.consensys.linea.zktracer.types.Bytecode;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

// https://github.com/Consensys/linea-monorepo/issues/2455

@ExtendWith(UnitTestWatcher.class)
public class MultiDelegationTest extends TracerTestBase {

  static final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  static final Address senderAddress =
      Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount senderAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(42)
          .address(senderAddress)
          .keyPair(senderKeyPair)
          .build();

  static final KeyPair defaultAuthorityKeyPair = new SECP256K1().generateKeyPair();
  static final Address defaultAuthorityAddress =
      Address.extract(Hash.hash(defaultAuthorityKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount defaultAuthorityAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(2))
          .nonce(67)
          .address(defaultAuthorityAddress)
          .keyPair(defaultAuthorityKeyPair)
          .build();

  static final KeyPair recipientKeyPair = new SECP256K1().generateKeyPair();
  static final Address recipientAddress =
      Address.extract(Hash.hash(recipientKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount recipientAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(4))
          .nonce(69)
          .address(recipientAddress)
          .keyPair(recipientKeyPair)
          .build();

  @ParameterizedTest
  @MethodSource("multiDelegationMonoTransactionTestSource")
  void multiDelegationMonoTransactionTest(
      DelegationCase delegationCase1,
      DelegationCase delegationCase2,
      DelegationCase delegationCase3,
      AuthorityCase authorityCase,
      ToyAccount authorityAccount,
      CodeDelegation delegation1,
      CodeDelegation delegation2,
      CodeDelegation delegation3,
      TestInfo testInfo) {
    final ToyAccount actualSenderAccount =
        authorityCase == AuthorityCase.AUTHORITY_IS_SENDER ? authorityAccount : senderAccount;
    final ToyAccount actualRecipientAccount =
        authorityCase == AuthorityCase.AUTHORITY_IS_RECIPIENT ? authorityAccount : recipientAccount;

    final Transaction tx =
        ToyTransaction.builder()
            .sender(actualSenderAccount)
            .to(actualRecipientAccount)
            .keyPair(actualSenderAccount.getKeyPair())
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(actualSenderAccount.getNonce())
            .gasLimit(96000L)
            .addCodeDelegation(delegation1)
            .addCodeDelegation(delegation2)
            .addCodeDelegation(delegation3)
            .build();

    final List<ToyAccount> accounts =
        switch (authorityCase) {
          case AUTHORITY_IS_RANDOM, AUTHORITY_IS_COINBASE ->
              List.of(senderAccount, recipientAccount, authorityAccount);
          case AUTHORITY_IS_SENDER -> List.of(authorityAccount, recipientAccount);
          case AUTHORITY_IS_RECIPIENT -> List.of(senderAccount, authorityAccount);
        };

    ToyExecutionEnvironmentV2 toyExecutionEnvironmentV2 =
        ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
            .accounts(accounts)
            .coinbase(
                authorityCase == AuthorityCase.AUTHORITY_IS_COINBASE
                    ? authorityAccount.getAddress()
                    : DEFAULT_COINBASE_ADDRESS)
            .transaction(tx)
            .build();
    toyExecutionEnvironmentV2.run();

    ModuleOperationStackedList<RlpAuthOperation> operations =
        toyExecutionEnvironmentV2.getHub().rlpAuth().operations();

    /*
    if (operations.size() != 3) {
      System.err.println("Expected 3 RlpAuthOperations, but got " + operations.size());
    }

    if (delegationCase1.tupleAnalysis != operations.get(0).authorizationFragment().tupleAnalysis()
        || delegationCase2.tupleAnalysis
            != operations.get(1).authorizationFragment().tupleAnalysis()
        || delegationCase3.tupleAnalysis
            != operations.get(2).authorizationFragment().tupleAnalysis()) {
      System.err.println("Tuple analyses do not match expected values.");
    }
    */

    assertEquals(3, operations.size());
    assertEquals(
        delegationCase1.tupleAnalysis, operations.get(0).authorizationFragment().tupleAnalysis());
    assertEquals(
        delegationCase2.tupleAnalysis, operations.get(1).authorizationFragment().tupleAnalysis());
    assertEquals(
        delegationCase3.tupleAnalysis, operations.get(2).authorizationFragment().tupleAnalysis());
  }

  @ParameterizedTest
  @MethodSource("multiDelegationMultiTransactionTestSource")
  void multiDelegationMultiTransactionTest(
      DelegationCase delegationCase1,
      DelegationCase delegationCase2,
      DelegationCase delegationCase3,
      AuthorityCase authorityCase,
      ToyAccount authorityAccount,
      CodeDelegation delegation1,
      CodeDelegation delegation2,
      CodeDelegation delegation3,
      TestInfo testInfo) {
    final ToyAccount actualSenderAccount =
        authorityCase == AuthorityCase.AUTHORITY_IS_SENDER ? authorityAccount : senderAccount;
    final ToyAccount actualRecipientAccount =
        authorityCase == AuthorityCase.AUTHORITY_IS_RECIPIENT ? authorityAccount : recipientAccount;

    long senderNonce = actualSenderAccount.getNonce();

    final Transaction tx1 =
        ToyTransaction.builder()
            .sender(actualSenderAccount)
            .to(actualRecipientAccount)
            .keyPair(actualSenderAccount.getKeyPair())
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(senderNonce)
            .gasLimit(96000L)
            .addCodeDelegation(delegation1)
            .build();

    senderNonce++;
    if (authorityCase == AuthorityCase.AUTHORITY_IS_SENDER && delegationCase1.isValid()) {
      senderNonce++;
    }

    final Transaction tx2 =
        ToyTransaction.builder()
            .sender(actualSenderAccount)
            .to(actualRecipientAccount)
            .keyPair(actualSenderAccount.getKeyPair())
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(senderNonce)
            .gasLimit(96000L)
            .addCodeDelegation(delegation2)
            .build();

    senderNonce++;
    if (authorityCase == AuthorityCase.AUTHORITY_IS_SENDER && delegationCase2.isValid()) {
      senderNonce++;
    }

    final Transaction tx3 =
        ToyTransaction.builder()
            .sender(actualSenderAccount)
            .to(actualRecipientAccount)
            .keyPair(actualSenderAccount.getKeyPair())
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(senderNonce)
            .gasLimit(96000L)
            .addCodeDelegation(delegation3)
            .build();

    final List<ToyAccount> accounts =
        switch (authorityCase) {
          case AUTHORITY_IS_RANDOM, AUTHORITY_IS_COINBASE ->
              List.of(senderAccount, recipientAccount, authorityAccount);
          case AUTHORITY_IS_SENDER -> List.of(authorityAccount, recipientAccount);
          case AUTHORITY_IS_RECIPIENT -> List.of(senderAccount, authorityAccount);
        };

    ToyExecutionEnvironmentV2 toyExecutionEnvironmentV2 =
        ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
            .accounts(accounts)
            .coinbase(
                authorityCase == AuthorityCase.AUTHORITY_IS_COINBASE
                    ? authorityAccount.getAddress()
                    : DEFAULT_COINBASE_ADDRESS)
            .transactions(List.of(tx1, tx2, tx3))
            .build();
    toyExecutionEnvironmentV2.run();

    ModuleOperationStackedList<RlpAuthOperation> operations =
        toyExecutionEnvironmentV2.getHub().rlpAuth().operations();

    /*
    if (operations.size() != 3) {
      System.err.println("Expected 3 RlpAuthOperations, but got " + operations.size());
    }

    if (delegationCase1.tupleAnalysis != operations.get(0).authorizationFragment().tupleAnalysis()
        || delegationCase2.tupleAnalysis
            != operations.get(1).authorizationFragment().tupleAnalysis()
        || delegationCase3.tupleAnalysis
            != operations.get(2).authorizationFragment().tupleAnalysis()) {
      System.err.println("Tuple analyses do not match expected values.");
    }
    */

    assertEquals(3, operations.size());
    assertEquals(
        delegationCase1.tupleAnalysis, operations.get(0).authorizationFragment().tupleAnalysis());
    assertEquals(
        delegationCase2.tupleAnalysis, operations.get(1).authorizationFragment().tupleAnalysis());
    assertEquals(
        delegationCase3.tupleAnalysis, operations.get(2).authorizationFragment().tupleAnalysis());
  }

  @RequiredArgsConstructor
  public enum DelegationCase {
    DELEGATION_TO_NEW_ADDRESS(TupleAnalysis.TUPLE_IS_VALID),
    DELEGATION_TO_CURRENT_DELEGATION(TupleAnalysis.TUPLE_IS_VALID),
    DELEGATION_RESET(TupleAnalysis.TUPLE_IS_VALID), // Address.ZERO
    DELEGATION_FAILURE_DUE_TO_NONCE_MISMATCH(
        TupleAnalysis
            .TUPLE_FAILS_DUE_TO_NONCE_MISMATCH), // authority is recovered and printed in the hub
    DELEGATION_FAILURE_DUE_TO_CHAIN_ID_MISMATCH(
        TupleAnalysis.TUPLE_FAILS_CHAIN_ID_CHECK); // authority is not recovered

    final TupleAnalysis tupleAnalysis;

    boolean isValid() {
      return this == DELEGATION_TO_NEW_ADDRESS
          || this == DELEGATION_TO_CURRENT_DELEGATION
          || this == DELEGATION_RESET;
    }

    short nonceIncrementDueToValidDelegation() {
      return this.isValid() ? (short) 1 : 0;
    }
  }

  public enum AuthorityCase {
    AUTHORITY_IS_RANDOM,
    AUTHORITY_IS_SENDER,
    AUTHORITY_IS_RECIPIENT,
    AUTHORITY_IS_COINBASE; // DEFAULT_COINBASE_ADDRESS
  }

  static Stream<Arguments> multiDelegationMonoTransactionTestSource() {
    return multiDelegationTestSourceBody(false);
  }

  static Stream<Arguments> multiDelegationMultiTransactionTestSource() {
    return multiDelegationTestSourceBody(true);
  }

  static Stream<Arguments> multiDelegationTestSourceBody(boolean isMultiTransaction) {
    ToyAccount authorityAccount;
    List<Arguments> arguments = new ArrayList<>();

    for (DelegationCase delegationCase1 : DelegationCase.values()) {
      for (DelegationCase delegationCase2 : DelegationCase.values()) {
        for (DelegationCase delegationCase3 : DelegationCase.values()) {
          for (AuthorityCase authorityCase : AuthorityCase.values()) {
            authorityAccount =
                switch (authorityCase) {
                  case AUTHORITY_IS_RANDOM, AUTHORITY_IS_COINBASE ->
                      defaultAuthorityAccount.deepCopy();
                  case AUTHORITY_IS_SENDER -> senderAccount.deepCopy();
                  case AUTHORITY_IS_RECIPIENT -> recipientAccount.deepCopy();
                };
            CodeDelegation delegation1;
            CodeDelegation delegation2;
            CodeDelegation delegation3;

            // mono transaction case does not touch this value anymore
            short nonceIncrementDueToAuthorityIsSender =
                authorityCase == AuthorityCase.AUTHORITY_IS_SENDER ? (short) 1 : 0;

            try {
              delegation1 =
                  craftCodeDelegation(
                      authorityAccount, delegationCase1, nonceIncrementDueToAuthorityIsSender);

              if (authorityCase == AuthorityCase.AUTHORITY_IS_SENDER && isMultiTransaction) {
                nonceIncrementDueToAuthorityIsSender++;
              }

              delegation2 =
                  craftCodeDelegation(
                      authorityAccount,
                      delegationCase2,
                      delegationCase1.nonceIncrementDueToValidDelegation()
                          + nonceIncrementDueToAuthorityIsSender);

              if (authorityCase == AuthorityCase.AUTHORITY_IS_SENDER && isMultiTransaction) {
                nonceIncrementDueToAuthorityIsSender++;
              }

              delegation3 =
                  craftCodeDelegation(
                      authorityAccount,
                      delegationCase3,
                      delegationCase1.nonceIncrementDueToValidDelegation()
                          + delegationCase2.nonceIncrementDueToValidDelegation()
                          + nonceIncrementDueToAuthorityIsSender);
            } catch (NoSuchElementException e) {
              // This happens when we try to create a tuple with DELEGATION_TO_CURRENT_DELEGATION
              // but there is no previous delegation
              continue;
            }
            arguments.add(
                Arguments.of(
                    delegationCase1,
                    delegationCase2,
                    delegationCase3,
                    authorityCase,
                    authorityAccount,
                    delegation1,
                    delegation2,
                    delegation3));
          }
        }
      }
    }
    return arguments.stream();
  }

  static final Address delegationAddressA =
      Address.fromHexString("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa");
  static final Address delegationAddressB =
      Address.fromHexString("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb");
  static final Address delegationAddressC =
      Address.fromHexString("0xcccccccccccccccccccccccccccccccccccccccc");

  static CodeDelegation craftCodeDelegation(
      ToyAccount authorityAccount, DelegationCase delegationCase, int nonceOffset) {
    final Optional<Address> previousDelegationAddress = getDelegationAddress(authorityAccount);
    final Address newDelegationAddress =
        switch (delegationCase) {
          case DELEGATION_TO_NEW_ADDRESS -> {
            if (previousDelegationAddress.isEmpty()
                || previousDelegationAddress.get().equals(delegationAddressB)) {
              yield delegationAddressA;
            } else {
              yield delegationAddressB;
            }
          }
          case DELEGATION_TO_CURRENT_DELEGATION -> previousDelegationAddress.orElseThrow();
          case DELEGATION_RESET -> Address.ZERO;
          case DELEGATION_FAILURE_DUE_TO_NONCE_MISMATCH,
              DELEGATION_FAILURE_DUE_TO_CHAIN_ID_MISMATCH ->
              delegationAddressC;
        };
    return org.hyperledger.besu.ethereum.core.CodeDelegation.builder()
        .chainId(
            BigInteger.valueOf(
                LINEA_CHAIN_ID
                    + (delegationCase == DelegationCase.DELEGATION_FAILURE_DUE_TO_CHAIN_ID_MISMATCH
                        ? 101 // arbitrary number to cause chain id mismatch
                        : 0)))
        .address(newDelegationAddress)
        .nonce(
            authorityAccount.getNonce()
                + nonceOffset
                + (delegationCase == DelegationCase.DELEGATION_FAILURE_DUE_TO_NONCE_MISMATCH
                    ? 666 // arbitrary number to cause nonce mismatch
                    : 0))
        .signAndBuild(authorityAccount.getKeyPair());
  }

  public static Optional<Address> getDelegationAddress(ToyAccount account) {
    final Bytecode accountBytecode = new Bytecode(account.getCode());
    return accountBytecode.getDelegateAddress();
  }
}

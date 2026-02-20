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
import org.junit.jupiter.api.Test;
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
          .nonce(69)
          .address(defaultAuthorityAddress)
          .keyPair(defaultAuthorityKeyPair)
          .build();

  static final KeyPair recipientKeyPair = new SECP256K1().generateKeyPair();
  static final Address recipientAddress =
      Address.extract(Hash.hash(recipientKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount recipientAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(4))
          .nonce(666)
          .address(recipientAddress)
          .keyPair(recipientKeyPair)
          .build();

  @ParameterizedTest
  @MethodSource("multiDelegationTestSource")
  void multiDelegationMonoTransactionTest(
      ToyAccount authorityAccount,
      AuthorityCase authorityCase,
      CodeDelegation d1,
      CodeDelegation d2,
      CodeDelegation d3,
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
            .addCodeDelegation(d1)
            .addCodeDelegation(d2)
            .addCodeDelegation(d3)
            .build();

    final List<ToyAccount> accounts =
        switch (authorityCase) {
          case AUTHORITY_IS_RANDOM, AUTHORITY_IS_COINBASE ->
              List.of(senderAccount, recipientAccount, authorityAccount);
          case AUTHORITY_IS_SENDER ->
              List.of(authorityAccount, recipientAccount);
          case AUTHORITY_IS_RECIPIENT ->
              List.of(senderAccount, authorityAccount);
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
    assertEquals(3, operations.size());
    for (RlpAuthOperation operation : operations.getAll()) {
      assertEquals(TupleAnalysis.TUPLE_IS_VALID, operation.authorizationFragment().tupleAnalysis());
    }
  }

  @Test
  void multiDelegationMultiTransactionTest() {
    // TODO
  }

  public enum DelegationCase {
    DELEGATION_TO_NEW_ADDRESS,
    DELEGATION_TO_CURRENT_DELEGATION,
    DELEGATION_RESET, // Address.ZERO
    DELEGATION_FAILURE_DUE_TO_NONCE_MISMATCH, // authority is recovered and printed in the hub
    DELEGATION_FAILURE_DUE_TO_CHAIN_ID_MISMATCH; // authority is not recovered

    boolean isValid() {
      return this == DELEGATION_TO_NEW_ADDRESS
          || this == DELEGATION_TO_CURRENT_DELEGATION
          || this == DELEGATION_RESET;
    }
  }

  public enum AuthorityCase {
    AUTHORITY_IS_RANDOM,
    AUTHORITY_IS_SENDER,
    AUTHORITY_IS_RECIPIENT,
    AUTHORITY_IS_COINBASE; // DEFAULT_COINBASE_ADDRESS
  }

  static Stream<Arguments> multiDelegationTestSource() {
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
            CodeDelegation d1;
            CodeDelegation d2;
            CodeDelegation d3;
            try {
              d1 = craftCodeDelegation(authorityAccount, delegationCase1, 0);
              d2 =
                  craftCodeDelegation(
                      authorityAccount, delegationCase2, delegationCase1.isValid() ? 1 : 0);
              d3 =
                  craftCodeDelegation(
                      authorityAccount,
                      delegationCase3,
                      delegationCase1.isValid() && delegationCase2.isValid()
                          ? 2
                          : (delegationCase2.isValid() ? 1 : 0));
            } catch (NoSuchElementException e) {
              // This happens when we try to create a tuple with DELEGATION_TO_CURRENT_DELEGATION
              // but there is no previous delegation
              continue;
            }
            arguments.add(Arguments.of(authorityAccount, authorityCase, d1, d2, d3));
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
    final Bytecode previousAuthoritAccountBytecode = new Bytecode(authorityAccount.getCode());
    final Optional<Address> previousDelegationAddress =
        previousAuthoritAccountBytecode.getDelegateAddress();

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
                        ? 1
                        : 0)))
        .address(newDelegationAddress)
        .nonce(
            authorityAccount.getNonce()
                + nonceOffset
                + (delegationCase == DelegationCase.DELEGATION_FAILURE_DUE_TO_NONCE_MISMATCH
                    ? 67
                    : 0))
        .signAndBuild(authorityAccount.getKeyPair());
  }
}

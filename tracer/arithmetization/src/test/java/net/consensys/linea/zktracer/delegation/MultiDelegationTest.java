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
import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.ArrayList;
import java.util.List;
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

  static final KeyPair authorityKeyPair = new SECP256K1().generateKeyPair();
  static final Address authorityAddress =
      Address.extract(Hash.hash(authorityKeyPair.getPublicKey().getEncodedBytes()));
  static final ToyAccount authorityAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(2))
          .nonce(69)
          .address(authorityAddress)
          .keyPair(authorityKeyPair)
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

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(senderAccount.getKeyPair())
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(senderAccount.getNonce())
            .gasLimit(96000L)
            .addCodeDelegation(d1)
            .addCodeDelegation(d2)
            .addCodeDelegation(d3)
            .build();

    final List<ToyAccount> accounts =
        switch (authorityCase) {
          case AUTHORITY_IS_RANDOM, AUTHORITY_IS_COINBASE ->
              List.of(senderAccount, recipientAccount, authorityAccount);
          case AUTHORITY_IS_SENDER, AUTHORITY_IS_RECIPIENT ->
              List.of(senderAccount, recipientAccount);
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
  }

  public enum AuthorityCase {
    AUTHORITY_IS_RANDOM,
    AUTHORITY_IS_SENDER,
    AUTHORITY_IS_RECIPIENT,
    AUTHORITY_IS_COINBASE; // DEFAULT_COINBASE_ADDRESS
  }

  static ToyAccount actualAuthorityAccount;

  static Stream<Arguments> multiDelegationTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (DelegationCase delegationCase1 : DelegationCase.values()) {
      for (DelegationCase delegationCase2 : DelegationCase.values()) {
        for (DelegationCase delegationCase3 : DelegationCase.values()) {
          for (AuthorityCase authorityCase : AuthorityCase.values()) {
            // TODO: create deep copy method for ToyAccount / be careful with pointers
            actualAuthorityAccount =
                switch (authorityCase) {
                  case AUTHORITY_IS_RANDOM, AUTHORITY_IS_COINBASE -> authorityAccount;
                  case AUTHORITY_IS_SENDER -> senderAccount;
                  case AUTHORITY_IS_RECIPIENT -> recipientAccount;
                };
            arguments.add(
                Arguments.of(
                    actualAuthorityAccount,
                    authorityCase,
                    craftCodeDelegation(delegationCase1),
                    craftCodeDelegation(delegationCase2),
                    craftCodeDelegation(delegationCase3)));
          }
        }
      }
    }
    return arguments.stream();
  }

  // TODO: maybe update the builder to keep track of the current delegation (via the code)
  static CodeDelegation craftCodeDelegation(DelegationCase delegationCase) {
    final Bytecode previousAuthoritAccountBytecode = new Bytecode(authorityAccount.getCode());
    final Optional<Address> previousDelegation =
        previousAuthoritAccountBytecode.getDelegateAddress();
    return null;
  }
}

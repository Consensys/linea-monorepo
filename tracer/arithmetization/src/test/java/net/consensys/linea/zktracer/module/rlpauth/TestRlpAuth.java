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

package net.consensys.linea.zktracer.module.rlpauth;

import static net.consensys.linea.zktracer.Trace.LINEA_CHAIN_ID;
import static net.consensys.linea.zktracer.module.ecdata.EcDataOperation.SECP256K1N;

import java.math.BigInteger;
import java.util.List;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

@ExtendWith(UnitTestWatcher.class)
public class TestRlpAuth extends TracerTestBase {

  // Cases where sender = authority
  final long SENDER_NONCE = 42;
  final long AUTHORITY_NONCE = 1337L;
  final long DELEGATION_NONCE = 69;
  final long RECIPIENT_NONCE = 0xc0ffeeL;

  final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  final Address senderAddress =
      Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
  final ToyAccount senderAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(SENDER_NONCE)
          .address(senderAddress)
          .build();

  final KeyPair authorityKeyPair = new SECP256K1().generateKeyPair();
  final Address authorityAddress =
      Address.extract(Hash.hash(authorityKeyPair.getPublicKey().getEncodedBytes()));
  final ToyAccount authorityAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(2))
          .nonce(AUTHORITY_NONCE)
          .address(authorityAddress)
          .build();

  final KeyPair delegationKeyPair = new SECP256K1().generateKeyPair();
  final Address delegationAddress =
      Address.extract(Hash.hash(delegationKeyPair.getPublicKey().getEncodedBytes()));
  final ToyAccount delegationAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(3))
          .nonce(DELEGATION_NONCE)
          .address(delegationAddress)
          .build();

  final KeyPair recipientSmcKeyPair = new SECP256K1().generateKeyPair();
  final Address recipientSmcAddress =
      Address.extract(Hash.hash(recipientSmcKeyPair.getPublicKey().getEncodedBytes()));
  final ToyAccount recipientAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(4))
          .nonce(RECIPIENT_NONCE)
          .address(recipientSmcAddress)
          .code(Bytes.fromHexString("0x5b")) // nontrivial code that does nothing
          .build();

  @ParameterizedTest
  @ValueSource(longs = {AUTHORITY_NONCE, AUTHORITY_NONCE + 1})
  void tupleNonceVsStateNonceTest(long nonceParam, TestInfo testInfo) {

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(senderKeyPair)
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(SENDER_NONCE)
            .addCodeDelegation(
                BigInteger.valueOf(LINEA_CHAIN_ID), delegationAddress, nonceParam, authorityKeyPair)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount, authorityAccount, delegationAccount, recipientAccount))
        .transaction(tx)
        .build()
        .run();
  }

  @ParameterizedTest
  @ValueSource(ints = {LINEA_CHAIN_ID, LINEA_CHAIN_ID + 1})
  void chainIdIsDifferentFromNetworkChainIdTest(int chainId, TestInfo testInfo) {

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(senderKeyPair)
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(SENDER_NONCE)
            .addCodeDelegation(
                BigInteger.valueOf(chainId), delegationAddress, AUTHORITY_NONCE, authorityKeyPair)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount, authorityAccount, delegationAccount, recipientAccount))
        .transaction(tx)
        .build()
        .run();
  }

  @ParameterizedTest
  @ValueSource(ints = {1, 2})
  void sIsGreaterThanHalfCurveOrderTest(int divisor, TestInfo testInfo) {

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(senderKeyPair)
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(SENDER_NONCE)
            .addCodeDelegation(
                BigInteger.valueOf(LINEA_CHAIN_ID),
                delegationAddress,
                AUTHORITY_NONCE,
                BigInteger.ZERO,
                SECP256K1N
                    .toBigInteger()
                    .divide(BigInteger.valueOf(divisor)), // TODO: import it from constants
                (byte) 0)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount, authorityAccount, delegationAccount, recipientAccount))
        .transaction(tx)
        .build()
        .run();
  }
}

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

package net.consensys.linea.zktracer.module.rlptxn;

import java.util.Random;

class TestRandomTxns {
  private final Random rnd = new Random(666);
  private static final int TEST_TX_COUNT = 200;

  //  @Test
  //  void test() {
  //    OpCodes.load();
  //    ToyWorld.ToyWorldBuilder world = ToyWorld.builder();
  //    List<Transaction> txList = new ArrayList<>();
  //
  //    for (int i = 0; i < TEST_TX_COUNT; i++) {
  //      KeyPair keyPair = new SECP256K1().generateKeyPair();
  //      Address senderAddress =
  // Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
  //      ToyAccount senderAccount = randToyAccount(senderAddress);
  //      ToyAccount receiverAccount = receiverAccount();
  //
  //      world.account(senderAccount).account(receiverAccount);
  //      txList.add(randTx(senderAccount, keyPair, receiverAccount));
  //    }
  //    ToyExecutionEnvironment.builder()
  //        .toyWorld(world.build())
  //        .transactions(txList)
  //
  // .transactionProcessingResultValidator(TransactionProcessingResultValidator.EMPTY_VALIDATOR)
  //        .build()
  //        .run();
  //  }
  //
  //  final Transaction randTx(ToyAccount senderAccount, KeyPair keyPair, ToyAccount
  // receiverAccount) {
  //
  //    int txType = rnd.nextInt(0, 6);
  //
  //    return switch (txType) {
  //      case 0 -> ToyTransaction.builder()
  //          .sender(senderAccount)
  //          .keyPair(keyPair)
  //          .transactionType(TransactionType.FRONTIER)
  //          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
  //          .value(Wei.of(randBigInt(true)))
  //          .payload(randData(false))
  //          .build();
  //
  //      case 1 -> ToyTransaction.builder()
  //          .sender(senderAccount)
  //          .keyPair(keyPair)
  //          .transactionType(TransactionType.FRONTIER)
  //          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
  //          .value(Wei.of(randBigInt(true)))
  //          .to(receiverAccount)
  //          .payload(randData(false))
  //          .build();
  //
  //      case 2 -> ToyTransaction.builder()
  //          .sender(senderAccount)
  //          .keyPair(keyPair)
  //          .transactionType(TransactionType.ACCESS_LIST)
  //          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
  //          .value(Wei.of(randLong()))
  //          .payload(randData(false))
  //          .accessList(randAccessList())
  //          .build();
  //
  //      case 3 -> ToyTransaction.builder()
  //          .sender(senderAccount)
  //          .keyPair(keyPair)
  //          .transactionType(TransactionType.ACCESS_LIST)
  //          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
  //          .value(Wei.of(randLong()))
  //          .to(receiverAccount)
  //          .payload(randData(false))
  //          .accessList(randAccessList())
  //          .build();
  //
  //      case 4 -> ToyTransaction.builder()
  //          .sender(senderAccount)
  //          .keyPair(keyPair)
  //          .transactionType(TransactionType.EIP1559)
  //          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
  //          .value(Wei.of(randLong()))
  //          .payload(randData(false))
  //          .accessList(randAccessList())
  //          .build();
  //
  //      case 5 -> ToyTransaction.builder()
  //          .sender(senderAccount)
  //          .keyPair(keyPair)
  //          .transactionType(TransactionType.EIP1559)
  //          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
  //          .value(Wei.of(randLong()))
  //          .to(receiverAccount)
  //          .payload(randData(false))
  //          .accessList(randAccessList())
  //          .build();
  //
  //      default -> throw new IllegalStateException("Unexpected value: " + txType);
  //    };
  //  }
  //
  //  final List<AccessListEntry> randAccessList() {
  //    List<AccessListEntry> accessList = new ArrayList<>();
  //    boolean entries = rnd.nextBoolean();
  //    if (entries) {
  //      for (int i = 1; i < 25; i++) {
  //        accessList.add(randAccessListEntry());
  //      }
  //    }
  //    return accessList;
  //  }
  //
  //  final AccessListEntry randAccessListEntry() {
  //    List<Bytes32> keyList = new ArrayList<>();
  //    boolean key = rnd.nextBoolean();
  //    if (key) {
  //      for (int nKey = 1; nKey < rnd.nextInt(1, 20); nKey++) {
  //        keyList.add(Bytes32.random(rnd));
  //      }
  //    }
  //    return new AccessListEntry(Address.wrap(Bytes.random(20, rnd)), keyList);
  //  }
  //
  //  final ToyAccount receiverAccount() {
  //
  //    return ToyAccount.builder()
  //        .balance(Wei.v_ONE)
  //        .nonce(6)
  //        .address(Address.wrap(Bytes.random(20, rnd)))
  //        .code(
  //            BytecodeCompiler.newProgram()
  //                .push(32, 0xbeef)
  //                .push(32, 0xdead)
  //                .op(OpCode.ADD)
  //                .compile())
  //        .build();
  //  }
  //
  //  final ToyAccount randToyAccount(Address senderAddress) {
  //
  //    return ToyAccount.builder()
  //        .balance(Wei.wrap(Bytes.random(16, rnd)))
  //        .nonce(randLong())
  //        .address(senderAddress)
  //        .build();
  //  }
}

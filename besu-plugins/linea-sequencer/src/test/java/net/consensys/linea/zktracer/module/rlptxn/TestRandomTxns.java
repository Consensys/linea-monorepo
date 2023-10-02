/*
 * Copyright ConsenSys AG.
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

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodes;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.ToyAccount;
import net.consensys.linea.zktracer.testing.ToyExecutionEnvironment;
import net.consensys.linea.zktracer.testing.ToyTransaction;
import net.consensys.linea.zktracer.testing.ToyWorld;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;

class TestRandomTxns {
  private final Random rnd = new Random(666);

  @Test
  void test() {
    OpCodes.load();
    ToyWorld.ToyWorldBuilder world = ToyWorld.builder();
    List<Transaction> txList = new ArrayList<>();

    for (int i = 0; i < 200; i++) {
      KeyPair keyPair = new SECP256K1().generateKeyPair();
      Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
      ToyAccount senderAccount = randToyAccount(senderAddress);
      ToyAccount receiverAccount = receiverAccount();

      world.account(senderAccount).account(receiverAccount);
      txList.add(randTx(senderAccount, keyPair, receiverAccount));
    }
    ToyExecutionEnvironment.builder()
        .toyWorld(world.build())
        .transactions(txList)
        .testValidator(x -> {})
        .build()
        .run();
  }

  final Transaction randTx(ToyAccount senderAccount, KeyPair keyPair, ToyAccount receiverAccount) {

    int txType = rnd.nextInt(0, 6);

    return switch (txType) {
      case 0 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.FRONTIER)
          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
          .value(Wei.of(randBigInt(true)))
          .payload(randData())
          .build();

      case 1 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.FRONTIER)
          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
          .value(Wei.of(randBigInt(true)))
          .to(receiverAccount)
          .payload(randData())
          .build();

      case 2 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.ACCESS_LIST)
          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
          .value(Wei.of(randLong()))
          .payload(randData())
          .accessList(randAccessList())
          .build();

      case 3 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.ACCESS_LIST)
          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
          .value(Wei.of(randLong()))
          .to(receiverAccount)
          .payload(randData())
          .accessList(randAccessList())
          .build();

      case 4 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.EIP1559)
          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
          .value(Wei.of(randLong()))
          .payload(randData())
          .accessList(randAccessList())
          .build();

      case 5 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.EIP1559)
          .gasLimit(rnd.nextLong(21000, 0xfffffffffffffL))
          .value(Wei.of(randLong()))
          .to(receiverAccount)
          .payload(randData())
          .accessList(randAccessList())
          .build();

      default -> throw new IllegalStateException("Unexpected value: " + txType);
    };
  }

  final BigInteger randBigInt(boolean onlySixteenByte) {
    int selectorBound = 4;
    if (!onlySixteenByte) {
      selectorBound += 1;
    }
    int selector = rnd.nextInt(0, selectorBound);

    return switch (selector) {
      case 0 -> BigInteger.ZERO;
      case 1 -> BigInteger.valueOf(rnd.nextInt(1, 128));
      case 2 -> BigInteger.valueOf(rnd.nextInt(128, 256));
      case 3 -> new BigInteger(16 * 8, rnd);
      case 4 -> new BigInteger(32 * 8, rnd);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }

  final Bytes randData() {
    int selector = rnd.nextInt(0, 6);
    return switch (selector) {
      case 0 -> Bytes.EMPTY;
      case 1 -> Bytes.of(0x0);
      case 2 -> Bytes.minimalBytes(rnd.nextLong(1, 128));
      case 3 -> Bytes.minimalBytes(rnd.nextLong(128, 256));
      case 4 -> Bytes.random(rnd.nextInt(1, 56), rnd);
      case 5 -> Bytes.random(rnd.nextInt(56, 666), rnd);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }

  final List<AccessListEntry> randAccessList() {
    List<AccessListEntry> accessList = new ArrayList<>();
    boolean entries = rnd.nextBoolean();
    if (entries) {
      for (int i = 1; i < 25; i++) {
        accessList.add(randAccessListEntry());
      }
    }
    return accessList;
  }

  final AccessListEntry randAccessListEntry() {
    List<Bytes32> keyList = new ArrayList<>();
    boolean key = rnd.nextBoolean();
    if (key) {
      for (int nKey = 1; nKey < rnd.nextInt(1, 20); nKey++) {
        keyList.add(Bytes32.random(rnd));
      }
    }
    return new AccessListEntry(Address.wrap(Bytes.random(20, rnd)), keyList);
  }

  final Long randLong() {
    int selector = rnd.nextInt(0, 4);
    return switch (selector) {
      case 0 -> 0L;
      case 1 -> rnd.nextLong(1, 128);
      case 2 -> rnd.nextLong(128, 256);
      case 3 -> rnd.nextLong(256, 0xfffffffffffffffL);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }

  final ToyAccount receiverAccount() {

    return ToyAccount.builder()
        .balance(Wei.ONE)
        .nonce(6)
        .address(Address.wrap(Bytes.random(20, rnd)))
        .code(
            BytecodeCompiler.newProgram()
                .push(32, 0xbeef)
                .push(32, 0xdead)
                .op(OpCode.ADD)
                .compile())
        .build();
  }

  final ToyAccount randToyAccount(Address senderAddress) {

    return ToyAccount.builder()
        .balance(Wei.MAX_WEI)
        .nonce(randLong())
        .address(senderAddress)
        .build();
  }
}

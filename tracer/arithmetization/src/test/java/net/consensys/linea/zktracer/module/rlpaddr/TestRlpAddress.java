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

package net.consensys.linea.zktracer.module.rlpaddr;

import static net.consensys.linea.zktracer.module.rlpcommon.RlpRandEdgeCase.randBigInt;
import static net.consensys.linea.zktracer.module.rlpcommon.RlpRandEdgeCase.randData;
import static net.consensys.linea.zktracer.module.rlpcommon.RlpRandEdgeCase.randLong;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

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
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;

public class TestRlpAddress {
  private static final int TEST_TX_COUNT = 30;
  private final Random rnd = new Random(666);

  @Test
  void test() {
    OpCodes.load();

    ToyWorld.ToyWorldBuilder world = ToyWorld.builder();
    List<Transaction> txList = new ArrayList<>();

    for (int i = 0; i < TEST_TX_COUNT; i++) {
      KeyPair keyPair = new SECP256K1().generateKeyPair();
      Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
      ToyAccount senderAccount = randSenderAccount(senderAddress);

      Bytes initCode = randData(true);
      int initCodeSize = initCode.size();

      int trigger = rnd.nextInt(0, 3);
      ToyAccount receiverAccount;
      switch (trigger) {
        case 0 -> { // create with tx.To = null
          world.account(senderAccount);
          txList.add(
              ToyTransaction.builder()
                  .sender(senderAccount)
                  .keyPair(keyPair)
                  .transactionType(TransactionType.FRONTIER)
                  .gasLimit(rnd.nextLong(21000, 0xffffL))
                  .payload(initCode)
                  .build());
        }
        case 1 -> { // Create OpCode
          receiverAccount = randCreate(initCodeSize);
          world.account(senderAccount).account(receiverAccount);
          txList.add(
              ToyTransaction.builder()
                  .sender(senderAccount)
                  .keyPair(keyPair)
                  .to(receiverAccount)
                  .transactionType(TransactionType.FRONTIER)
                  .gasLimit(rnd.nextLong(21000, 0xffffL))
                  .build());
        }
        case 2 -> { // Create2 OpCode
          receiverAccount = randCreateTwo(initCodeSize);
          world.account(senderAccount).account(receiverAccount);
          txList.add(
              ToyTransaction.builder()
                  .sender(senderAccount)
                  .keyPair(keyPair)
                  .to(receiverAccount)
                  .transactionType(TransactionType.FRONTIER)
                  .gasLimit(rnd.nextLong(21000, 0xffffL))
                  .build());
        }
      }
    }

    ToyExecutionEnvironment.builder()
        .toyWorld(world.build())
        .transactions(txList)
        .testValidator(x -> {})
        .build()
        .run();
  }

  private ToyAccount randCreate(int initCodeSize) {
    byte[] value = bigIntegerToBytes(BigInteger.valueOf(randLong())).toArray();
    return ToyAccount.builder()
        .balance(Wei.fromEth(1000))
        .nonce(randLong())
        .address(Address.wrap(Bytes.repeat((byte) 0x01, 20)))
        .code(
            BytecodeCompiler.newProgram()
                .push(initCodeSize)
                .push(1)
                .push(1)
                .op(OpCode.CALLDATACOPY)
                .push(initCodeSize)
                .push(1)
                .push(value.length, value)
                .op(OpCode.CREATE)
                .compile())
        .build();
  }

  private ToyAccount randCreateTwo(int initCodeSize) {
    byte[] salt = bigIntegerToBytes(randBigInt(false)).toArray();
    byte[] value = bigIntegerToBytes(BigInteger.valueOf(randLong())).toArray();

    return ToyAccount.builder()
        .balance(Wei.fromEth(1000))
        .nonce(randLong())
        .address(Address.wrap(Bytes.repeat((byte) 0x02, 20)))
        .code(
            BytecodeCompiler.newProgram()
                .push(initCodeSize)
                .push(1)
                .push(1)
                .op(OpCode.CALLDATACOPY)
                .push(salt.length, salt)
                .push(1)
                .push(1)
                .push(value.length, value)
                .op(OpCode.CREATE2)
                .compile())
        .build();
  }

  final ToyAccount randSenderAccount(Address senderAddress) {
    return ToyAccount.builder()
        .balance(Wei.of(randBigInt(true)))
        .nonce(randLong())
        .address(senderAddress)
        .build();
  }
}

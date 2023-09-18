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

class RandomTxn {

  @Test
  void test() {
    OpCodes.load();
    KeyPair keyPair = new SECP256K1().generateKeyPair();
    // Bytes32 bytes32 = Bytes32.repeat((byte) 1);
    // SECPPrivateKey privateKey = new SECP256K1().createPrivateKey(bytes32);
    // KeyPair keyPair = new SECP256K1().createKeyPair(privateKey);
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));

    long randLongSmall = new Random().nextLong(1, 128);
    long randLongMedium = new Random().nextLong(128, 256);
    long randLongLong = new Random().nextLong(128, 0xfffffffffffffffL);
    BigInteger randBigIntSmall = new BigInteger(7, new Random());
    BigInteger randBigIntSixteenBytes = new BigInteger(16 * 8, new Random());
    BigInteger randBigIntThirtyTwoBytes = new BigInteger(32 * 8, new Random());
    int randIntLEFiveFive = new Random().nextInt(2, 55);
    int randIntGEFiveSix = new Random().nextInt(56, 1234);

    ToyAccount senderAccount =
        ToyAccount.builder()
            .balance(Wei.of(5))

            // Choose the value of the sender's nonce
            .nonce(0)
            // .nonce(randLongSmall)
            // .nonce(randLongLong)

            .address(senderAddress)
            .build();

    ToyAccount receiverAccount =
        ToyAccount.builder()
            .balance(Wei.ONE)
            .nonce(6)

            // Choose the receiver's address
            // .address(Address.fromHexString("0x00112233445566778899aabbccddeeff00112233"))
            // .address(Address.fromHexString("0x0"))
            .address(Address.wrap(Bytes.random(20)))
            .code(
                BytecodeCompiler.newProgram()
                    .push(32, 0xbeef)
                    .push(32, 0xdead)
                    .op(OpCode.ADD)
                    .compile())
            .build();

    Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .keyPair(keyPair)
            //            .to(receiverAccount)

            // Choose the type of transaction
            .transactionType(TransactionType.FRONTIER)
            // .transactionType(TransactionType.ACCESS_LIST)
            // .transactionType(TransactionType.EIP1559)

            // Choose the value of GasLimit
            // .gasLimit(0L)
            // .gasLimit(randLongSmall)
            .gasLimit(randLongLong)

            // Choose the value of the value
            .value(Wei.of(BigInteger.ZERO))
            // .value(Wei.of(randBigIntSmall))
            // .value(Wei.of(randBigIntSixteenBytes))

            // Choose the data
            .payload(Bytes.EMPTY)
            // .payload(Bytes.minimalBytes(randLongSmall))
            // .payload(Bytes.minimalBytes(randLongMedium))
            // .payload(Bytes.random(randIntLEFiveFive))
            // .payload(Bytes.random(randIntGEFiveSix))
            // .payload(Bytes.random(140))
            .build();

    ToyWorld toyWorld =
        ToyWorld.builder().accounts(List.of(senderAccount, receiverAccount)).build();

    ToyExecutionEnvironment.builder().toyWorld(toyWorld).transaction(tx).build().run();
  }
}

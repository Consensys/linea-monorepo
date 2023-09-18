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

package net.consensys.linea.zktracer.testing;

import java.util.List;

import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(EvmExtension.class)
class ExampleTxTest {

  @Test
  void test() {
    KeyPair keyPair = new SECP256K1().generateKeyPair();
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));

    ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.of(5)).nonce(5).address(senderAddress).build();

    ToyAccount receiverAccount =
        ToyAccount.builder()
            .balance(Wei.ONE)
            .nonce(6)
            .address(Address.fromHexString("0x111111"))
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
            .to(receiverAccount)
            //            .transactionType(TransactionType.ACCESS_LIST)
            //            .accessList(List.of(new AccessListEntry(Address.fromHexString("0x00"),
            // List.of())))
            .keyPair(keyPair)
            .build();

    ToyWorld toyWorld =
        ToyWorld.builder().accounts(List.of(senderAccount, receiverAccount)).build();

    ToyExecutionEnvironment.builder().toyWorld(toyWorld).transaction(tx).build().run();
  }
}

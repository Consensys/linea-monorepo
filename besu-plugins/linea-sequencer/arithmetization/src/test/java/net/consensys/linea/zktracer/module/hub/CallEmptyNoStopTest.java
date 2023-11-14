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

package net.consensys.linea.zktracer.module.hub;

import java.util.List;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.testing.EvmExtension;
import net.consensys.linea.zktracer.testing.ToyAccount;
import net.consensys.linea.zktracer.testing.ToyExecutionEnvironment;
import net.consensys.linea.zktracer.testing.ToyTransaction;
import net.consensys.linea.zktracer.testing.ToyWorld;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

/** Ensure that calling a contract with empty code does not generate a virtual STOP trace */
@ExtendWith(EvmExtension.class)
public class CallEmptyNoStopTest {
  @Test
  void test() {
    KeyPair keyPair = new SECP256K1().generateKeyPair();
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));

    ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.fromEth(1)).nonce(5).address(senderAddress).build();

    ToyAccount receiverAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(1))
            .nonce(6)
            .address(Address.fromHexString("0x111111"))
            .code(
                BytecodeCompiler.newProgram()
                    .push(0) // retSize
                    .push(0) // retOffset
                    .push(0) // argsSize
                    .push(0) // argsOffset
                    .push(10) // value
                    .push(0x222222) // address
                    .push(10000) // gas
                    .op(OpCode.CALL)
                    .compile())
            .build();

    ToyAccount emptyCodeAccount =
        ToyAccount.builder()
            .balance(Wei.ONE)
            .nonce(1)
            .address(Address.fromHexString("0x222222"))
            .code(Bytes.EMPTY)
            .build();

    Transaction tx =
        ToyTransaction.builder().sender(senderAccount).to(receiverAccount).keyPair(keyPair).build();

    ToyWorld toyWorld =
        ToyWorld.builder()
            .accounts(List.of(senderAccount, receiverAccount, emptyCodeAccount))
            .build();

    ToyExecutionEnvironment.builder()
        .toyWorld(toyWorld)
        .transaction(tx)
        .zkTracerValidator(
            zkTracer -> {
              // Ensure we don't have any superfluous STOP
              assert zkTracer.getHub().state().currentTxTrace().getTrace().size() == 11;
            })
        .build()
        .run();
  }
}

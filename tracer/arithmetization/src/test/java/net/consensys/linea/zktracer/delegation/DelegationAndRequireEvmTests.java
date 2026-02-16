/*
 * Copyright ConsenSys Inc.
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

import static io.opentelemetry.api.internal.Utils.checkArgument;
import static net.consensys.linea.zktracer.module.hub.AccountSnapshot.EIP_7702_DELEGATION_INDICATOR_BYTES;
import static net.consensys.linea.zktracer.module.hub.AccountSnapshot.isDelegation;

import java.util.List;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class DelegationAndRequireEvmTests extends TracerTestBase {

  @Disabled
  @Test
  void delegationAndRequireEvmTests(TestInfo testInfo) {
    final KeyPair keyPair = new SECP256K1().generateKeyPair();
    final Address senderAddress =
        Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
    final ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.fromEth(1789)).nonce(0).address(senderAddress).build();

    final ToyAccount recipientAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(1))
            .nonce(0)
            .address(Address.fromHexString("0x1122334455667788990011223344556677889900"))
            .build();

    final ToyAccount eoaAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(1))
            .nonce(0)
            .address(Address.fromHexString("0x0e0ae0ae0ae0ae0ae0ae0ae0ae0ae0ae0ae0ae0a"))
            .build();

    final ToyAccount smcAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(1))
            .nonce(0)
            .address(Address.fromHexString("0x1111111111111111111111111111111111111111"))
            // random code, we don't care what's happening
            .code(
                BytecodeCompiler.newProgram(chainConfig)
                    .op(OpCode.PUSH0)
                    .push(2)
                    .op(OpCode.MUL)
                    .op(OpCode.PUSH0)
                    .op(OpCode.TSTORE)
                    .compile())
            .build();

    final ToyAccount delegatedAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(1))
            .nonce(0)
            .address(Address.fromHexString("0xde7e6a7ed0de7e6a7ed0de7e6a7ed0de7e6a7ed0"))
            .code(Bytes.concatenate(EIP_7702_DELEGATION_INDICATOR_BYTES, smcAccount.getAddress()))
            .build();

    checkArgument(
        isDelegation(delegatedAccount.getCode()), "This account is supposed to be delegated");

    // tx1, no evm to evm because recipîent is getting delegated
    final Transaction tx1 =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(keyPair)
            .value(Wei.of(123))
            .gasLimit(1000000L)
            .transactionType(TransactionType.DELEGATE_CODE)
            // add some stupid delegation
            .addCodeDelegation(chainConfig.id, smcAccount.getAddress(), 0, keyPair)
            .build();

    // tx2, evm to no evm recipîent is from delegated to smc to an eoa
    final Transaction tx2 =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(keyPair)
            .value(Wei.of(123))
            .gasLimit(1000000L)
            .transactionType(TransactionType.DELEGATE_CODE)
            // add some stupid delegation
            .addCodeDelegation(chainConfig.id, eoaAccount.getAddress(), 0, keyPair)
            .build();

    // tx3, evm to no evm recipîent is from delegated to smc to an eoa
    final Transaction tx3 =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(keyPair)
            .value(Wei.of(123))
            .gasLimit(1000000L)
            .transactionType(TransactionType.DELEGATE_CODE)
            // add some stupid delegation
            .addCodeDelegation(chainConfig.id, eoaAccount.getAddress(), 0, keyPair)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(
            List.of(senderAccount, recipientAccount, eoaAccount, delegatedAccount, smcAccount))
        .transactions(List.of(tx1, tx2, tx3))
        .zkTracerValidator(zkTracer -> {})
        .build()
        .run();
  }
}

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

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;
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
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

public class DelegatesAndRevertsTests extends TracerTestBase {

  /**
   * We require tests like so: mono transaction block contains a single type 4 transaction.
   *
   * <p>this tx has 1 delegation with all combinations of the following
   *
   * <ul>
   *   <li>delegations are valid: <b>[yes / no]</b>
   *   <li>(for valid delegations) authority exists: <b>[yes / no]</b>
   *   <li>TX_REQUIRES_EVM_EXECUTION: <b>[yes / no]</b>
   *   <li>If yes: tx
   *       <ul>
   *         <li>reverts: <b>[yes / no]</b>
   *         <li>tx incurs another refund (say: a single SSTORE that resets storage): <b>[yes /
   *             no]</b>
   *       </ul>
   *   <li>if no: no further requirements
   * </ul>
   */
  @ParameterizedTest
  @MethodSource("delegatesAndRevertsTestsSource")
  void delegatesAndRevertsTest(scenario sc, TestInfo testInfo) {

    // sender account
    final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
    final Address senderAddress =
        Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
    final ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.fromEth(56)).nonce(119).address(senderAddress).build();

    // authority
    final KeyPair authorityKeyPair = new SECP256K1().generateKeyPair();
    final Address authorityAddress =
        Address.extract(Hash.hash(authorityKeyPair.getPublicKey().getEncodedBytes()));
    final long authNonce = 16454;
    final ToyAccount authorityAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(2))
            .nonce(authNonce)
            .address(authorityAddress)
            .build();

    // SMC
    final Address smcAddress = Address.fromHexString("0x1122334455667788990011223344556677889900");
    final ToyAccount smcAccount =
        ToyAccount.builder().balance(Wei.fromEth(22)).nonce(3).address(smcAddress).build();

    final ToyTransaction.ToyTransactionBuilder tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(authorityAccount)
            .keyPair(senderKeyPair)
            .gasLimit(300000L)
            .transactionType(TransactionType.DELEGATE_CODE)
            .value(Wei.of(1000));

    final List<ToyAccount> accountsInTheWorld = new ArrayList<>();
    accountsInTheWorld.add(senderAccount);
    accountsInTheWorld.add(authorityAccount);

    if (sc == scenario.DELEGATION_VALID___NON) {
      // delegation is known to fail because of chainID, signature, etc
      tx.addCodeDelegation(
          chainConfig.id.and(
              Bytes.fromHexString("0x17891789178917891789178917891789178917891789178900000000")
                  .toUnsignedBigInteger()),
          Address.ZERO,
          0L,
          BigInteger.valueOf(78),
          BigInteger.valueOf(89),
          (byte) 0);
    } else {
      tx.addCodeDelegation(chainConfig.id, smcAddress, authNonce, authorityKeyPair);
      if (sc != scenario.DELEGATION_VALID___OUI___AUTHORITY_EXIST___NON) {
        accountsInTheWorld.add(smcAccount);
        if (sc != scenario.DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___NON) {
          switch (sc) {
            case
              DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___NON___OTHER_REFUNDS___NON -> {
              smcAccount.setCode(
                  BytecodeCompiler.newProgram(chainConfig)
                      .push(1)
                      .push(2)
                      .push(3)
                      .op(OpCode.ADDMOD)
                      .op(OpCode.POP)
                      .compile());
            }
            case
              DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___NON___OTHER_REFUNDS___OUI -> {
              smcAccount.setCode(
                BytecodeCompiler.newProgram(chainConfig)
                  .push(0x11111111)
                  .push(0x5107) // 0x 5107 <> slot
                  .op(OpCode.SSTORE) // write nontrivial value
                  .push(0)
                  .push(0x5107)
                  .op(OpCode.SSTORE) // incur refund (reset to zero)
                  .push(0)
                  .push(0)
                  .op(OpCode.REVERT)
                  .compile());
            }
            case
              DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___OUI___OTHER_REFUNDS___NON -> {
              smcAccount.setCode(
                  BytecodeCompiler.newProgram(chainConfig)
                      .push(1)
                      .push(2)
                      .push(3)
                      .op(OpCode.ADDMOD)
                      .op(OpCode.POP)
                      .push(0)
                      .push(0)
                      .op(OpCode.REVERT)
                      .compile());
            }
            case
              DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___OUI___OTHER_REFUNDS___OUI -> {
              smcAccount.setCode(
                  BytecodeCompiler.newProgram(chainConfig)
                      .push(0x11111111)
                      .push(0x5107)
                      .op(OpCode.SSTORE) // write nontrivial value
                      .push(0)
                      .push(0x5107)
                      .op(OpCode.SSTORE) // incur refund (reset to zero)
                      .push(0)
                      .push(0)
                      .op(OpCode.REVERT)
                      .compile());
            }
            default -> throw new IllegalArgumentException("Unknown scenario:" + sc);
          }
        }
      }
    }

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(accountsInTheWorld)
        .transaction(tx.build())
        .zkTracerValidator(zkTracer -> {})
        .build()
        .run();
  }

  private static Stream<Arguments> delegatesAndRevertsTestsSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (scenario sc1 : scenario.values()) {
      arguments.add(Arguments.of(sc1));
    }
    return arguments.stream();
  }

  private enum scenario {
    DELEGATION_VALID___NON,
    DELEGATION_VALID___OUI___AUTHORITY_EXIST___NON,
    DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___NON,
    DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___NON___OTHER_REFUNDS___NON,
    DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___NON___OTHER_REFUNDS___OUI,
    DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___OUI___OTHER_REFUNDS___NON,
    DELEGATION_VALID___OUI___AUTHORITY_EXIST___OUI___EVM_EXECUTION___OUI___REVERTS___OUI___OTHER_REFUNDS___OUI
  }
}

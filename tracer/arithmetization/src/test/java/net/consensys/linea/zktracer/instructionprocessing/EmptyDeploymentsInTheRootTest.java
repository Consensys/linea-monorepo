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
package net.consensys.linea.zktracer.instructionprocessing;

import java.util.List;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.testing.TransactionProcessingResultValidator;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 * The purpose of {@link EmptyDeploymentsInTheRootTest} is to make sure that deployment transactions
 * work and lead to the actual deployment of bytecode. In particular the final RETURN instruction
 * (if present) should be treated properly and the update to the deployment account is accounted for
 * in the relevant row.
 */
@ExtendWith(UnitTestWatcher.class)
public class EmptyDeploymentsInTheRootTest {

  final Bytes initCodeEmptyDeployment =
      BytecodeCompiler.newProgram()
          .push(0) // size
          .push(0) // offset
          .op(OpCode.RETURN)
          .compile();

  final Bytes initCodeNonemptyDeployment =
      BytecodeCompiler.newProgram()
          .op(OpCode.TIMESTAMP) // value, initially was DIFFICULTY
          .push(0) // offset
          .op(OpCode.MSTORE)
          .push(32) // size
          .push(0) // offset
          .op(OpCode.RETURN)
          .compile();

  final Bytes initCodeEmptyRevert =
      BytecodeCompiler.newProgram()
          .push(0) // size
          .push(0) // offset
          .op(OpCode.REVERT)
          .compile();

  final Bytes initCodeNonemptyRevert =
      BytecodeCompiler.newProgram()
          .op(OpCode.BLOCKHASH) // value
          .push(0) // offset
          .op(OpCode.MSTORE)
          .push(32) // size
          .push(0) // offset
          .op(OpCode.REVERT)
          .compile();

  final Bytes deployerOfEmptyDeploymentInitCode = deployerOf(initCodeEmptyDeployment);
  final Bytes deployerOfNonemptyDeploymentInitCode = deployerOf(initCodeNonemptyDeployment);
  final Bytes deployerOfEmptyRevert = deployerOf(initCodeEmptyRevert);
  final Bytes deployerOfNonemptyRevert = deployerOf(initCodeNonemptyRevert);

  /** We test <b>deployment transactions</b>. */
  @Test
  void deploymentTransactionLeadsToEmptyDeploymentTest() {
    Transaction deploymentTransaction = deploymentTansactionFromInitCode(initCodeEmptyDeployment);
    runTransaction(deploymentTransaction);
  }

  @Test
  void deploymentTransactionLeadsToNonemptyDeploymentTest() {
    Transaction deploymentTransaction =
        deploymentTansactionFromInitCode(initCodeNonemptyDeployment);
    runTransaction(deploymentTransaction);
  }

  @Test
  void deploymentTransactionEmptyReverts() {
    Transaction deploymentTransaction = deploymentTansactionFromInitCode(initCodeEmptyRevert);
    runTransaction(deploymentTransaction);
  }

  @Test
  void deploymentTransactionNonemptyReverts() {
    Transaction deploymentTransaction = deploymentTansactionFromInitCode(initCodeNonemptyRevert);
    runTransaction(deploymentTransaction);
  }

  /** We test <b>CREATE's</b>. */
  @Test
  void createDeploysEmptyByteCode() {
    Transaction messageCallTransaction =
        messageCallTransactionToDeployerAccount(accountDeployerOfEmptyInitCode);
    runTransaction(messageCallTransaction);
  }

  @Test
  void createDeploysNonemptyByteCode() {
    Transaction messageCallTransaction =
        messageCallTransactionToDeployerAccount(accountDeployerOfNonemptyInitCode);
    runTransaction(messageCallTransaction);
  }

  @Test
  void createRevertsWithEmptyReturnData() {
    Transaction messageCallTransaction =
        messageCallTransactionToDeployerAccount(
            accountGeneratorOfRevertedCreateWithEmptyReturnData);
    runTransaction(messageCallTransaction);
  }

  @Test
  void createRevertsWithNonemptyReturnData() {
    Transaction messageCallTransaction =
        messageCallTransactionToDeployerAccount(
            accountGeneratorOfRevertedCreateWithNonemptyReturnData);
    runTransaction(messageCallTransaction);
  }

  KeyPair keyPair = new SECP256K1().generateKeyPair();
  Address userAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));
  ToyAccount userAccount =
      ToyAccount.builder().balance(Wei.fromEth(100)).nonce(1).address(userAddress).build();
  ToyAccount accountDeployerOfEmptyInitCode =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(13)
          .address(Address.fromHexString("0x1337"))
          .code(deployerOfEmptyDeploymentInitCode)
          .build();
  ToyAccount accountDeployerOfNonemptyInitCode =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(81)
          .address(Address.fromHexString("0xadd7e550"))
          .code(deployerOfNonemptyDeploymentInitCode)
          .build();
  ToyAccount accountGeneratorOfRevertedCreateWithEmptyReturnData =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(255)
          .address(Address.fromHexString("0x69420"))
          .code(deployerOfEmptyRevert)
          .build();
  ToyAccount accountGeneratorOfRevertedCreateWithNonemptyReturnData =
      ToyAccount.builder()
          .balance(Wei.fromEth(1))
          .nonce(1024)
          .address(Address.fromHexString("0xdeadbeef"))
          .code(deployerOfNonemptyRevert)
          .build();

  List<ToyAccount> accounts =
      List.of(
          userAccount,
          accountDeployerOfEmptyInitCode,
          accountDeployerOfNonemptyInitCode,
          accountGeneratorOfRevertedCreateWithEmptyReturnData,
          accountGeneratorOfRevertedCreateWithNonemptyReturnData);

  Transaction deploymentTansactionFromInitCode(Bytes initCode) {
    return ToyTransaction.builder()
        .sender(userAccount)
        .payload(initCode)
        .transactionType(TransactionType.FRONTIER)
        .value(Wei.ONE)
        .keyPair(keyPair)
        .gasLimit(100_000L)
        .gasPrice(Wei.of(8))
        .build();
  }

  Transaction messageCallTransactionToDeployerAccount(ToyAccount deployerAccount) {
    return ToyTransaction.builder()
        .sender(userAccount)
        .to(deployerAccount)
        .transactionType(TransactionType.FRONTIER)
        .value(Wei.ONE)
        .keyPair(keyPair)
        .gasLimit(100_000L)
        .gasPrice(Wei.of(8))
        .build();
  }

  private void runTransaction(Transaction transaction) {
    ToyExecutionEnvironmentV2.builder()
        .accounts(accounts)
        .transaction(transaction)
        .transactionProcessingResultValidator(TransactionProcessingResultValidator.EMPTY_VALIDATOR)
        .build()
        .run();
  }

  /**
   * @param initCode assumed to fit on at most <b>32</b> bytes
   * @return
   */
  private Bytes deployerOf(Bytes initCode) {
    return BytecodeCompiler.newProgram()
        .push(initCode)
        .push(0) // offset
        .op(OpCode.MSTORE)
        .push(initCode.size()) // size
        .push(32 - initCode.size()) // offset
        .push(255) // value
        .op(OpCode.CREATE)
        .op(OpCode.EXTCODESIZE) // get code size of newly deployed smart contract
        .compile();
  }
}

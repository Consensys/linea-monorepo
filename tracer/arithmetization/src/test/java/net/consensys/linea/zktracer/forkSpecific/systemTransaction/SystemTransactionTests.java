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

package net.consensys.linea.zktracer.forkSpecific.systemTransaction;

import static net.consensys.linea.testing.MultiBlockExecutionEnvironment.DEFAULT_DELTA_TIMESTAMP_BETWEEN_BLOCKS;
import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.DEFAULT_BLOCK_NUMBER;
import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.DEFAULT_TIME_STAMP;
import static net.consensys.linea.zktracer.Fork.isPostShanghai;
import static net.consensys.linea.zktracer.forkSpecific.systemTransaction.SystemTransactionTestUtils.byteCodeCallingBeaconRootSystemAccount;
import static net.consensys.linea.zktracer.module.hub.section.systemTransaction.EIP2935HistoricalHash.EIP2935_HISTORY_STORAGE_ADDRESS;
import static net.consensys.linea.zktracer.module.hub.section.systemTransaction.EIP4788BeaconBlockRoot.EIP4788_BEACONROOT_ADDRESS;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.MultiBlockExecutionEnvironment;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyTransaction;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

public class SystemTransactionTests extends TracerTestBase {

  // This test checks the consistency of system account by calling the system account eip-2935 in
  // happy path: not genesis block, system address exists
  @Test
  void systemTransaction2935ConsistencyTest(TestInfo testInfo) {
    BytecodeRunner.of(
            byteCodeCallingBeaconRootSystemAccount(
                chainConfig, EIP2935_HISTORY_STORAGE_ADDRESS, DEFAULT_BLOCK_NUMBER - 1))
        .run(chainConfig, testInfo);
  }

  // This test checks the consistency of system account by calling the system account of eip-4788 in
  // happy path: not genesis block, system address exists
  @Test
  void systemTransaction4788ConsistencyTest(TestInfo testInfo) {
    BytecodeRunner.of(
            byteCodeCallingBeaconRootSystemAccount(
                chainConfig, EIP4788_BEACONROOT_ADDRESS, DEFAULT_TIME_STAMP))
        .run(chainConfig, testInfo);
  }

  private static Stream<Arguments> scenariiForSystemContract() {

    final List<Arguments> scenarii = new ArrayList<>();
    scenarii.add(Arguments.of(0, false, 0, false));
    for (int system2935ContractDeployedBeforeBlockNumber = 1;
        system2935ContractDeployedBeforeBlockNumber <= 2;
        system2935ContractDeployedBeforeBlockNumber++) {
      for (short valueTransferedPriorToDeploymentOf2935 = 0;
          valueTransferedPriorToDeploymentOf2935 <= 1;
          valueTransferedPriorToDeploymentOf2935++) {
        for (int system4788ContractDeployedBeforeBlockNumber = 1;
            system4788ContractDeployedBeforeBlockNumber <= 2;
            system4788ContractDeployedBeforeBlockNumber++) {
          for (short valueTransferedPriorToDeploymentOf4788 = 0;
              valueTransferedPriorToDeploymentOf4788 <= 1;
              valueTransferedPriorToDeploymentOf4788++) {
            scenarii.add(
                Arguments.of(
                    system2935ContractDeployedBeforeBlockNumber,
                    valueTransferedPriorToDeploymentOf2935 == 1,
                    system4788ContractDeployedBeforeBlockNumber,
                    valueTransferedPriorToDeploymentOf4788 == 1));
          }
        }
      }
    }
    return scenarii.stream();
  }

  Long senderNonce = 0L;
  // random sender account that will send a tx to system account
  private static final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  private static final Address senderAddress =
      Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
  private static final ToyAccount senderAccount =
      ToyAccount.builder()
          .balance(Wei.fromEth(123))
          .nonce(0)
          .address(senderAddress)
          .balance(Wei.fromEth(12))
          .build();

  // This EOA calls 3 times the system account with as input the three block number of the
  // conflation
  private static final ToyAccount callerOf2935 =
      ToyAccount.builder()
          .address(Address.wrap(leftPadTo(Bytes.fromHexString("0x2935"), Address.SIZE)))
          .code(
              Bytes.concatenate(
                  byteCodeCallingBeaconRootSystemAccount(
                      chainConfig, EIP2935_HISTORY_STORAGE_ADDRESS, 0),
                  byteCodeCallingBeaconRootSystemAccount(
                      chainConfig, EIP2935_HISTORY_STORAGE_ADDRESS, 1),
                  byteCodeCallingBeaconRootSystemAccount(
                      chainConfig, EIP2935_HISTORY_STORAGE_ADDRESS, 2)))
          .build();

  // This EOA calls 3 times the system account with as input the three block timestamp of the
  // conflation
  private static final ToyAccount callerOf4788 =
      ToyAccount.builder()
          .address(Address.wrap(leftPadTo(Bytes.fromHexString("0x4788"), Address.SIZE)))
          .code(
              Bytes.concatenate(
                  byteCodeCallingBeaconRootSystemAccount(
                      chainConfig, EIP4788_BEACONROOT_ADDRESS, DEFAULT_TIME_STAMP),
                  byteCodeCallingBeaconRootSystemAccount(
                      chainConfig,
                      EIP4788_BEACONROOT_ADDRESS,
                      DEFAULT_TIME_STAMP + DEFAULT_DELTA_TIMESTAMP_BETWEEN_BLOCKS),
                  byteCodeCallingBeaconRootSystemAccount(
                      chainConfig,
                      EIP4788_BEACONROOT_ADDRESS,
                      DEFAULT_TIME_STAMP + 2 * DEFAULT_DELTA_TIMESTAMP_BETWEEN_BLOCKS)))
          .build();

  @ParameterizedTest
  @MethodSource("scenariiForSystemContract")
  void genesisBlockTest(
      int system2935ContractDeployedBeforeBlockNumber,
      boolean valueTransferedPriorToDeploymentOf2935,
      int system4788ContractDeployedBeforeBlockNumber,
      boolean valueTransferedPriorToDeploymentOf4788,
      TestInfo testInfo) {

    // Note: this test uses a PUSH0 in the deployment code of the system contract. Therefore, the
    // deployment fails if not in Shanghai or after, and our test framework doesn't allow failing
    // transaction:
    // org.opentest4j.AssertionFailedError: Transaction not successful:
    if (!isPostShanghai(chainConfig.fork)) {
      return;
    }

    // Note: do not modify me: value coming from https://eips.ethereum.org/EIPS/eip-2935
    final ToyAccount deployerOf2935 =
        ToyAccount.builder()
            .nonce(0)
            .address(Address.fromHexString("0x3462413Af4609098e1E27A490f554f260213D685"))
            .balance(Wei.fromEth(109))
            .build();
    final Transaction deploy2935 =
        ToyTransaction.builder()
            .sender(deployerOf2935)
            .transactionType(TransactionType.FRONTIER)
            .gasLimit(0x3d090L)
            .gasPrice(Wei.fromHexString("0xe8d4a51000"))
            .payload(
                Bytes.fromHexString(
                    "0x60538060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604657602036036042575f35600143038111604257611fff81430311604257611fff9006545f5260205ff35b5f5ffd5b5f35611fff60014303065500"))
            .value(Wei.of(0))
            .signature(
                Bytes.concatenate(
                    Bytes32.fromHexStringLenient("0x539"), // r
                    Bytes32.fromHexStringLenient("0xaa12693182426612186309f02cfe8a80a0000"), // s
                    Bytes.fromHexString("0x00") // v = 0x1B
                    ))
            .build();

    // Note: do not modify me: value coming from https://eips.ethereum.org/EIPS/eip-4788
    final ToyAccount deployerOf4788 =
        ToyAccount.builder()
            .nonce(0)
            .address(Address.fromHexString("0x0B799C86a49DEeb90402691F1041aa3AF2d3C875"))
            .balance(Wei.fromEth(109))
            .build();
    final Transaction deploy4788 =
        ToyTransaction.builder()
            .sender(deployerOf4788)
            .transactionType(TransactionType.FRONTIER)
            .gasLimit(0x3d090L)
            .gasPrice(Wei.fromHexString("0xe8d4a51000"))
            .payload(
                Bytes.fromHexString(
                    "0x60618060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604d57602036146024575f5ffd5b5f35801560495762001fff810690815414603c575f5ffd5b62001fff01545f5260205ff35b5f5ffd5b62001fff42064281555f359062001fff015500"))
            .value(Wei.of(0))
            .signature(
                Bytes.concatenate(
                    Bytes32.fromHexStringLenient("0x539"), // r
                    Bytes32.fromHexStringLenient("0x1b9b6eb1f0"), // s
                    Bytes.fromHexString("0x00") // v = 0x1B
                    ))
            .build();

    final List<Transaction> genesisBlockTransactions = new ArrayList<>();
    genesisBlockTransactions.add(check2935Tx());
    genesisBlockTransactions.add(check4788Tx());
    if (valueTransferedPriorToDeploymentOf2935) {
      genesisBlockTransactions.add(transferValueTo2935Tx());
    }
    if (valueTransferedPriorToDeploymentOf4788) {
      genesisBlockTransactions.add(transferValueTo4788Tx());
    }
    if (system2935ContractDeployedBeforeBlockNumber == 1) {
      genesisBlockTransactions.add(deploy2935);
    }
    if (system4788ContractDeployedBeforeBlockNumber == 1) {
      genesisBlockTransactions.add(deploy4788);
    }
    genesisBlockTransactions.add(check2935Tx());
    genesisBlockTransactions.add(check4788Tx());

    final List<Transaction> blockNb1Transactions = new ArrayList<>();
    blockNb1Transactions.add(check2935Tx());
    blockNb1Transactions.add(check4788Tx());
    if (system2935ContractDeployedBeforeBlockNumber == 2) {
      blockNb1Transactions.add(deploy2935);
      blockNb1Transactions.add(check2935Tx());
    }
    if (system4788ContractDeployedBeforeBlockNumber == 2) {
      blockNb1Transactions.add(deploy4788);
      blockNb1Transactions.add(check4788Tx());
    }

    final List<Transaction> blockNb2Transactions = new ArrayList<>();
    blockNb2Transactions.add(check2935Tx());
    blockNb2Transactions.add(check4788Tx());

    final MultiBlockExecutionEnvironment.MultiBlockExecutionEnvironmentBuilder builder =
        MultiBlockExecutionEnvironment.builder(
            chainConfig,
            testInfo,
            system2935ContractDeployedBeforeBlockNumber == 0
                || system4788ContractDeployedBeforeBlockNumber == 0,
            0);
    builder
        .accounts(
            List.of(senderAccount, deployerOf2935, deployerOf4788, callerOf2935, callerOf4788))
        .addBlock(genesisBlockTransactions)
        .addBlock(blockNb1Transactions)
        .addBlock(blockNb2Transactions)
        .build()
        .run();
  }

  private Transaction check2935Tx() {
    return ToyTransaction.builder()
        .sender(senderAccount)
        .to(callerOf2935)
        .nonce(senderNonce++)
        .keyPair(senderKeyPair)
        .build();
  }

  private Transaction transferValueTo2935Tx() {
    return ToyTransaction.builder()
        .sender(senderAccount)
        .toAddress(EIP2935_HISTORY_STORAGE_ADDRESS)
        .nonce(senderNonce++)
        .value(Wei.ONE)
        .keyPair(senderKeyPair)
        .build();
  }

  private Transaction transferValueTo4788Tx() {
    return ToyTransaction.builder()
        .sender(senderAccount)
        .toAddress(EIP4788_BEACONROOT_ADDRESS)
        .nonce(senderNonce++)
        .value(Wei.ONE)
        .keyPair(senderKeyPair)
        .build();
  }

  private Transaction check4788Tx() {
    return ToyTransaction.builder()
        .sender(senderAccount)
        .to(callerOf4788)
        .nonce(senderNonce++)
        .keyPair(senderKeyPair)
        .build();
  }
}

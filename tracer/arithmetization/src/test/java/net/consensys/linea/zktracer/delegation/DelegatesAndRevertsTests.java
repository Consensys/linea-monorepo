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
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
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
   * <p>this tx has 1 delegation with all combinations of the following delegations are valid: yes /
   * no (for valid ones) authority exists: yes / no TX_REQUIRES_EVM_EXECUTION: yes / no If yes: tx
   * reverts: yes / no tx incurs another refund (say a single SSTORE that resets storage or so): yes
   * / no if no: no further requirements
   */
  @ParameterizedTest
  @MethodSource("delegatesAndRevertsTestsSource")
  void delegatesAndRevertsTestsSource(scenario sc, TestInfo testInfo) {

    // sender account
    final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
    final Address senderAddress =
        Address.extract(Hash.hash(senderKeyPair.getPublicKey().getEncodedBytes()));
    final ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.fromEth(56)).nonce(119).address(senderAddress).build();

    // eoa
    final KeyPair eoaKeyPair = new SECP256K1().generateKeyPair();
    final Address eoaAddress =
        Address.extract(Hash.hash(eoaKeyPair.getPublicKey().getEncodedBytes()));
    final ToyAccount eoaAccount =
        ToyAccount.builder().balance(Wei.fromEth(12)).nonce(0).address(eoaAddress).build();

    final ToyTransaction.ToyTransactionBuilder tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(eoaAccount)
            .keyPair(senderKeyPair)
            .gasLimit(300000L)
            .transactionType(TransactionType.DELEGATE_CODE)
            .value(Wei.of(1000));

    if (sc == scenario.DELEGATION_VALID_NON) {
      // delegation is known to fail because of chainID
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
      tx.addCodeDelegation(chainConfig.id, Address.ZERO, 0, eoaKeyPair);
    }

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of())
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
    DELEGATION_VALID_NON,
    DELEGATION_VALID_OUI_AUTHORITY_EXIST_NON,
    DELEGATION_VALID_OUI_AUTHORITY_EXIST_OUI_EVM_EXECUTION_NON,
    DELEGATION_VALID_OUI_AUTHORITY_EXIST_OUI_EVM_EXECUTION_OUI_REVERTS_NON,
    DELEGATION_VALID_OUI_AUTHORITY_EXIST_OUI_EVM_EXECUTION_OUI_REVERTS_OUI_OTHER_REFUND_NON,
    DELEGATION_VALID_OUI_AUTHORITY_EXIST_OUI_EVM_EXECUTION_OUI_REVERTS_OUI_OTHER_REFUND_OUI
  }
}

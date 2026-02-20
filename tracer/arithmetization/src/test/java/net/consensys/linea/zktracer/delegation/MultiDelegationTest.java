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

package net.consensys.linea.zktracer.delegation;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedList;
import net.consensys.linea.zktracer.module.hub.section.TupleAnalysis;
import net.consensys.linea.zktracer.module.rlpAuth.RlpAuthOperation;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.crypto.SECPPrivateKey;
import org.hyperledger.besu.crypto.SECPPublicKey;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

// https://github.com/Consensys/linea-monorepo/issues/2455

@ExtendWith(UnitTestWatcher.class)
public class MultiDelegationTest extends TracerTestBase {

  // TODO: add static ToyAccountBuilders ... and flag in the test to tell if authority delegates already some address
  //  and expected validity

  @ParameterizedTest
  @MethodSource("multiDelegationTestSource")
  void multiDelegationMonoTransactionTest(
      ToyAccount senderAccount,
      ToyAccount authorityAccount,
      ToyAccount delegationAccount,
      ToyAccount recipientAccount,
      CodeDelegation d1,
      CodeDelegation d2,
      CodeDelegation d3,
      TestInfo testInfo) {
    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(recipientAccount)
            .keyPair(senderAccount.getKeyPair())
            .transactionType(TransactionType.DELEGATE_CODE)
            .nonce(senderAccount.getNonce())
            .gasLimit(96000L)
            .addCodeDelegation(d1)
            .addCodeDelegation(d2)
            .addCodeDelegation(d3)
            .build();

    ToyExecutionEnvironmentV2 toyExecutionEnvironmentV2 =
        ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
            .accounts(List.of(senderAccount, authorityAccount, delegationAccount, recipientAccount))
            .transaction(tx)
            .build();
    toyExecutionEnvironmentV2.run();

    ModuleOperationStackedList<RlpAuthOperation> operations =
        toyExecutionEnvironmentV2.getHub().rlpAuth().operations();
    assertEquals(3, operations.size());
    for (RlpAuthOperation operation : operations.getAll()) {
      assertEquals(TupleAnalysis.TUPLE_IS_VALID, operation.authorizationFragment().tupleAnalysis());
    }
  }

  @Test
  void multiDelegationMultiTransactionTest(TestInfo testInfo) {
    // TODO
  }

  public enum DelegationCase {
    DELEGATION_TO_NEW_ADDRESS,
    DELEGATION_TO_CURRENT_DELEGATION,
    DELEGATION_RESET, // Address.ZERO
    DELEGATION_FAILURE_DUE_TO_NONCE_MISMATCH, // authority is recovered and printed in the hub
    DELEGATION_FAILURE_DUE_TO_CHAIN_ID_MISMATCH; // authority is not recovered
  }

  @RequiredArgsConstructor
  @Getter
  public enum AuthorityCase {
    AUTHORITY_IS_RANDOM(42),
    AUTHORITY_IS_SENDER(69),
    AUTHORITY_IS_RECIPIENT(101),
    AUTHORITY_IS_COINBASE(666); // DEFAULT_COINBASE_ADDRESS

    final int baseNonce;
  }

  static Stream<Arguments> multiDelegationTestSource() {
    List<Arguments> arguments = new ArrayList<>();

    return arguments.stream();
  }
}

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

package net.consensys.linea.testing;

import java.math.BigInteger;
import java.util.*;
import java.util.function.Consumer;
import java.util.function.Supplier;
import java.util.stream.Collectors;

import lombok.Builder;
import lombok.Singular;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.hyperledger.besu.datatypes.*;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.mainnet.ProtocolSpec;
import org.hyperledger.besu.ethereum.referencetests.GeneralStateTestCaseEipSpec;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState;

@Builder
@Slf4j
public class ToyExecutionEnvironmentV2 {
  public static final BigInteger CHAIN_ID = BigInteger.valueOf(1337);
  public static final Address DEFAULT_COINBASE_ADDRESS =
      Address.fromHexString("0xc019ba5e00000000c019ba5e00000000c019ba5e");
  public static final long DEFAULT_BLOCK_NUMBER = 6678980;

  private static final long DEFAULT_TIME_STAMP = 1347310;
  private static final Hash DEFAULT_HASH =
      Hash.fromHexStringLenient("0xdeadbeef123123666dead666dead666");

  @Builder.Default private final List<ToyAccount> accounts = Collections.emptyList();
  @Builder.Default private final Address coinbase = DEFAULT_COINBASE_ADDRESS;
  @Builder.Default public static final Wei DEFAULT_BASE_FEE = Wei.of(7);

  @Singular private final List<Transaction> transactions;

  /**
   * A transaction validator of each transaction; by default, it asserts that the transaction was
   * successfully processed.
   */
  @Builder.Default
  private final TransactionProcessingResultValidator transactionProcessingResultValidator =
      TransactionProcessingResultValidator.EMPTY_VALIDATOR;

  // This was previously DEFAULT_VALIDATOR, however some tests we write are supposed to generate
  // failing transactions
  // Thus we cannot use the DEFAULT_VALIDATOR since it asserts that the transaction is successful

  @Builder.Default private final Consumer<ZkTracer> zkTracerValidator = x -> {};

  private final ZkTracer tracer = new ZkTracer(CHAIN_ID);

  public void run() {
    ProtocolSpec protocolSpec = ExecutionEnvironment.getProtocolSpec(CHAIN_ID);
    GeneralStateTestCaseEipSpec generalStateTestCaseEipSpec =
        this.buildGeneralStateTestCaseSpec(protocolSpec);

    GeneralStateReferenceTestTools.executeTest(
        generalStateTestCaseEipSpec,
        protocolSpec,
        tracer,
        transactionProcessingResultValidator,
        zkTracerValidator);
  }

  public Hub getHub() {
    return tracer.getHub();
  }

  public GeneralStateTestCaseEipSpec buildGeneralStateTestCaseSpec(ProtocolSpec protocolSpec) {
    Map<String, ReferenceTestWorldState.AccountMock> accountMockMap =
        accounts.stream()
            .collect(
                Collectors.toMap(
                    toyAccount -> toyAccount.getAddress().toHexString(),
                    ToyAccount::toAccountMock));
    ReferenceTestWorldState referenceTestWorldState =
        ReferenceTestWorldState.create(accountMockMap, protocolSpec.getEvm().getEvmConfiguration());
    BlockHeader blockHeader =
        ExecutionEnvironment.getLineaBlockHeaderBuilder(Optional.empty())
            .number(DEFAULT_BLOCK_NUMBER)
            .coinbase(coinbase)
            .timestamp(DEFAULT_TIME_STAMP)
            .parentHash(DEFAULT_HASH)
            .baseFee(DEFAULT_BASE_FEE)
            .buildBlockHeader();

    List<Supplier<Transaction>> txSuppliers = new ArrayList<>();
    for (Transaction tx : transactions) {
      txSuppliers.add(() -> tx);
    }

    return new GeneralStateTestCaseEipSpec(
        /*fork*/ protocolSpec.getEvm().getEvmVersion().getName().toLowerCase(),
        txSuppliers,
        referenceTestWorldState,
        /*expectedRootHash*/ null,
        /*expectedLogsHash*/ null,
        blockHeader,
        /*dataIndex*/ -1,
        /*gasIndex*/ -1,
        /*valueIndex*/ -1,
        /*expectException*/ null);
  }
}
